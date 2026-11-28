//go:build integration

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"ai-staffing/internal/testutil/authdb"
)

type skillCatalogDTO struct {
	ID       int         `json:"id"`
	Name     string      `json:"name"`
	Category *catalogDTO `json:"category"`
}

func TestSkillCatalogApplicationPostgresCRUDAndRoles(t *testing.T) {
	a := newAuthApp(t)
	f := a.fixture
	director := f.AddAccount(t, authdb.AccountOptions{Email: "skill-director@example.invalid", Roles: []string{"DIRECTOR", "HR", "ADMIN"}})
	cookie := a.login(t, "skill-director@example.invalid", authdb.FixturePassword, nil)
	token := a.csrf(t, cookie)
	create := func(path string, body any) skillCatalogDTO {
		t.Helper()
		_, raw := catalogCall(t, a, "POST", path, cookie, token, body, 201)
		var item skillCatalogDTO
		authdb.Require(t, "đọc mục kỹ năng", json.Unmarshal(raw, &item))
		var fields map[string]json.RawMessage
		authdb.Require(t, "đọc allowlist mục", json.Unmarshal(raw, &fields))
		wantFields := 2
		if path == "/api/skills" {
			wantFields = 3
		}
		if item.ID < 1 || len(fields) != wantFields {
			t.Fatal("DTO danh mục vượt allowlist hoặc thiếu ID")
		}
		return item
	}
	first := create("/api/skill-categories", map[string]any{"name": "\u00a0Nhóm %_\u3000"})
	second := create("/api/skill-categories", map[string]any{"name": "Nhóm khác"})
	if first.Name != "Nhóm %_" {
		t.Fatal("tên nhóm chưa trim Unicode")
	}
	skill := create("/api/skills", map[string]any{"name": "  Kỹ năng %_  ", "category_id": first.ID})
	if skill.Name != "Kỹ năng %_" || skill.Category == nil || skill.Category.ID != first.ID || skill.Category.Name != first.Name {
		t.Fatal("projection kỹ năng/nhóm sai")
	}
	for _, path := range []string{"/api/skill-categories?q=%25_", "/api/skills?q=%25_&category_id=" + strconv.Itoa(first.ID)} {
		_, raw := catalogCall(t, a, "GET", path, cookie, "", nil, 200)
		var page struct {
			Items    []skillCatalogDTO `json:"items"`
			Page     int               `json:"page"`
			PageSize int               `json:"page_size"`
			Total    int64             `json:"total"`
		}
		authdb.Require(t, "page danh mục", json.Unmarshal(raw, &page))
		if len(page.Items) != 1 || page.Total != 1 || page.Page != 1 || page.PageSize != 20 {
			t.Fatal("literal search/filter/page sai")
		}
	}
	for _, roles := range [][]string{nil, {"HR"}, {"ADMIN"}} {
		email := "skill-reader-" + strconv.Itoa(len(roles))
		if len(roles) > 0 {
			email += roles[0]
		}
		email += "@example.invalid"
		f.AddAccount(t, authdb.AccountOptions{Email: email, Roles: roles})
		reader := a.login(t, email, authdb.FixturePassword, nil)
		csrf := a.csrf(t, reader)
		for _, path := range []string{"/api/skill-categories", "/api/skills"} {
			catalogCall(t, a, "GET", path, reader, "", nil, 200)
			body := map[string]any{"name": "Không được ghi"}
			if path == "/api/skills" {
				body["category_id"] = first.ID
			}
			_, raw := catalogCall(t, a, "POST", path, reader, csrf, body, 403)
			assertErrorCode(t, raw, "FORBIDDEN")
		}
	}
	f.AddAccount(t, authdb.AccountOptions{Email: "skill-limited@example.invalid", Roles: []string{"DIRECTOR"}, MustChangePassword: true})
	limited := a.login(t, "skill-limited@example.invalid", authdb.FixturePassword, nil)
	for _, path := range []string{"/api/skill-categories", "/api/skills"} {
		_, raw := catalogCall(t, a, "GET", path, limited, "", nil, 403)
		assertErrorCode(t, raw, "PASSWORD_CHANGE_REQUIRED")
		catalogCall(t, a, "GET", path, nil, "", nil, 401)
	}
	skillPath := "/api/skills/" + strconv.Itoa(skill.ID)
	before := map[string]any{"name": skill.Name, "category_id": first.ID}
	var audits int
	authdb.Require(t, "audit trước no-op", f.DB.Conn.QueryRow(t.Context(), `SELECT count(*) FROM audit_events`).Scan(&audits))
	catalogCall(t, a, "PATCH", skillPath, cookie, token, map[string]any{"name": skill.Name, "category_id": first.ID, "expected_before": before}, 200)
	var afterNoop int
	authdb.Require(t, "audit sau no-op", f.DB.Conn.QueryRow(t.Context(), `SELECT count(*) FROM audit_events`).Scan(&afterNoop))
	if audits != afterNoop {
		t.Fatal("no-op ghi audit giả")
	}
	_, raw := catalogCall(t, a, "PATCH", skillPath, cookie, token, map[string]any{"name": "Tên mới", "category_id": second.ID, "expected_before": before}, 200)
	var updated skillCatalogDTO
	authdb.Require(t, "kỹ năng chuyển nhóm", json.Unmarshal(raw, &updated))
	if updated.ID != skill.ID || updated.Name != "Tên mới" || updated.Category == nil || updated.Category.ID != second.ID {
		t.Fatal("đổi tên/nhóm mất ID hoặc sai nhóm")
	}
	_, raw = catalogCall(t, a, "PATCH", skillPath, cookie, token, map[string]any{"name": "Không ghi đè", "category_id": first.ID, "expected_before": before}, 409)
	assertErrorCode(t, raw, "STALE_DATA")
	_, raw = catalogCall(t, a, "POST", "/api/skills", cookie, token, map[string]any{"name": "TÊN MỚI", "category_id": first.ID}, 409)
	assertErrorCode(t, raw, "DUPLICATE_NAME")
	_, raw = catalogCall(t, a, "DELETE", "/api/skill-categories/"+strconv.Itoa(second.ID), cookie, token, map[string]any{"expected_before": map[string]any{"name": second.Name}}, 409)
	assertErrorCode(t, raw, "CATALOG_IN_USE")
	catalogCall(t, a, "DELETE", skillPath, cookie, token, map[string]any{"expected_before": map[string]any{"name": updated.Name, "category_id": second.ID}}, 204)
	catalogCall(t, a, "DELETE", skillPath, cookie, token, map[string]any{"expected_before": map[string]any{"name": updated.Name, "category_id": second.ID}}, 404)
	categoryPath := "/api/skill-categories/" + strconv.Itoa(first.ID)
	catalogCall(t, a, "PATCH", categoryPath, cookie, token, map[string]any{"name": "Nhóm đổi tên", "expected_before": map[string]any{"name": first.Name}}, 200)
	_, raw = catalogCall(t, a, "DELETE", categoryPath, cookie, token, map[string]any{"expected_before": map[string]any{"name": first.Name}}, 409)
	assertErrorCode(t, raw, "STALE_DATA")
	catalogCall(t, a, "DELETE", categoryPath, cookie, token, map[string]any{"expected_before": map[string]any{"name": "Nhóm đổi tên"}}, 204)
	catalogCall(t, a, "DELETE", "/api/skill-categories/"+strconv.Itoa(second.ID), cookie, token, map[string]any{"expected_before": map[string]any{"name": second.Name}}, 204)
	_, err := f.DB.Conn.Exec(t.Context(), `DELETE FROM user_roles WHERE user_id=$1 AND role_code='DIRECTOR'`, director.UserID)
	authdb.Require(t, "thu quyền Giám đốc hiện hành", err)
	catalogCall(t, a, "GET", "/api/skills", cookie, "", nil, 200)
	_, raw = catalogCall(t, a, "POST", "/api/skill-categories", cookie, token, map[string]any{"name": "Quyền đã thu hồi"}, 403)
	assertErrorCode(t, raw, "FORBIDDEN")
}

