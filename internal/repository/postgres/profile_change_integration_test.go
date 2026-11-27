//go:build integration

package postgres_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	pc "ai-staffing/internal/service/profilechange"
	"ai-staffing/internal/testutil/authdb"
)

func pcText(s string) *string { return &s }
func pcService(t *testing.T, db accountBeginner) *pc.Service {
	t.Helper()
	r, e := postgres.NewProfileChangeRepository(db, 10*time.Second)
	authdb.Require(t, "adapter PCA", e)
	return pc.NewService(r)
}
func pcInput(a authdb.Account, n int) pc.SubmitInput {
	return pc.SubmitInput{ActorID: a.UserID, EmployeeID: a.EmployeeID, Key: fmt.Sprintf("00000000-0000-4000-8000-%012d", n), Before: pc.Values{"address": nil}, Proposed: pc.Values{"address": pcText("Địa chỉ tổng hợp")}}
}
func pcExec(t *testing.T, f *authdb.Fixture, sql string, args ...any) {
	t.Helper()
	_, e := f.DB.Conn.Exec(t.Context(), sql, args...)
	authdb.Require(t, "SQL PCA tổng hợp", e)
}
func pcNoWrites(t *testing.T, f *authdb.Fixture) {
	t.Helper()
	for _, table := range []string{"employee_profile_change_requests", "notifications", "idempotency_records"} {
		if catalogCount(t, f, "SELECT count(*) FROM public."+table) != 0 {
			t.Fatal("PCA để lại ghi dở")
		}
	}
}
func pcError(t *testing.T, e, want error) {
	t.Helper()
	if !errors.Is(e, want) {
		t.Fatalf("sai phân loại PCA, cần %v; không xuất dữ liệu/lỗi driver", want)
	}
}

