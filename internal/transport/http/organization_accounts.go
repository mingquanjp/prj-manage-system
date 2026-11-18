package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"unicode/utf8"

	org "ai-staffing/internal/service/organization"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type accountUseCase interface {
	Create(context.Context, org.CreateAccountInput) (org.AccountCreation, error)
	List(context.Context, org.ListAccountsQuery) (org.AccountPage, error)
	Preview(context.Context, int) (string, error)
}

type accountResponse struct {
	UserID             int      `json:"user_id"`
	EmployeeID         int      `json:"employee_id"`
	EmployeeCode       string   `json:"employee_code"`
	Fullname           string   `json:"fullname"`
	CompanyEmail       string   `json:"company_email"`
	Status             string   `json:"status"`
	SystemRoles        []string `json:"system_roles"`
	MustChangePassword bool     `json:"must_change_password"`
}

func accountDTO(a org.AccountProjection) accountResponse {
	roles := append([]string{}, a.SystemRoles...)
	return accountResponse{a.UserID, a.EmployeeID, a.EmployeeCode, a.Fullname, a.CompanyEmail, a.Status, roles, a.MustChangePassword}
}

func registerAccountListRoute(router chi.Router, sessions sessionAuthenticator, service accountUseCase, p authProtection) {
	router.Route("/api/users", func(r chi.Router) {
		r.Use(serverAuthRequestID, recoverAuth, catalogNoStore, authenticateSession(sessions))
		r.Get("/", accountHandler(service, p, "list"))
		r.NotFound(func(w http.ResponseWriter, r *http.Request) { writeAccountError(w, r, org.ErrNotFound) })
		r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
			writeAuthError(w, r, 405, "METHOD_NOT_ALLOWED", "Phương thức không được hỗ trợ.")
		})
	})
}

func accountHandler(service accountUseCase, protection authProtection, operation string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromRequest(r)
		if !ok {
			writeAccountError(w, r, org.ErrUnauthenticated)
			return
		}
		admin := false
		for _, role := range session.principal.SystemRoles {
			if role == "ADMIN" {
				admin = true
			}
		}
		if !admin {
			writeAccountError(w, r, org.ErrForbidden)
			return
		}
		switch operation {
		case "list":
			q, err := parseCatalogQuery(r.URL.RawQuery)
			if err != nil {
				writeAccountError(w, r, err)
				return
			}
			page, err := service.List(r.Context(), org.ListAccountsQuery{ActorID: session.principal.UserID, Page: q.Page, PageSize: q.PageSize, Query: q.Query})
			if err != nil {
				writeAccountError(w, r, err)
				return
			}
			response := struct {
				Items      []accountResponse `json:"items"`
				Pagination struct {
					Page     int64 `json:"page"`
					PageSize int   `json:"page_size"`
					Total    int64 `json:"total"`
				} `json:"pagination"`
			}{Items: make([]accountResponse, 0, len(page.Items))}
			response.Pagination.Page, response.Pagination.PageSize, response.Pagination.Total = page.Page, page.PageSize, page.Total
			for _, item := range page.Items {
				response.Items = append(response.Items, accountDTO(item))
			}
			writeAuthJSON(w, 200, response)
		case "preview":
			if r.URL.RawQuery != "" {
				writeAccountError(w, r, org.ErrValidation)
				return
			}
			code, err := service.Preview(r.Context(), session.principal.UserID)
			if err != nil {
				writeAccountError(w, r, err)
				return
			}
			writeAuthJSON(w, 200, struct {
				Code    string `json:"employee_code"`
				Preview bool   `json:"is_preview"`
			}{code, true})
		case "create":
			if !protection.allowsMutation(r, session.secret) {
				writeAccountError(w, r, org.ErrForbidden)
				return
			}
			if r.URL.RawQuery != "" {
				writeAccountError(w, r, org.ErrValidation)
				return
			}
			if !hasAuthJSONContentType(r) {
				writeAuthError(w, r, 415, "UNSUPPORTED_MEDIA_TYPE", "Yêu cầu phải dùng application/json.")
				return
			}
			body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, authBodyLimit))
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				writeAuthError(w, r, 413, "PAYLOAD_TOO_LARGE", "Dữ liệu vượt quá 8 KiB.")
				return
			}
			if err != nil {
				writeAccountError(w, r, org.ErrValidation)
				return
			}
			input, err := decodeAccountBody(body, session.principal.UserID)
			if err != nil {
				writeAccountError(w, r, err)
				return
			}
			result, err := service.Create(r.Context(), input)
			if err != nil {
				writeAccountError(w, r, err)
				return
			}
			writeAuthJSON(w, 201, struct {
				Created bool            `json:"account_created"`
				Account accountResponse `json:"account"`
				Mail    struct {
					Status org.MailStatus `json:"status"`
				} `json:"mail"`
			}{true, accountDTO(result.Account), struct {
				Status org.MailStatus `json:"status"`
			}{result.Mail}})
		}
	}
}