func TestSkillCatalogApplicationChunkedBodiesAndSecurity(t *testing.T) {
	a := newAuthApp(t)
	a.fixture.AddAccount(t, authdb.AccountOptions{Email: "skill-body@example.invalid", Roles: []string{"DIRECTOR"}})
	cookie := a.login(t, "skill-body@example.invalid", authdb.FixturePassword, nil)
	token := a.csrf(t, cookie)
	_, raw := catalogCall(t, a, "POST", "/api/skill-categories", cookie, token, map[string]any{"name": "Nhóm cho kỹ năng"}, 201)
	var category catalogDTO
	authdb.Require(t, "nhóm fixture", json.Unmarshal(raw, &category))
	for _, collection := range []string{"skill-categories", "skills"} {
		path := "/api/" + collection
		createBody := map[string]any{"name": "Before"}
		if collection == "skills" {
			createBody["category_id"] = category.ID
		}
		_, raw = catalogCall(t, a, "POST", path, cookie, token, createBody, 201)
		var item skillCatalogDTO
		authdb.Require(t, "mục fixture body", json.Unmarshal(raw, &item))
		for _, method := range []string{"POST", "PATCH", "DELETE"} {
			target := path
			body := map[string]any{"name": "Chunked"}
			if collection == "skills" {
				body["category_id"] = category.ID
			}
			want := 201
			if method != "POST" {
				target += "/" + strconv.Itoa(item.ID)
				before := map[string]any{"name": "Before"}
				if method == "DELETE" {
					before["name"] = "Renamed"
				}
				if collection == "skills" {
					before["category_id"] = category.ID
				}
				body = map[string]any{"expected_before": before}
				want = 204
				if method == "PATCH" {
					body["name"] = "Renamed"
					if collection == "skills" {
						body["category_id"] = category.ID
					}
					want = 200
				}
			}
			encoded, err := json.Marshal(body)
			authdb.Require(t, "encode body", err)
			padded := string(encoded) + strings.Repeat(" ", 8192-len(encoded))
			for _, size := range []int{8193, 8192} {
				payload := padded
				expected := want
				if size == 8193 {
					payload += " "
					expected = 413
				}
				req, err := http.NewRequestWithContext(t.Context(), method, a.server.URL+target, io.NopCloser(strings.NewReader(payload)))
				authdb.Require(t, "HTTP body không Content-Length", err)
				req.ContentLength = -1
				req.AddCookie(cookie)
				req.Header.Set("Origin", testOrigin)
				req.Header.Set("X-CSRF-Token", token)
				req.Header.Set("Content-Type", "application/json")
				response, err := a.server.Client().Do(req)
				authdb.Require(t, "HTTP chunked", err)
				data, err := io.ReadAll(response.Body)
				response.Body.Close()
				authdb.Require(t, "đọc response chunked", err)
				if response.StatusCode != expected || response.Header.Get("Cache-Control") != "no-store" || response.Header.Get("Set-Cookie") != "" || (expected == 204 && len(data) != 0) {
					t.Fatalf("%s %s %d byte: nhận%d, cần%d", method, collection, size, response.StatusCode, expected)
				}
				if size == 8193 {
					assertErrorCode(t, data, "PAYLOAD_TOO_LARGE")
				}
			}
		}
	}
	for _, tc := range []struct {
		body, contentType, origin, csrf string
		status                          int
		code                            string
	}{
		{`{"name":"Sai nguồn"}`, "application/json", "https://wrong.invalid", token, 403, "FORBIDDEN"},
		{`{"name":"Thiếu CSRF"}`, "application/json", testOrigin, "", 403, "FORBIDDEN"},
		{`{"name":"Sai media"}`, "text/plain", testOrigin, token, 415, "UNSUPPORTED_MEDIA_TYPE"},
		{`{"name":"Sai charset"}`, "application/json; charset=latin1", testOrigin, token, 415, "UNSUPPORTED_MEDIA_TYPE"},
		{`{"name":"Tên","actor_id":1}`, "application/json", testOrigin, token, 400, "VALIDATION_ERROR"},
		{`{"name":"Một","name":"Hai"}`, "application/json", testOrigin, token, 400, "VALIDATION_ERROR"},
		{`{"name":"\ud800"}`, "application/json", testOrigin, token, 400, "VALIDATION_ERROR"},
		{`{"name":"\u0000"}`, "application/json", testOrigin, token, 400, "VALIDATION_ERROR"},
		{`{"name":"` + strings.Repeat("ế", 101) + `"}`, "application/json", testOrigin, token, 400, "VALIDATION_ERROR"},
	} {
		req, err := http.NewRequestWithContext(t.Context(), "POST", a.server.URL+"/api/skill-categories", strings.NewReader(tc.body))
		authdb.Require(t, "HTTP security fixture", err)
		req.AddCookie(cookie)
		req.Header.Set("Content-Type", tc.contentType)
		req.Header.Set("Origin", tc.origin)
		req.Header.Set("X-CSRF-Token", tc.csrf)
		response, err := a.server.Client().Do(req)
		authdb.Require(t, "HTTP security", err)
		data, err := io.ReadAll(response.Body)
		response.Body.Close()
		authdb.Require(t, "response security", err)
		if response.StatusCode != tc.status || response.Header.Get("Cache-Control") != "no-store" || response.Header.Get("Set-Cookie") != "" {
			t.Fatal("status/header security sai")
		}
		assertErrorCode(t, data, tc.code)
	}
}