func TestProfileChangeSubmitReplayAndRead(t *testing.T) {
	f := authdb.NewFixture(t)
	hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR", "ADMIN", "DIRECTOR"}})
	a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
	other := f.AddAccount(t, authdb.AccountOptions{Email: "other@example.invalid", Roles: []string{"ADMIN", "DIRECTOR"}})
	s := pcService(t, f.DB.Conn)
	in := pcInput(a, 1)
	in.Before["phone"] = nil
	in.Proposed["phone"] = pcText(" 12345 ")
	in.Explanation = pcText(" Giải trình tổng hợp ")
	d, e := s.Submit(t.Context(), in)
	authdb.Require(t, "gửi patch", e)
	if d.EmployeeID != a.EmployeeID || d.RequestedBy != a.UserID || d.Status != "PENDING" || d.Current["address"] != nil || *d.Proposed["phone"] != "12345" || *d.Explanation != "Giải trình tổng hợp" {
		t.Fatal("projection submit sai")
	}
	if catalogCount(t, f, `SELECT count(*) FROM employees WHERE employee_id=$1 AND address IS NULL AND phone IS NULL`, a.EmployeeID) != 1 {
		t.Fatal("submit đã sửa official")
	}
	if catalogCount(t, f, `SELECT count(*) FROM notifications WHERE recipient_id=$1 AND resource_id=$2 AND notification_type='PROFILE_CHANGE_SUBMITTED' AND resource_type='PROFILE_CHANGE_REQUEST' AND message NOT LIKE '%12345%'`, hr.UserID, d.RequestID) != 1 {
		t.Fatal("HR nhiều role phải nhận đúng một notification chung")
	}
	if catalogCount(t, f, `SELECT count(*) FROM notifications`) != 1 || catalogCount(t, f, `SELECT count(*) FROM audit_events`) != 0 {
		t.Fatal("notification/audit ngoài phạm vi")
	}
	if catalogCount(t, f, `SELECT count(*) FROM idempotency_records WHERE expires_at=created_at+interval '24h' AND octet_length(request_hash)=32 AND (SELECT count(*) FROM jsonb_object_keys(result_reference))=3`) != 1 {
		t.Fatal("key/result/TTL không đầy đủ")
	}
	pcExec(t, f, `UPDATE employees SET address='Official mới',phone='999' WHERE employee_id=$1`, a.EmployeeID)
	again, e := s.Submit(t.Context(), in)
	authdb.Require(t, "replay trước stale", e)
	if again.RequestID != d.RequestID || *again.Current["address"] != "Official mới" || len(again.Current) != 2 {
		t.Fatal("replay dựng snapshot giả hoặc mất ID")
	}
	// Dựng kết luận bằng SQL để kiểm replay; không chạy transaction HR duyệt.
	pcExec(t, f, `UPDATE employee_profile_change_requests SET status='REJECTED',resolved_by=$1,resolved_at=clock_timestamp(),decision_reason='Lý do tổng hợp' WHERE request_id=$2`, hr.UserID, d.RequestID)
	again, e = s.Submit(t.Context(), in)
	authdb.Require(t, "replay terminal hiện hành", e)
	if again.Status != "REJECTED" || again.RequestID != d.RequestID {
		t.Fatal("replay terminal bị giả pending")
	}
	if catalogCount(t, f, `SELECT count(*) FROM employee_profile_change_requests`) != 1 || catalogCount(t, f, `SELECT count(*) FROM notifications`) != 1 {
		t.Fatal("replay tạo tác động lần hai")
	}
	in.Proposed["phone"] = pcText("555")
	_, e = s.Submit(t.Context(), in)
	pcError(t, e, pc.ErrKeyReused)
	in.Proposed["phone"] = pcText("12345")
	pcExec(t, f, `UPDATE idempotency_records SET created_at=clock_timestamp()-interval '2 days',expires_at=clock_timestamp()-interval '1 day'`)
	// Cùng timestamp chính xác để bảo đảm record đúng TTL, không phụ thuộc chênh lệch hai clock calls.
	pcExec(t, f, `UPDATE idempotency_records SET expires_at=created_at+interval '24 hours'`)
	_, e = s.Submit(t.Context(), in)
	pcError(t, e, pc.ErrKeyExpired)
	for _, actor := range []authdb.Account{a, hr} {
		read, e := s.Read(t.Context(), pc.DetailQuery{ActorID: actor.UserID, RequestID: d.RequestID})
		authdb.Require(t, "self/HR đọc detail", e)
		if read.RequestID != d.RequestID || len(read.Current) != 2 {
			t.Fatal("current_values sai keyset")
		}
	}
	for _, id := range []int{d.RequestID, 99999} {
		_, e = s.Read(t.Context(), pc.DetailQuery{ActorID: other.UserID, RequestID: id})
		pcError(t, e, pc.ErrNotFound)
	}
	_, e = s.List(t.Context(), pc.ListQuery{ActorID: other.UserID, Scope: "review"})
	pcError(t, e, pc.ErrForbidden)
	p, e := s.List(t.Context(), pc.ListQuery{ActorID: hr.UserID})
	authdb.Require(t, "HR mặc định vẫn self", e)
	if p.Total != 0 {
		t.Fatal("HR mặc định thấy mọi người")
	}
	p, e = s.List(t.Context(), pc.ListQuery{ActorID: hr.UserID, Scope: "review", Status: "REJECTED"})
	authdb.Require(t, "HR review", e)
	if p.Total != 1 || len(p.Items) != 1 || !reflect.DeepEqual(p.Items[0].ChangedFields, []string{"address", "phone"}) {
		t.Fatal("review projection/filter sai")
	}
	p, e = s.List(t.Context(), pc.ListQuery{ActorID: a.UserID, Page: 99})
	authdb.Require(t, "page rỗng", e)
	if p.Total != 1 || len(p.Items) != 0 {
		t.Fatal("page rỗng mất total")
	}
}

