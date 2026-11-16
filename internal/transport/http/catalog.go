package http

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	org "ai-staffing/internal/service/organization"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type catalogUseCase interface {
	List(context.Context, org.ListCatalogQuery) (org.CatalogPage, error)
	Create(context.Context, org.CatalogCommand) (org.CatalogItem, error)
	Rename(context.Context, org.CatalogCommand) (org.CatalogItem, error)
	Delete(context.Context, org.CatalogCommand) error
}

// Bảng route đóng dùng chung cho đăng ký HTTP và lọc contract Swagger.
var catalogCollections = []org.CatalogKind{org.Departments, org.Positions, org.JobLevels}

func registerCatalogRoutes(router chi.Router, sessions sessionAuthenticator, service catalogUseCase, p authProtection) {
	for _, kind := range catalogCollections {
		router.Route("/api/"+string(kind), func(r chi.Router) {
			r.Use(serverAuthRequestID, recoverAuth, catalogNoStore, authenticateSession(sessions))
			r.Get("/", catalogHandler(service, p, kind, http.MethodGet))
			r.Post("/", catalogHandler(service, p, kind, http.MethodPost))
			r.Patch("/{id}", catalogHandler(service, p, kind, http.MethodPatch))
			r.Delete("/{id}", catalogHandler(service, p, kind, http.MethodDelete))
			r.NotFound(func(w http.ResponseWriter, r *http.Request) {
				writeAuthError(w, r, 404, "NOT_FOUND", "Không tìm thấy đường dẫn danh mục.")
			})
			r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
				writeAuthError(w, r, 405, "METHOD_NOT_ALLOWED", "Phương thức không được hỗ trợ.")
			})
		})
	}
}
func catalogNoStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

type catalogItemResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
type catalogPageResponse struct {
	Items      []catalogItemResponse `json:"items"`
	Pagination struct {
		Page     int64 `json:"page"`
		PageSize int   `json:"page_size"`
		Total    int64 `json:"total"`
	} `json:"pagination"`
}