func TestSkillCatalogApplicationAuditFailure(t *testing.T) {
	a := newAuthApp(t)
	a.fixture.AddAccount(t, authdb.AccountOptions{Email: "skill-audit@example.invalid", Roles: []string{"DIRECTOR"}})
	cookie := a.login(t, "skill-audit@example.invalid", authdb.FixturePassword, nil)
	token := a.csrf(t, cookie)
	_, err := a.fixture.DB.Conn.Exec(t.Context(), `CREATE FUNCTION fail_skill_catalog_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic private audit fault'; END $$; CREATE TRIGGER fail_skill_catalog_audit BEFORE INSERT ON audit_events FOR EACH ROW EXECUTE FUNCTION fail_skill_catalog_audit()`)
	authdb.Require(t, "fault audit chỉ DB case", err)
	_, raw := catalogCall(t, a, "POST", "/api/skill-categories", cookie, token, map[string]any{"name": "Không được lưu"}, 500)
	assertErrorCode(t, raw, "INTERNAL_ERROR")
	if strings.Contains(string(raw), "synthetic") || strings.Contains(string(raw), "audit_events") {
		t.Fatal("lộ lỗi DB")
	}
	var rows int
	authdb.Require(t, "đối soát rollback audit", a.fixture.DB.Conn.QueryRow(t.Context(), `SELECT count(*) FROM skill_categories`).Scan(&rows))
	if rows != 0 {
		t.Fatal("audit lỗi vẫn giữ thay đổi danh mục")
	}
}