func TestProfileChangeInputAndSelfFailuresDoNotWrite(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
	other := f.AddAccount(t, authdb.AccountOptions{Email: "other@example.invalid"})
	s := pcService(t, f.DB.Conn)
	for _, tc := range []struct {
		edit func(*pc.SubmitInput)
		want error
	}{
		{func(v *pc.SubmitInput) { v.EmployeeID = other.EmployeeID }, pc.ErrNotFound},
		{func(v *pc.SubmitInput) { v.Before["address"] = pcText("") }, pc.ErrStale},
		{func(v *pc.SubmitInput) { v.Proposed["address"] = nil }, pc.ErrValidation},
		{func(v *pc.SubmitInput) { v.Proposed["address"] = pcText(" \u00a0") }, pc.ErrValidation},
		{func(v *pc.SubmitInput) { v.Before["address"] = pcText("same"); v.Proposed["address"] = pcText("same") }, pc.ErrValidation},
		{func(v *pc.SubmitInput) { v.Proposed["role"] = pcText("HR") }, pc.ErrValidation},
	} {
		v := pcInput(a, 1)
		tc.edit(&v)
		d, e := s.Submit(t.Context(), v)
		pcError(t, e, tc.want)
		if d.RequestID != 0 {
			t.Fatal("lỗi có kết quả thành công")
		}
		pcNoWrites(t, f)
	}
	// Không có HR vẫn lưu pending; nhiều key được phép tạo nhiều pending cùng employee.
	for n := 1; n <= 2; n++ {
		_, e := s.Submit(t.Context(), pcInput(a, n))
		authdb.Require(t, "nhiều pending/no HR", e)
	}
	if catalogCount(t, f, `SELECT count(*) FROM notifications`) != 0 || catalogCount(t, f, `SELECT count(*) FROM employee_profile_change_requests`) != 2 {
		t.Fatal("áp unique pending hoặc ép phải có HR")
	}
	// Before legacy không bị cắt, phone NULL và address không xuất hiện là giữ nguyên.
	pcExec(t, f, `UPDATE employees SET address=$1,phone='  legacy  ' WHERE employee_id=$2`, strings.Repeat("ế", 2100), a.EmployeeID)
	v := pcInput(a, 3)
	v.Before = pc.Values{"phone": pcText("  legacy  ")}
	v.Proposed = pc.Values{"phone": nil}
	d, e := s.Submit(t.Context(), v)
	authdb.Require(t, "phone NULL", e)
	if len(d.Current) != 1 || *d.Current["phone"] != "  legacy  " {
		t.Fatal("tự normalize before/current hoặc kéo address vào patch")
	}
	v = pcInput(a, 4)
	v.Before["address"] = pcText(strings.Repeat("ế", 2100))
	d, e = s.Submit(t.Context(), v)
	authdb.Require(t, "before address legacy dài", e)
	if len(*d.Before["address"]) != len(strings.Repeat("ế", 2100)) {
		t.Fatal("cắt before legacy")
	}
}

func TestProfileChangeAtomicFaultsAndCommitOutcomes(t *testing.T) {
	for _, mode := range []string{"notification", "key", "deferred commit", "known rollback", "lost commit"} {
		t.Run(mode, func(t *testing.T) {
			f := authdb.NewFixture(t)
			a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid", Roles: []string{"HR"}})
			var db accountBeginner = f.DB.Conn
			switch mode {
			case "notification", "key", "deferred commit":
				table := "notifications"
				if mode != "notification" {
					table = "idempotency_records"
				}
				pcExec(t, f, `CREATE FUNCTION pca_fault() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic fault' USING ERRCODE='23514'; END $$`)
				trigger := "CREATE TRIGGER pca_fault BEFORE INSERT ON "
				ending := " FOR EACH ROW EXECUTE FUNCTION pca_fault()"
				if mode == "deferred commit" {
					trigger = "CREATE CONSTRAINT TRIGGER pca_fault AFTER INSERT ON "
					ending = " DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION pca_fault()"
				}
				pcExec(t, f, trigger+table+ending)
			case "known rollback":
				db = accountCommitRollbackDB{f.DB.Conn}
			case "lost commit":
				db = catalogHookDB{db: f.DB.Conn, commitLost: true}
			}
			d, e := pcService(t, db).Submit(t.Context(), pcInput(a, 1))
			if d.RequestID != 0 {
				t.Fatal("fault trả success")
			}
			if mode == "lost commit" {
				pcError(t, e, pc.ErrCommitUnknown)
				if catalogCount(t, f, `SELECT count(*) FROM employee_profile_change_requests`) != 1 || catalogCount(t, f, `SELECT count(*) FROM notifications`) != 1 || catalogCount(t, f, `SELECT count(*) FROM idempotency_records`) != 1 {
					t.Fatal("lost response không phải commit thật đúng một lần")
				}
				replay, e := pcService(t, f.DB.Conn).Submit(t.Context(), pcInput(a, 1))
				authdb.Require(t, "đối soát key sau lost commit", e)
				if replay.RequestID == 0 {
					t.Fatal("replay không tìm được result")
				}
			} else {
				pcError(t, e, pc.ErrPersistence)
				if errors.Is(e, pc.ErrCommitUnknown) {
					t.Fatal("rollback đã xác nhận bị gọi unknown")
				}
				pcNoWrites(t, f)
			}
		})
	}
}

