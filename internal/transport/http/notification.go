package http

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	n "ai-staffing/internal/service/notification"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type notificationUseCase interface {
	List(context.Context, n.ListQuery) (n.Page, error)
	MarkRead(context.Context, n.ReadCommand) error
}
type notificationResponse struct {
	ID           int        `json:"notification_id"`
	Type         string     `json:"notification_type"`
	ResourceType string     `json:"resource_type"`
	ResourceID   int        `json:"resource_id"`
	Message      string     `json:"message"`
	CreatedAt    time.Time  `json:"created_at"`
	ReadAt       *time.Time `json:"read_at"`
}
type notificationPageResponse struct {
	Items       []notificationResponse `json:"items"`
	Page        int                    `json:"page"`
	PageSize    int                    `json:"page_size"`
	Total       int64                  `json:"total"`
	UnreadCount int64                  `json:"unread_count"`
}

func registerNotificationRoutes(router chi.Router, sessions sessionAuthenticator, s notificationUseCase, p authProtection) {
	router.Route("/api/notifications", func(r chi.Router) {
		r.Use(serverAuthRequestID, recoverAuth, catalogNoStore, authenticateSession(sessions))
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			session, ok := sessionFromRequest(r)
			if !ok {
				writeNotificationError(w, r, n.ErrUnauthenticated)
				return
			}
			q, err := parseNotificationQuery(r.URL.RawQuery)
			if err != nil {
				writeNotificationError(w, r, err)
				return
			}
			q.ActorID = session.principal.UserID
			page, err := s.List(r.Context(), q)
			if err != nil {
				writeNotificationError(w, r, err)
				return
			}
			response := notificationPageResponse{Items: []notificationResponse{}, Page: page.Page, PageSize: page.PageSize, Total: page.Total, UnreadCount: page.UnreadCount}
			for _, v := range page.Items {
				response.Items = append(response.Items, notificationResponse{v.ID, v.Type, v.ResourceType, v.ResourceID, v.Message, v.CreatedAt, v.ReadAt})
			}
			writeAuthJSON(w, 200, response)
		})
		r.Post("/{id}/read", func(w http.ResponseWriter, r *http.Request) {
			session, ok := sessionFromRequest(r)
			if !ok {
				writeNotificationError(w, r, n.ErrUnauthenticated)
				return
			}
			if !p.allowsMutation(r, session.secret) {
				writeAuthError(w, r, 403, "FORBIDDEN", "Nguồn hoặc CSRF không hợp lệ.")
				return
			}
			id, err := parseCatalogID(chi.URLParam(r, "id"))
			if err != nil || r.URL.RawQuery != "" {
				writeNotificationError(w, r, n.ErrValidation)
				return
			}
			ct := r.Header.Values("Content-Type")
			media, params, e := mime.ParseMediaType(r.Header.Get("Content-Type"))
			charset, hasCharset := params["charset"]
			if len(ct) != 1 || e != nil || media != "application/json" || (hasCharset && !strings.EqualFold(charset, "utf-8")) || len(r.Header.Values("Content-Encoding")) != 0 {
				writeAuthError(w, r, 415, "UNSUPPORTED_MEDIA_TYPE", "Yêu cầu phải dùng JSON UTF-8 không mã hóa.")
				return
			}
			body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 65536))
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				writeAuthError(w, r, 413, "PAYLOAD_TOO_LARGE", "Dữ liệu vượt quá 64 KiB.")
				return
			}
			_, valid := decodeStringFields(body)
			if err != nil || !valid {
				writeNotificationError(w, r, n.ErrValidation)
				return
			}
			if err = s.MarkRead(r.Context(), n.ReadCommand{ActorID: session.principal.UserID, NotificationID: id}); err != nil {
				writeNotificationError(w, r, err)
				return
			}
			w.WriteHeader(204)
		})
		r.NotFound(func(w http.ResponseWriter, r *http.Request) { writeNotificationError(w, r, n.ErrNotFound) })
		r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
			writeAuthError(w, r, 405, "METHOD_NOT_ALLOWED", "Phương thức không được hỗ trợ.")
		})
	})
}
func parseNotificationQuery(raw string) (n.ListQuery, error) {
	q := n.ListQuery{Filter: "all", Page: 1, PageSize: 20}
	values, err := url.ParseQuery(raw)
	if err != nil {
		return q, n.ErrValidation
	}
	for k, v := range values {
		if len(v) != 1 {
			return q, n.ErrValidation
		}
		switch k {
		case "filter":
			if v[0] != "all" && v[0] != "unread" {
				return q, n.ErrValidation
			}
			q.Filter = v[0]
		case "page", "page_size":
			id, err := parseCatalogID(v[0])
			if err != nil || (k == "page_size" && id > 100) {
				return q, n.ErrValidation
			}
			if k == "page" {
				q.Page = id
			} else {
				q.PageSize = id
			}
		default:
			return q, n.ErrValidation
		}
	}
	return q, nil
}
func writeNotificationError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := 500, "INTERNAL_ERROR", "Không thể xử lý thông báo."
	switch {
	case errors.Is(err, n.ErrCommitUnknown):
		message = "Kết quả đánh dấu đọc chưa được xác định. Hãy tải lại danh sách."
		if errors.Is(err, n.ErrUnavailable) || errors.Is(err, context.DeadlineExceeded) {
			status, code = 503, "SERVICE_UNAVAILABLE"
		}
	case errors.Is(err, n.ErrValidation):
		status, code, message = 400, "VALIDATION_ERROR", "Dữ liệu thông báo không hợp lệ."
	case errors.Is(err, n.ErrUnauthenticated):
		status, code, message = 401, "UNAUTHENTICATED", "Phiên đăng nhập không còn hợp lệ."
	case errors.Is(err, n.ErrPasswordChangeRequired):
		status, code, message = 403, "PASSWORD_CHANGE_REQUIRED", "Bạn cần đổi mật khẩu trước khi tiếp tục."
	case errors.Is(err, n.ErrNotFound):
		status, code, message = 404, "NOT_FOUND", "Không tìm thấy thông báo."
	case errors.Is(err, n.ErrUnavailable), errors.Is(err, context.DeadlineExceeded):
		status, code, message = 503, "SERVICE_UNAVAILABLE", "Dịch vụ tạm thời không khả dụng."
	}
	if status >= 500 {
		slog.ErrorContext(r.Context(), "xử lý thông báo thất bại", "code", code, "request_id", middleware.GetReqID(r.Context()))
	}
	writeAuthError(w, r, status, code, message)
}