// Parser chỉ nhận một object, kiểm key trùng trước khi đọc kiểu và allowlist field.
func decodeAccountBody(body []byte, actorID int) (org.CreateAccountInput, error) {
	var input org.CreateAccountInput
	if !utf8.Valid(body) {
		return input, org.ErrValidation
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return input, org.ErrValidation
	}
	fields := make(map[string]json.RawMessage)
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok {
			return input, org.ErrValidation
		}
		if _, exists := fields[key]; exists {
			return input, org.ErrValidation
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil {
			return input, org.ErrValidation
		}
		fields[key] = value
	}
	if token, err = decoder.Token(); err != nil || token != json.Delim('}') {
		return input, org.ErrValidation
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return input, org.ErrValidation
	}
	for _, key := range []string{"fullname", "company_email", "position_id"} {
		raw, exists := fields[key]
		if !exists || bytes.Equal(raw, []byte("null")) {
			return input, org.ErrValidation
		}
	}
	for key, raw := range fields {
		var target any
		switch key {
		case "fullname":
			target = &input.Fullname
		case "company_email":
			target = &input.CompanyEmail
		case "position_id":
			target = &input.PositionID
		case "department_id":
			target = &input.DepartmentID
		case "job_level_id":
			target = &input.JobLevelID
		case "phone":
			target = &input.Phone
		case "address":
			target = &input.Address
		case "system_roles":
			if bytes.Equal(raw, []byte("null")) {
				return input, org.ErrValidation
			}
			target = &input.SystemRoles
		default:
			return input, org.ErrValidation
		}
		if len(raw) > 0 && raw[0] == '"' && !validJSONSurrogates(raw) {
			return input, org.ErrValidation
		}
		if json.Unmarshal(raw, target) != nil {
			return input, org.ErrValidation
		}
	}
	input.ActorID = actorID
	input, err = org.ValidateAccountInput(input)
	return input, err
}

func writeAccountError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := 500, "INTERNAL_ERROR", "Không thể xử lý yêu cầu tài khoản."
	switch {
	case errors.Is(err, org.ErrCommitUnknown):
		message = "Kết quả tạo chưa được xác định. Hãy kiểm tra danh sách tài khoản trước khi thao tác tiếp."
		if errors.Is(err, org.ErrUnavailable) || errors.Is(err, context.DeadlineExceeded) {
			status, code = 503, "SERVICE_UNAVAILABLE"
		}
	case errors.Is(err, org.ErrValidation):
		status, code, message = 400, "VALIDATION_ERROR", "Dữ liệu tài khoản không hợp lệ."
	case errors.Is(err, org.ErrUnauthenticated):
		status, code, message = 401, "UNAUTHENTICATED", "Phiên đăng nhập không còn hợp lệ."
	case errors.Is(err, org.ErrPasswordChangeRequired):
		status, code, message = 403, "PASSWORD_CHANGE_REQUIRED", "Bạn cần đổi mật khẩu trước khi tiếp tục."
	case errors.Is(err, org.ErrForbidden):
		status, code, message = 403, "FORBIDDEN", "Bạn không được phép thực hiện yêu cầu này."
	case errors.Is(err, org.ErrEmailConflict):
		status, code, message = 409, "EMAIL_CONFLICT", "Email đã được dùng cho một tài khoản."
	case errors.Is(err, org.ErrNotFound):
		status, code, message = 404, "NOT_FOUND", "Không tìm thấy đường dẫn tài khoản."
	case errors.Is(err, org.ErrUnavailable), errors.Is(err, context.DeadlineExceeded):
		status, code, message = 503, "SERVICE_UNAVAILABLE", "Dịch vụ tạm thời không khả dụng."
	}
	if status >= 500 {
		slog.ErrorContext(r.Context(), "xử lý tài khoản thất bại", "code", code, "request_id", middleware.GetReqID(r.Context()))
	}
	writeAuthError(w, r, status, code, message)
}