func TestProfileChangeReadSnapshotAndCorruptJSON(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
	hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR"}})
	writer := pcService(t, f.DB.Connect(t))
	d, e := writer.Submit(t.Context(), pcInput(a, 1))
	authdb.Require(t, "pending đầu", e)
	inserted := false
	reader := pcService(t, catalogHookDB{db: f.DB.Connect(t), afterRow: func(sql string) {
		if !inserted && strings.HasPrefix(sql, "SELECT count(*)") {
			inserted = true
			_, e := writer.Submit(t.Context(), pcInput(a, 2))
			authdb.Require(t, "mutation xen count/items", e)
		}
	}})
	p, e := reader.List(t.Context(), pc.ListQuery{ActorID: hr.UserID, Scope: "review"})
	authdb.Require(t, "snapshot count/items", e)
	if !inserted || p.Total != 1 || len(p.Items) != 1 {
		t.Fatal("list dùng hai snapshot")
	}
	p, e = writer.List(t.Context(), pc.ListQuery{ActorID: hr.UserID, Scope: "review"})
	authdb.Require(t, "snapshot kế tiếp", e)
	if p.Total != 2 || len(p.Items) != 2 || p.Items[0].RequestID <= p.Items[1].RequestID {
		t.Fatal("sort hoặc snapshot mới sai")
	}
	// Fault dữ liệu bằng fixture: validator phải fail closed, không phát JSONB thô.
	pcExec(t, f, `ALTER TABLE employee_profile_change_requests DROP CONSTRAINT ck_pca_values`)
	pcExec(t, f, `UPDATE employee_profile_change_requests SET proposed_values='{"schema_version":99,"address":"invalid"}' WHERE request_id=$1`, d.RequestID)
	_, e = writer.Read(t.Context(), pc.DetailQuery{ActorID: a.UserID, RequestID: d.RequestID})
	pcError(t, e, pc.ErrPersistence)
	_, e = writer.List(t.Context(), pc.ListQuery{ActorID: a.UserID})
	pcError(t, e, pc.ErrPersistence)
}

func TestProfileChangeReadGatesAndReplayIntegrity(t *testing.T) {
	for _, scenario := range []string{"locked actor", "inactive actor", "limited actor", "role lost", "missing reference", "bad result", "bad hash", "bad TTL"} {
		t.Run(scenario, func(t *testing.T) {
			f := authdb.NewFixture(t)
			a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
			hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR", "ADMIN"}})
			s := pcService(t, f.DB.Conn)
			in := pcInput(a, 1)
			d, e := s.Submit(t.Context(), in)
			authdb.Require(t, "pending trước đổi điều kiện", e)
			switch scenario {
			case "locked actor", "inactive actor", "limited actor":
				sql := `UPDATE users SET status='LOCKED' WHERE user_id=$1`
				want := pc.ErrUnauthenticated
				if scenario == "inactive actor" {
					sql = `UPDATE employees SET employment_status='INACTIVE' WHERE user_id=$1`
				}
				if scenario == "limited actor" {
					sql = `UPDATE users SET must_change_password=true WHERE user_id=$1`
					want = pc.ErrPasswordChangeRequired
				}
				pcExec(t, f, sql, a.UserID)
				_, e = s.Read(t.Context(), pc.DetailQuery{ActorID: a.UserID, RequestID: d.RequestID})
				pcError(t, e, want)
				_, e = s.List(t.Context(), pc.ListQuery{ActorID: a.UserID})
				pcError(t, e, want)
				_, e = s.Submit(t.Context(), in)
				pcError(t, e, want)
				// Actor HR khác vẫn được đọc target inactive/locked; không dùng gate actor cho target.
				_, e = s.Read(t.Context(), pc.DetailQuery{ActorID: hr.UserID, RequestID: d.RequestID})
				authdb.Require(t, "HR đọc target không còn active", e)
			case "role lost":
				pcExec(t, f, `DELETE FROM user_roles WHERE user_id=$1 AND role_code='HR'`, hr.UserID)
				_, e = s.List(t.Context(), pc.ListQuery{ActorID: hr.UserID, Scope: "review"})
				pcError(t, e, pc.ErrForbidden)
				_, e = s.Read(t.Context(), pc.DetailQuery{ActorID: hr.UserID, RequestID: d.RequestID})
				pcError(t, e, pc.ErrNotFound)
			default:
				switch scenario {
				case "missing reference":
					pcExec(t, f, `DELETE FROM employee_profile_change_requests WHERE request_id=$1`, d.RequestID)
				case "bad result":
					pcExec(t, f, `UPDATE idempotency_records SET result_reference='{"schema_version":99,"request_id":1,"result_code":"CREATED"}'`)
				case "bad hash":
					pcExec(t, f, `ALTER TABLE idempotency_records DROP CONSTRAINT idempotency_records_request_hash_check`)
					pcExec(t, f, `UPDATE idempotency_records SET request_hash=decode('ab','hex')`)
				case "bad TTL":
					pcExec(t, f, `UPDATE idempotency_records SET expires_at=created_at+interval '1h'`)
				}
				_, e = s.Submit(t.Context(), in)
				pcError(t, e, pc.ErrPersistence)
				if catalogCount(t, f, `SELECT count(*) FROM idempotency_records`) != 1 {
					t.Fatal("integrity lỗi đã tự tạo lại key")
				}
			}
		})
	}
}

