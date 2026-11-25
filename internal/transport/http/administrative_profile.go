package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	org "ai-staffing/internal/service/organization"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type administrativeProfileUseCase interface {
	List(context.Context, org.ListAdministrativeEmployeesQuery) (org.AdministrativeEmployeePage, error)
	Read(context.Context, org.AdministrativeProfileQuery) (org.AdministrativeProfile, error)
	Edit(context.Context, org.EditAdministrativeProfileCommand) (org.AdministrativeProfile, error)
}

type administrativeEmployeeResponse struct {
	EmployeeID       int                  `json:"employee_id"`
	EmployeeCode     string               `json:"employee_code"`
	Fullname         string               `json:"fullname"`
	CompanyEmail     string               `json:"company_email"`
	Position         catalogItemResponse  `json:"position"`
	Department       *catalogItemResponse `json:"department"`
	JobLevel         *catalogItemResponse `json:"job_level"`
	EmploymentStatus string               `json:"employment_status"`
}

func administrativeCatalogDTO(v *org.CatalogItem) *catalogItemResponse {
	if v == nil {
		return nil
	}
	return &catalogItemResponse{ID: v.ID, Name: v.Name}
}
func administrativeProfileDTO(v org.AdministrativeProfile) selfProfileResponse {
	return selfProfileResponse{UserID: v.UserID, EmployeeID: v.EmployeeID, EmployeeCode: v.EmployeeCode, Fullname: v.Fullname, CompanyEmail: v.CompanyEmail, Phone: v.Phone, Address: v.Address, Position: catalogItemResponse{ID: v.Position.ID, Name: v.Position.Name}, Department: administrativeCatalogDTO(v.Department), JobLevel: administrativeCatalogDTO(v.JobLevel), EmploymentStatus: v.EmploymentStatus}
}

func administrativeProfileHandler(s administrativeProfileUseCase, p authProtection, operation string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromRequest(r)
		if !ok {
			writeAdministrativeProfileError(w, r, org.ErrUnauthenticated)
			return
		}
		hr := false
		for _, role := range session.principal.SystemRoles {
			hr = hr || role == "HR"
		}
		if !hr {
			writeAdministrativeProfileError(w, r, org.ErrForbidden)
			return
		}
		if operation == "list" {
			q, err := parseCatalogQuery(r.URL.RawQuery)
			if err != nil {
				writeAdministrativeProfileError(w, r, err)
				return
			}
			page, err := s.List(r.Context(), org.ListAdministrativeEmployeesQuery{ActorID: session.principal.UserID, Page: q.Page, PageSize: q.PageSize, Query: q.Query})
			if err != nil {
				writeAdministrativeProfileError(w, r, err)
				return
			}
			response := struct {
				Items    []administrativeEmployeeResponse `json:"items"`
				Page     int64                            `json:"page"`
				PageSize int                              `json:"page_size"`
				Total    int64                            `json:"total"`
			}{Items: make([]administrativeEmployeeResponse, 0, len(page.Items)), Page: page.Page, PageSize: page.PageSize, Total: page.Total}
			for _, v := range page.Items {
				response.Items = append(response.Items, administrativeEmployeeResponse{v.EmployeeID, v.EmployeeCode, v.Fullname, v.CompanyEmail, catalogItemResponse{ID: v.Position.ID, Name: v.Position.Name}, administrativeCatalogDTO(v.Department), administrativeCatalogDTO(v.JobLevel), v.EmploymentStatus})
			}
			writeAuthJSON(w, 200, response)
			return
		}
		if r.URL.RawQuery != "" {
			writeAdministrativeProfileError(w, r, org.ErrValidation)
			return
		}
		id, err := parseCatalogID(chi.URLParam(r, "id"))
		if err != nil {
			writeAdministrativeProfileError(w, r, err)
			return
		}
		var result org.AdministrativeProfile
		if operation == "read" {
			result, err = s.Read(r.Context(), org.AdministrativeProfileQuery{ActorID: session.principal.UserID, EmployeeID: id})
		} else {
			if !p.allowsMutation(r, session.secret) {
				writeAdministrativeProfileError(w, r, org.ErrForbidden)
				return
			}
			if !hasProfileChangeJSONContentType(r) {
				writeAuthError(w, r, 415, "UNSUPPORTED_MEDIA_TYPE", "Yêu cầu phải dùng JSON UTF-8 không nén.")
				return
			}
			body, readErr := io.ReadAll(http.MaxBytesReader(w, r.Body, profileChangeBodyLimit))
			var tooLarge *http.MaxBytesError
			if errors.As(readErr, &tooLarge) {
				writeAuthError(w, r, 413, "PAYLOAD_TOO_LARGE", "Dữ liệu vượt quá 64 KiB.")
				return
			}
			if readErr != nil {
				writeAdministrativeProfileError(w, r, org.ErrValidation)
				return
			}
			input, decodeErr := decodeAdministrativeProfileBody(body)
			if decodeErr != nil {
				writeAdministrativeProfileError(w, r, decodeErr)
				return
			}
			input.ActorID, input.EmployeeID = session.principal.UserID, id
			input, err = org.ValidateEditAdministrativeProfile(input)
			if err != nil {
				writeAdministrativeProfileError(w, r, err)
				return
			}
			result, err = s.Edit(r.Context(), input)
		}
		if err != nil {
			writeAdministrativeProfileError(w, r, err)
			return
		}
		writeAuthJSON(w, 200, administrativeProfileDTO(result))
	}
}

