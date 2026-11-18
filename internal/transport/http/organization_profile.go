package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	org "ai-staffing/internal/service/organization"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type profileUseCase interface {
	ReadSelf(context.Context, org.SelfProfileQuery) (org.SelfProfile, error)
}
type selfProfileResponse struct {
	UserID           int                  `json:"user_id"`
	EmployeeID       int                  `json:"employee_id"`
	EmployeeCode     string               `json:"employee_code"`
	Fullname         string               `json:"fullname"`
	CompanyEmail     string               `json:"company_email"`
	Phone            *string              `json:"phone"`
	Address          *string              `json:"address"`
	Position         catalogItemResponse  `json:"position"`
	Department       *catalogItemResponse `json:"department"`
	JobLevel         *catalogItemResponse `json:"job_level"`
	EmploymentStatus string               `json:"employment_status"`
}

func registerEmployeeRoutes(router chi.Router, sessions sessionAuthenticator, s profileUseCase, accounts accountUseCase, protection authProtection) {
	router.Route("/api/employees", func(r chi.Router) {
		r.Use(serverAuthRequestID, recoverAuth, catalogNoStore, authenticateSession(sessions))
		if accounts != nil {
			r.Post("/", accountHandler(accounts, protection, "create"))
			r.Get("/next-code", accountHandler(accounts, protection, "preview"))
		}
		if s != nil {
			r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
				session, ok := sessionFromRequest(r)
				if !ok {
					writeProfileError(w, r, org.ErrUnauthenticated)
					return
				}
				if r.URL.RawQuery != "" {
					writeProfileError(w, r, org.ErrValidation)
					return
				}
				id, err := parseCatalogID(chi.URLParam(r, "id"))
				if err != nil {
					writeProfileError(w, r, err)
					return
				}
				p, err := s.ReadSelf(r.Context(), org.SelfProfileQuery{ActorID: session.principal.UserID, EmployeeID: id})
				if err != nil {
					writeProfileError(w, r, err)
					return
				}
				item := func(v *org.CatalogItem) *catalogItemResponse {
					if v == nil {
						return nil
					}
					return &catalogItemResponse{ID: v.ID, Name: v.Name}
				}
				response := selfProfileResponse{
					UserID: p.UserID, EmployeeID: p.EmployeeID,
					EmployeeCode: p.EmployeeCode, Fullname: p.Fullname,
					CompanyEmail: p.CompanyEmail, Phone: p.Phone, Address: p.Address,
					Position:   catalogItemResponse{ID: p.Position.ID, Name: p.Position.Name},
					Department: item(p.Department), JobLevel: item(p.JobLevel),
					EmploymentStatus: p.EmploymentStatus,
				}
				writeAuthJSON(w, 200, response)
			})
		}
		r.NotFound(func(w http.ResponseWriter, r *http.Request) { writeProfileError(w, r, org.ErrNotFound) })
		r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
			writeAuthError(w, r, 405, "METHOD_NOT_ALLOWED", "Phương thức không được hỗ trợ.")
		})
	})
}
func writeProfileError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := 500, "INTERNAL_ERROR", "Không thể đọc hồ sơ."
	switch {
	case errors.Is(err, org.ErrValidation):
		status, code, message = 400, "VALIDATION_ERROR", "Mã hồ sơ không hợp lệ."
	case errors.Is(err, org.ErrUnauthenticated):
		status, code, message = 401, "UNAUTHENTICATED", "Phiên đăng nhập không còn hợp lệ."
	case errors.Is(err, org.ErrPasswordChangeRequired):
		status, code, message = 403, "PASSWORD_CHANGE_REQUIRED", "Bạn cần đổi mật khẩu trước khi tiếp tục."
	case errors.Is(err, org.ErrNotFound):
		status, code, message = 404, "NOT_FOUND", "Không tìm thấy hồ sơ."
	case errors.Is(err, org.ErrUnavailable), errors.Is(err, context.DeadlineExceeded):
		status, code, message = 503, "SERVICE_UNAVAILABLE", "Dịch vụ tạm thời không khả dụng."
	}
	if status >= 500 {
		slog.ErrorContext(r.Context(), "đọc hồ sơ thất bại", "code", code, "request_id", middleware.GetReqID(r.Context()))
	}
	writeAuthError(w, r, status, code, message)
}
