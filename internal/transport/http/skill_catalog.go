package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	sc "ai-staffing/internal/service/skillcatalog"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type skillCatalogUseCase interface {
	List(context.Context, sc.ListQuery) (sc.Page, error)
	Create(context.Context, sc.Command) (sc.Item, error)
	Update(context.Context, sc.Command) (sc.Item, error)
	Delete(context.Context, sc.Command) error
}
type skillCategoryResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
type skillCatalogResponse struct {
	ID       int                   `json:"id"`
	Name     string                `json:"name"`
	Category skillCategoryResponse `json:"category"`
}

func skillCatalogDTO(v sc.Item) skillCatalogResponse {
	return skillCatalogResponse{v.ID, v.Name, skillCategoryResponse{v.Category.ID, v.Category.Name}}
}

func registerSkillCatalogRoutes(router chi.Router, sessions sessionAuthenticator, s skillCatalogUseCase, p authProtection) {
	for _, kind := range []sc.Kind{sc.Categories, sc.Skills} {
		router.Route("/api/"+string(kind), func(r chi.Router) {
			r.Use(serverAuthRequestID, recoverAuth, catalogNoStore, authenticateSession(sessions))
			r.Get("/", skillCatalogHandler(s, p, kind, http.MethodGet))
			r.Post("/", skillCatalogHandler(s, p, kind, http.MethodPost))
			r.Patch("/{id}", skillCatalogHandler(s, p, kind, http.MethodPatch))
			r.Delete("/{id}", skillCatalogHandler(s, p, kind, http.MethodDelete))
			r.NotFound(func(w http.ResponseWriter, r *http.Request) { writeSkillCatalogError(w, r, sc.ErrNotFound) })
			r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
				writeAuthError(w, r, 405, "METHOD_NOT_ALLOWED", "Phương thức không được hỗ trợ.")
			})
		})
	}
}
func skillCatalogHandler(s skillCatalogUseCase, p authProtection, kind sc.Kind, method string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromRequest(r)
		if !ok {
			writeSkillCatalogError(w, r, sc.ErrUnauthenticated)
			return
		}
		if method == http.MethodGet {
			q, err := parseSkillCatalogQuery(r.URL.RawQuery, kind)
			if err != nil {
				writeSkillCatalogError(w, r, err)
				return
			}
			q.ActorID = session.principal.UserID
			page, err := s.List(r.Context(), q)
			if err != nil {
				writeSkillCatalogError(w, r, err)
				return
			}
			if kind == sc.Categories {
				response := struct {
					Items    []skillCategoryResponse `json:"items"`
					Page     int                     `json:"page"`
					PageSize int                     `json:"page_size"`
					Total    int64                   `json:"total"`
				}{Items: []skillCategoryResponse{}, Page: page.Page, PageSize: page.PageSize, Total: page.Total}
				for _, v := range page.Items {
					response.Items = append(response.Items, skillCategoryResponse{v.ID, v.Name})
				}
				writeAuthJSON(w, 200, response)
			} else {
				response := struct {
					Items    []skillCatalogResponse `json:"items"`
					Page     int                    `json:"page"`
					PageSize int                    `json:"page_size"`
					Total    int64                  `json:"total"`
				}{Items: []skillCatalogResponse{}, Page: page.Page, PageSize: page.PageSize, Total: page.Total}
				for _, v := range page.Items {
					response.Items = append(response.Items, skillCatalogDTO(v))
				}
				writeAuthJSON(w, 200, response)
			}
			return
		}
		director := false
		for _, role := range session.principal.SystemRoles {
			director = director || role == "DIRECTOR"
		}
		if !director || !p.allowsMutation(r, session.secret) {
			writeSkillCatalogError(w, r, sc.ErrForbidden)
			return
		}
		if r.URL.RawQuery != "" {
			writeSkillCatalogError(w, r, sc.ErrValidation)
			return
		}
		id := 0
		if method != http.MethodPost {
			var err error
			id, err = parseCatalogID(chi.URLParam(r, "id"))
			if err != nil {
				writeSkillCatalogError(w, r, sc.ErrValidation)
				return
			}
		}
		if !hasProfileChangeJSONContentType(r) || len(r.Header.Values("Content-Encoding")) != 0 {
			writeAuthError(w, r, 415, "UNSUPPORTED_MEDIA_TYPE", "Yêu cầu phải dùng JSON UTF-8 không mã hóa.")
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, authBodyLimit))
		var large *http.MaxBytesError
		if errors.As(err, &large) {
			writeAuthError(w, r, 413, "PAYLOAD_TOO_LARGE", "Dữ liệu vượt quá 8 KiB.")
			return
		}
		if err != nil {
			writeSkillCatalogError(w, r, sc.ErrValidation)
			return
		}
		c, err := decodeSkillCatalogBody(body, kind, method)
		if err != nil {
			writeSkillCatalogError(w, r, err)
			return
		}
		c.ActorID = session.principal.UserID
		c.ID = id
		var item sc.Item
		switch method {
		case http.MethodPost:
			item, err = s.Create(r.Context(), c)
		case http.MethodPatch:
			item, err = s.Update(r.Context(), c)
		case http.MethodDelete:
			err = s.Delete(r.Context(), c)
		}
		if err != nil {
			writeSkillCatalogError(w, r, err)
			return
		}
		if method == http.MethodDelete {
			w.WriteHeader(204)
			return
		}
		status := 200
		if method == http.MethodPost {
			status = 201
		}
		if kind == sc.Categories {
			writeAuthJSON(w, status, skillCategoryResponse{item.ID, item.Name})
		} else {
			writeAuthJSON(w, status, skillCatalogDTO(item))
		}
	}
}
func parseSkillCatalogQuery(raw string, kind sc.Kind) (sc.ListQuery, error) {
	q := sc.ListQuery{Kind: kind, Page: 1, PageSize: 20}
	values, err := url.ParseQuery(raw)
	if err != nil {
		return q, sc.ErrValidation
	}
	for key, v := range values {
		if len(v) != 1 {
			return q, sc.ErrValidation
		}
		switch key {
		case "q":
			q.Query = v[0]
		case "page", "page_size", "category_id":
			id, e := parseCatalogID(v[0])
			if e != nil {
				return q, sc.ErrValidation
			}
			switch key {
			case "page":
				q.Page = id
			case "page_size":
				if id > 100 {
					return q, sc.ErrValidation
				}
				q.PageSize = id
			case "category_id":
				if kind != sc.Skills {
					return q, sc.ErrValidation
				}
				q.CategoryID = id
			}
		default:
			return q, sc.ErrValidation
		}
	}
	return q, nil
}
func decodeSkillCatalogBody(body []byte, kind sc.Kind, method string) (sc.Command, error) {
	c := sc.Command{Kind: kind}
	if !validProfileChangeJSON(body) {
		return c, sc.ErrValidation
	}
	names := []string{}
	if method != http.MethodPost {
		names = append(names, "expected_before")
	}
	if method != http.MethodDelete {
		names = append(names, "name")
		if kind == sc.Skills {
			names = append(names, "category_id")
		}
	}
	fields, err := profileChangeObject(body, names...)
	if err != nil || len(fields) != len(names) {
		return c, sc.ErrValidation
	}
	if method != http.MethodDelete {
		if bytes.Equal(fields["name"], []byte("null")) || json.Unmarshal(fields["name"], &c.Name) != nil {
			return c, sc.ErrValidation
		}
		if kind == sc.Skills && (json.Unmarshal(fields["category_id"], &c.CategoryID) != nil || c.CategoryID < 1) {
			return c, sc.ErrValidation
		}
	}
	if method != http.MethodPost {
		beforeNames := []string{"name"}
		if kind == sc.Skills {
			beforeNames = append(beforeNames, "category_id")
		}
		before, e := profileChangeObject(fields["expected_before"], beforeNames...)
		if e != nil || len(before) != len(beforeNames) {
			return c, sc.ErrValidation
		}
		if bytes.Equal(before["name"], []byte("null")) || json.Unmarshal(before["name"], &c.ExpectedBefore.Name) != nil {
			return c, sc.ErrValidation
		}
		if kind == sc.Skills && (json.Unmarshal(before["category_id"], &c.ExpectedBefore.CategoryID) != nil || c.ExpectedBefore.CategoryID < 1) {
			return c, sc.ErrValidation
		}
	}
	return c, nil
}
func writeSkillCatalogError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := 500, "INTERNAL_ERROR", "Không thể xử lý danh mục kỹ năng."
	switch {
	case errors.Is(err, sc.ErrCommitUnknown):
		status, code, message = 503, "COMMIT_OUTCOME_UNKNOWN", "Kết quả lưu chưa được xác định. Hãy tải lại danh sách trước khi thao tác tiếp."
	case errors.Is(err, sc.ErrValidation):
		status, code, message = 400, "VALIDATION_ERROR", "Dữ liệu danh mục kỹ năng không hợp lệ."
	case errors.Is(err, sc.ErrUnauthenticated):
		status, code, message = 401, "UNAUTHENTICATED", "Phiên đăng nhập không còn hợp lệ."
	case errors.Is(err, sc.ErrPasswordChangeRequired):
		status, code, message = 403, "PASSWORD_CHANGE_REQUIRED", "Bạn cần đổi mật khẩu trước khi tiếp tục."
	case errors.Is(err, sc.ErrForbidden):
		status, code, message = 403, "FORBIDDEN", "Bạn không được phép thay đổi danh mục kỹ năng."
	case errors.Is(err, sc.ErrNotFound):
		status, code, message = 404, "NOT_FOUND", "Không tìm thấy mục danh mục."
	case errors.Is(err, sc.ErrStale):
		status, code, message = 409, "STALE_DATA", "Danh mục đã thay đổi. Hãy tải lại trước khi thao tác."
	case errors.Is(err, sc.ErrDuplicateName):
		status, code, message = 409, "DUPLICATE_NAME", "Tên đã tồn tại trong danh mục."
	case errors.Is(err, sc.ErrCatalogInUse):
		status, code, message = 409, "CATALOG_IN_USE", "Danh mục đang được tham chiếu nên không thể xóa."
	case errors.Is(err, sc.ErrUnavailable), errors.Is(err, context.DeadlineExceeded):
		status, code, message = 503, "SERVICE_UNAVAILABLE", "Dịch vụ tạm thời không khả dụng."
	}
	if status >= 500 {
		slog.ErrorContext(r.Context(), "xử lý danh mục kỹ năng thất bại", "code", code, "request_id", middleware.GetReqID(r.Context()))
	}
	writeAuthError(w, r, status, code, message)
}