func decodeAdministrativeProfileBody(body []byte) (org.EditAdministrativeProfileCommand, error) {
	var result org.EditAdministrativeProfileCommand
	if !validProfileChangeJSON(body) {
		return result, org.ErrValidation
	}
	fields, err := profileChangeObject(body, "expected_before_values", "changes")
	if err != nil || len(fields) != 2 {
		return result, org.ErrValidation
	}
	result.ExpectedBeforeValues, err = decodeAdministrativeValues(fields["expected_before_values"])
	if err != nil {
		return result, err
	}
	result.Changes, err = decodeAdministrativeValues(fields["changes"])
	if err != nil {
		return result, err
	}
	return result, nil
}
func decodeAdministrativeValues(raw []byte) (org.AdministrativeValues, error) {
	var result org.AdministrativeValues
	fields, err := profileChangeObject(raw, "fullname", "company_email", "phone", "address", "position_id", "department_id", "job_level_id")
	if err != nil || len(fields) == 0 {
		return result, org.ErrValidation
	}
	for key, value := range fields {
		var target any
		switch key {
		case "fullname":
			result.Fullname.Present = true
			target = &result.Fullname.Value
		case "company_email":
			result.CompanyEmail.Present = true
			target = &result.CompanyEmail.Value
		case "phone":
			result.Phone.Present = true
			target = &result.Phone.Value
		case "address":
			result.Address.Present = true
			target = &result.Address.Value
		case "position_id":
			result.PositionID.Present = true
			target = &result.PositionID.Value
		case "department_id":
			result.DepartmentID.Present = true
			target = &result.DepartmentID.Value
		case "job_level_id":
			result.JobLevelID.Present = true
			target = &result.JobLevelID.Value
		}
		if json.Unmarshal(value, target) != nil {
			return result, org.ErrValidation
		}
	}
	return result, nil
}

func writeAdministrativeProfileError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := 500, "INTERNAL_ERROR", "Không thể xử lý hồ sơ hành chính."
	switch {
	case errors.Is(err, org.ErrCommitUnknown):
		status, code, message = 503, "SERVICE_UNAVAILABLE", "Kết quả lưu chưa rõ. Hãy tải lại hồ sơ để đối soát, không tự gửi lại."
	case errors.Is(err, org.ErrValidation):
		status, code, message = 400, "VALIDATION_ERROR", "Dữ liệu hồ sơ hành chính không hợp lệ."
	case errors.Is(err, org.ErrUnauthenticated):
		status, code, message = 401, "UNAUTHENTICATED", "Phiên đăng nhập không còn hợp lệ."
	case errors.Is(err, org.ErrPasswordChangeRequired):
		status, code, message = 403, "PASSWORD_CHANGE_REQUIRED", "Bạn cần đổi mật khẩu trước khi tiếp tục."
	case errors.Is(err, org.ErrForbidden):
		status, code, message = 403, "FORBIDDEN", "Bạn không được phép thực hiện yêu cầu này."
	case errors.Is(err, org.ErrNotFound):
		status, code, message = 404, "NOT_FOUND", "Không tìm thấy hồ sơ."
	case errors.Is(err, org.ErrStale):
		status, code, message = 409, "STALE_DATA", "Dữ liệu nền đã thay đổi. Hãy tải lại trước khi thao tác."
	case errors.Is(err, org.ErrEmailConflict):
		status, code, message = 409, "EMAIL_CONFLICT", "Email đã được sử dụng."
	case errors.Is(err, org.ErrCatalogNotFound):
		status, code, message = 409, "CATALOG_NOT_FOUND", "Danh mục đã thay đổi hoặc không còn tồn tại."
	case errors.Is(err, org.ErrUnavailable), errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		status, code, message = 503, "SERVICE_UNAVAILABLE", "Dịch vụ tạm thời không khả dụng."
	}
	if status >= 500 {
		slog.ErrorContext(r.Context(), "xử lý hồ sơ hành chính thất bại", "code", code, "request_id", middleware.GetReqID(r.Context()))
	}
	writeAuthError(w, r, status, code, message)
}