func catalogHandler(s catalogUseCase, p authProtection, kind org.CatalogKind, method string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromRequest(r)
		if !ok {
			writeCatalogError(w, r, org.ErrUnauthenticated)
			return
		}
		allowed := false
		for _, role := range session.principal.SystemRoles {
			if role == "HR" || (method == http.MethodGet && role == "ADMIN") {
				allowed = true
			}
		}
		if !allowed {
			writeCatalogError(w, r, org.ErrForbidden)
			return
		}
		if method == http.MethodGet {
			q, err := parseCatalogQuery(r.URL.RawQuery)
			if err != nil {
				writeCatalogError(w, r, err)
				return
			}
			q.ActorID = session.principal.UserID
			q.Kind = kind
			page, err := s.List(r.Context(), q)
			if err != nil {
				writeCatalogError(w, r, err)
				return
			}
			response := catalogPageResponse{Items: make([]catalogItemResponse, 0, len(page.Items))}
			response.Pagination.Page = page.Page
			response.Pagination.PageSize = page.PageSize
			response.Pagination.Total = page.Total
			for _, item := range page.Items {
				response.Items = append(response.Items, catalogItemResponse{item.ID, item.Name})
			}
			writeAuthJSON(w, 200, response)
			return
		}
		if !p.allowsMutation(r, session.secret) {
			writeCatalogError(w, r, org.ErrForbidden)
			return
		}
		if r.URL.RawQuery != "" {
			writeCatalogError(w, r, org.ErrValidation)
			return
		}
		c := org.CatalogCommand{ActorID: session.principal.UserID, Kind: kind}
		if method != http.MethodPost {
			id, err := parseCatalogID(chi.URLParam(r, "id"))
			if err != nil {
				writeCatalogError(w, r, err)
				return
			}
			c.ID = id
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
			writeCatalogError(w, r, org.ErrValidation)
			return
		}
		fields, valid := decodeCatalogBody(body, method)
		if !valid {
			writeCatalogError(w, r, org.ErrValidation)
			return
		}
		c.Name = fields["name"]
		c.ExpectedBeforeName = fields["expected_before_name"]
		var item org.CatalogItem
		switch method {
		case http.MethodPost:
			item, err = s.Create(r.Context(), c)
		case http.MethodPatch:
			item, err = s.Rename(r.Context(), c)
		case http.MethodDelete:
			err = s.Delete(r.Context(), c)
		}
		if err != nil {
			writeCatalogError(w, r, err)
			return
		}
		if method == http.MethodDelete {
			w.WriteHeader(204)
			return
		}
		status := 200
		if method == http.MethodPost {
			status = 201
			w.Header().Set("Location", "/api/"+string(kind)+"/"+strconv.Itoa(item.ID))
		}
		writeAuthJSON(w, status, catalogItemResponse{item.ID, item.Name})
	}
}
func parseCatalogID(raw string) (int, error) {
	if raw == "" {
		return 0, org.ErrValidation
	}
	for _, c := range raw {
		if c < '0' || c > '9' {
			return 0, org.ErrValidation
		}
	}
	n, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || n < 1 {
		return 0, org.ErrValidation
	}
	return int(n), nil
}
func parseCatalogQuery(raw string) (org.ListCatalogQuery, error) {
	q := org.ListCatalogQuery{Page: 1, PageSize: 20}
	values, err := url.ParseQuery(raw)
	if err != nil {
		return q, org.ErrValidation
	}
	for k, v := range values {
		if len(v) != 1 {
			return q, org.ErrValidation
		}
		switch k {
		case "page":
			n, e := parseCatalogID(v[0])
			if e != nil {
				return q, e
			}
			q.Page = int64(n)
		case "page_size":
			n, e := parseCatalogID(v[0])
			if e != nil || n > 100 {
				return q, org.ErrValidation
			}
			q.PageSize = n
		case "q":
			q.Query = v[0]
		default:
			return q, org.ErrValidation
		}
	}
	return q, nil
}
func decodeCatalogBody(body []byte, method string) (map[string]string, bool) {
	switch method {
	case http.MethodPatch:
		return decodeAuthStringPair(body, "name", "expected_before_name")
	case http.MethodPost:
		return decodeStringFields(body, "name")
	case http.MethodDelete:
		return decodeStringFields(body, "expected_before_name")
	}
	return nil, false
}
func writeCatalogError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := 500, "INTERNAL_ERROR", "Không thể xử lý yêu cầu danh mục."
	switch {
	case errors.Is(err, org.ErrCommitUnknown):
		if errors.Is(err, org.ErrUnavailable) || errors.Is(err, context.DeadlineExceeded) {
			status, code, message = 503, "SERVICE_UNAVAILABLE", "Kết quả lưu chưa được xác định. Hãy tải lại danh sách trước khi thao tác tiếp."
		}
	case errors.Is(err, org.ErrValidation):
		status, code, message = 400, "VALIDATION_ERROR", "Dữ liệu danh mục không hợp lệ."
	case errors.Is(err, org.ErrUnauthenticated):
		status, code, message = 401, "UNAUTHENTICATED", "Phiên đăng nhập không còn hợp lệ."
	case errors.Is(err, org.ErrPasswordChangeRequired):
		status, code, message = 403, "PASSWORD_CHANGE_REQUIRED", "Bạn cần đổi mật khẩu trước khi tiếp tục."
	case errors.Is(err, org.ErrForbidden):
		status, code, message = 403, "FORBIDDEN", "Bạn không được phép thực hiện yêu cầu này."
	case errors.Is(err, org.ErrNotFound):
		status, code, message = 404, "NOT_FOUND", "Không tìm thấy mục danh mục."
	case errors.Is(err, org.ErrStale):
		status, code, message = 409, "STALE_DATA", "Danh mục đã thay đổi. Hãy tải lại trước khi thao tác."
	case errors.Is(err, org.ErrDuplicateName):
		status, code, message = 409, "DUPLICATE_NAME", "Tên đã tồn tại trong danh mục."
	case errors.Is(err, org.ErrCatalogInUse):
		status, code, message = 409, "CATALOG_IN_USE", "Danh mục đang được sử dụng nên không thể xóa."
	case errors.Is(err, org.ErrUnavailable), errors.Is(err, context.DeadlineExceeded):
		status, code, message = 503, "SERVICE_UNAVAILABLE", "Dịch vụ tạm thời không khả dụng."
	}
	if status >= 500 {
		slog.ErrorContext(r.Context(), "xử lý danh mục thất bại", "code", code, "request_id", middleware.GetReqID(r.Context()))
	}
	writeAuthError(w, r, status, code, message)
}
