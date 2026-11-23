package http

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	pc "ai-staffing/internal/service/profilechange"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type profileChangeUseCase interface {
	Submit(context.Context, pc.SubmitInput) (pc.Detail, error)
	List(context.Context, pc.ListQuery) (pc.Page, error)
	Read(context.Context, pc.DetailQuery) (pc.Detail, error)
	Approve(context.Context, pc.DetailQuery) (pc.Detail, error)
	Reject(context.Context, pc.RejectInput) (pc.Detail, error)
}

func registerProfileChangeRoutes(router chi.Router, sessions sessionAuthenticator, service profileChangeUseCase, p authProtection) {
	router.Route("/api/profile-change-requests", func(r chi.Router) {
		r.Use(serverAuthRequestID, recoverAuth, catalogNoStore, authenticateSession(sessions))
		r.Get("/", profileChangeHandler(service, p, "list"))
		r.Get("/{id}", profileChangeHandler(service, p, "read"))
		r.Post("/{id}/approve", profileChangeHandler(service, p, "approve"))
		r.Post("/{id}/reject", profileChangeHandler(service, p, "reject"))
		r.NotFound(func(w http.ResponseWriter, r *http.Request) { writeProfileChangeError(w, r, pc.ErrNotFound) })
		r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
			writeAuthError(w, r, 405, "METHOD_NOT_ALLOWED", "Phương thức không được hỗ trợ.")
		})
	})
}

type profileChangeSummaryResponse struct {
	RequestID     int        `json:"request_id"`
	EmployeeID    int        `json:"employee_id"`
	EmployeeCode  string     `json:"employee_code"`
	Fullname      string     `json:"fullname"`
	Status        string     `json:"status"`
	ChangedFields []string   `json:"changed_fields"`
	RequestedAt   time.Time  `json:"requested_at"`
	ResolvedAt    *time.Time `json:"resolved_at"`
}

type profileChangeDetailResponse struct {
	RequestID      int        `json:"request_id"`
	EmployeeID     int        `json:"employee_id"`
	EmployeeCode   string     `json:"employee_code"`
	Fullname       string     `json:"fullname"`
	Status         string     `json:"status"`
	Before         pc.Values  `json:"before_values"`
	Proposed       pc.Values  `json:"proposed_values"`
	Current        pc.Values  `json:"current_values"`
	Explanation    *string    `json:"explanation"`
	RequestedBy    int        `json:"requested_by"`
	RequestedAt    time.Time  `json:"requested_at"`
	ResolvedBy     *int       `json:"resolved_by"`
	ResolvedAt     *time.Time `json:"resolved_at"`
	DecisionReason *string    `json:"decision_reason"`
}

func profileChangeDetailDTO(v pc.Detail) profileChangeDetailResponse {
	return profileChangeDetailResponse{
		RequestID: v.RequestID, EmployeeID: v.EmployeeID, EmployeeCode: v.EmployeeCode,
		Fullname: v.Fullname, Status: v.Status, Before: v.Before, Proposed: v.Proposed, Current: v.Current,
		Explanation: v.Explanation, RequestedBy: v.RequestedBy, RequestedAt: v.RequestedAt,
		ResolvedBy: v.ResolvedBy, ResolvedAt: v.ResolvedAt, DecisionReason: v.DecisionReason,
	}
}