func TestSkillCatalogApplicationDatabaseFailures(t *testing.T) {
	a := newAuthApp(t)
	a.fixture.AddAccount(t, authdb.AccountOptions{Email: "skill-db-error@example.invalid", Roles: []string{"DIRECTOR"}})
	cookie := a.login(t, "skill-db-error@example.invalid", authdb.FixturePassword, nil)
	_, err := a.fixture.DB.Conn.Exec(t.Context(), `ALTER TABLE public.skills RENAME TO skills_fault_case`)
	authdb.Require(t, "fault query chỉ trên DB case", err)
	_, raw := catalogCall(t, a, "GET", "/api/skills", cookie, "", nil, 500)
	assertErrorCode(t, raw, "INTERNAL_ERROR")
	if strings.Contains(string(raw), "skills_fault_case") || strings.Contains(string(raw), "relation") {
		t.Fatal("response lộ lỗi DB")
	}
	_, err = a.fixture.DB.Conn.Exec(t.Context(), `ALTER TABLE public.skills_fault_case RENAME TO skills`)
	authdb.Require(t, "khôi phục bảng sau fault", err)
	blocker, err := a.fixture.DB.Conn.Begin(t.Context())
	authdb.Require(t, "transaction chặn query", err)
	defer blocker.Rollback(t.Context())
	_, err = blocker.Exec(t.Context(), `LOCK TABLE public.skills IN ACCESS EXCLUSIVE MODE`)
	authdb.Require(t, "khóa bảng để kiểm timeout PG thật", err)
	_, raw = catalogCall(t, a, "GET", "/api/skills", cookie, "", nil, 503)
	assertErrorCode(t, raw, "SERVICE_UNAVAILABLE")
	authdb.Require(t, "nhả khóa sau timeout", blocker.Rollback(t.Context()))
	catalogCall(t, a, "GET", "/api/skills", cookie, "", nil, 200)
}