func TestProfileChangeReplayReferenceCannotPointToOtherPayload(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
	s := pcService(t, f.DB.Conn)
	one := pcInput(a, 1)
	_, e := s.Submit(t.Context(), one)
	authdb.Require(t, "request gốc", e)
	two := pcInput(a, 2)
	two.Proposed["address"] = pcText("Địa chỉ khác")
	d, e := s.Submit(t.Context(), two)
	authdb.Require(t, "request khác cùng owner", e)
	pcExec(t, f, `UPDATE idempotency_records SET result_reference=jsonb_set(result_reference,'{request_id}',to_jsonb($1::int)) WHERE idempotency_key=$2`, d.RequestID, one.Key)
	_, e = s.Submit(t.Context(), one)
	pcError(t, e, pc.ErrPersistence)
	if catalogCount(t, f, `SELECT count(*) FROM employee_profile_change_requests`) != 2 {
		t.Fatal("reference lỗi đã tạo thêm request")
	}
}

func TestProfileChangeUnicodeBoundariesThroughUseCase(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
	s := pcService(t, f.DB.Conn)
	n := 1
	for _, field := range []struct {
		name  string
		limit int
	}{{"fullname", 100}, {"phone", 20}, {"address", 2000}} {
		in := pcInput(a, n)
		n++
		in.Before = pc.Values{field.name: nil}
		if field.name == "fullname" {
			in.Before[field.name] = pcText(a.Fullname)
		}
		value := strings.Repeat("ế", field.limit)
		in.Proposed = pc.Values{field.name: pcText("\u00a0" + value + "\u3000")}
		d, e := s.Submit(t.Context(), in)
		authdb.Require(t, "biên Unicode qua service và PG", e)
		if *d.Proposed[field.name] != value {
			t.Fatal("trim/code point lệch khi persist")
		}
		baseline := catalogCount(t, f, `SELECT count(*) FROM employee_profile_change_requests`)
		in.Key = pcInput(a, n).Key
		n++
		in.Proposed[field.name] = pcText(value + "ế")
		_, e = s.Submit(t.Context(), in)
		pcError(t, e, pc.ErrValidation)
		if catalogCount(t, f, `SELECT count(*) FROM employee_profile_change_requests`) != baseline {
			t.Fatal("vượt giới hạn đã ghi DB")
		}
	}
	in := pcInput(a, n)
	in.Explanation = pcText(strings.Repeat("ế", 2000))
	d, e := s.Submit(t.Context(), in)
	authdb.Require(t, "explanation biên2000", e)
	if len(*d.Explanation) != len(strings.Repeat("ế", 2000)) {
		t.Fatal("explanation bị cắt")
	}
	in.Explanation = pcText(strings.Repeat("ế", 2001))
	_, e = s.Submit(t.Context(), in)
	pcError(t, e, pc.ErrValidation)
}