func profileChangeHandler(s profileChangeUseCase, p authProtection, operation string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromRequest(r)
		if !ok {
			writeProfileChangeError(w, r, pc.ErrUnauthenticated)
			return
		}
		hr := false
		for _, role := range session.principal.SystemRoles {
			hr = hr || role == "HR"
		}
		if (operation == "approve" || operation == "reject") && !hr {
			writeProfileChangeError(w, r, pc.ErrForbidden)
			return
		}
		if operation == "list" {
			q, err := parseProfileChangeQuery(r.URL.RawQuery)
			if err != nil {
				writeProfileChangeError(w, r, err)
				return
			}
			if q.Scope == "review" && !hr {
				writeProfileChangeError(w, r, pc.ErrForbidden)
				return
			}
			q.ActorID = session.principal.UserID
			page, err := s.List(r.Context(), q)
			if err != nil {
				writeProfileChangeError(w, r, err)
				return
			}
			response := struct {
				Items    []profileChangeSummaryResponse `json:"items"`
				Page     int64                          `json:"page"`
				PageSize int                            `json:"page_size"`
				Total    int64                          `json:"total"`
			}{Items: make([]profileChangeSummaryResponse, 0, len(page.Items)), Page: page.Page, PageSize: page.PageSize, Total: page.Total}
			for _, v := range page.Items {
				response.Items = append(response.Items, profileChangeSummaryResponse{v.RequestID, v.EmployeeID, v.EmployeeCode, v.Fullname, v.Status, v.ChangedFields, v.RequestedAt, v.ResolvedAt})
			}
			writeAuthJSON(w, 200, response)
			return
		}
		if r.URL.RawQuery != "" {
			writeProfileChangeError(w, r, pc.ErrValidation)
			return
		}
		id, err := parseCatalogID(chi.URLParam(r, "id"))
		if err != nil {
			writeProfileChangeError(w, r, pc.ErrValidation)
			return
		}
		query := pc.DetailQuery{ActorID: session.principal.UserID, RequestID: id}
		var detail pc.Detail
		status := 200
		if operation == "read" {
			detail, err = s.Read(r.Context(), query)
		} else {
			if !p.allowsMutation(r, session.secret) {
				writeProfileChangeError(w, r, pc.ErrForbidden)
				return
			}
			if !hasProfileChangeJSONContentType(r) {
				writeAuthError(w, r, 415, "UNSUPPORTED_MEDIA_TYPE", "Yêu cầu phải dùng JSON UTF-8 không nén.")
				return
			}
			key := ""
			if operation == "submit" {
				keys := r.Header.Values("Idempotency-Key")
				if len(keys) != 1 || !profileChangeKeyPattern.MatchString(keys[0]) {
					writeProfileChangeError(w, r, pc.ErrValidation)
					return
				}
				key = strings.ToLower(keys[0])
			}
			body, readErr := io.ReadAll(http.MaxBytesReader(w, r.Body, profileChangeBodyLimit))
			var tooLarge *http.MaxBytesError
			if errors.As(readErr, &tooLarge) {
				writeAuthError(w, r, 413, "PAYLOAD_TOO_LARGE", "Dữ liệu vượt quá 64 KiB.")
				return
			}
			if readErr != nil {
				writeProfileChangeError(w, r, pc.ErrValidation)
				return
			}
			input, reason, decodeErr := decodeProfileChangeBody(body, operation)
			if decodeErr != nil {
				writeProfileChangeError(w, r, decodeErr)
				return
			}
			switch operation {
			case "submit":
				input.ActorID, input.EmployeeID, input.Key = session.principal.UserID, id, key
				detail, err = s.Submit(r.Context(), input)
				status = 201
			case "approve":
				detail, err = s.Approve(r.Context(), query)
			case "reject":
				detail, err = s.Reject(r.Context(), pc.RejectInput{ActorID: query.ActorID, RequestID: id, Reason: reason})
			default:
				err = pc.ErrValidation
			}
		}
		if err != nil {
			writeProfileChangeError(w, r, err)
			return
		}
		writeAuthJSON(w, status, profileChangeDetailDTO(detail))
	}
}

func writeProfileChangeError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := 500, "INTERNAL_ERROR", "Không thể xử lý đề nghị hồ sơ."
	switch {
	case errors.Is(err, pc.ErrCommitUnknown):
		status, code, message = 503, "SERVICE_UNAVAILABLE", "Kết quả lưu chưa rõ. Hãy tải lại danh sách để đối soát, không tự gửi lại."
	case errors.Is(err, pc.ErrValidation):
		status, code, message = 400, "VALIDATION_ERROR", "Dữ liệu đề nghị hồ sơ không hợp lệ."
	case errors.Is(err, pc.ErrUnauthenticated):
		status, code, message = 401, "UNAUTHENTICATED", "Phiên đăng nhập không còn hợp lệ."
	case errors.Is(err, pc.ErrPasswordChangeRequired):
		status, code, message = 403, "PASSWORD_CHANGE_REQUIRED", "Bạn cần đổi mật khẩu trước khi tiếp tục."
	case errors.Is(err, pc.ErrForbidden):
		status, code, message = 403, "FORBIDDEN", "Bạn không được phép thực hiện yêu cầu này."
	case errors.Is(err, pc.ErrNotFound):
		status, code, message = 404, "NOT_FOUND", "Không tìm thấy đề nghị hồ sơ."
	case errors.Is(err, pc.ErrStale):
		status, code, message = 409, "STALE_DATA", "Dữ liệu nền đã thay đổi. Hãy tải lại trước khi thao tác."
	case errors.Is(err, pc.ErrAlreadyResolved):
		status, code, message = 409, "REQUEST_ALREADY_RESOLVED", "Đề nghị đã được xử lý."
	case errors.Is(err, pc.ErrEmployeeInactive):
		status, code, message = 409, "EMPLOYEE_INACTIVE", "Nhân viên không còn hoạt động nên không thể chấp thuận."
	case errors.Is(err, pc.ErrKeyReused):
		status, code, message = 409, "IDEMPOTENCY_KEY_REUSED", "Mã gửi đề nghị đã được dùng với nội dung khác."
	case errors.Is(err, pc.ErrKeyExpired):
		status, code, message = 409, "IDEMPOTENCY_KEY_EXPIRED", "Mã gửi đề nghị đã hết thời gian đối soát. Hãy xem lại lịch sử."
	case errors.Is(err, pc.ErrUnavailable), errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		status, code, message = 503, "SERVICE_UNAVAILABLE", "Dịch vụ tạm thời không khả dụng."
	}
	if status >= 500 {
		slog.ErrorContext(r.Context(), "xử lý đề nghị hồ sơ thất bại", "code", code, "request_id", middleware.GetReqID(r.Context()))
	}
	writeAuthError(w, r, status, code, message)
}
