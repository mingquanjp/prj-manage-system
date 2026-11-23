//go:build integration

package postgres_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	pc "ai-staffing/internal/service/profilechange"
	"ai-staffing/internal/testutil/authdb"
)

func pcPending(t *testing.T, f *authdb.Fixture, a authdb.Account, n int) pc.Detail {
	t.Helper()
	d, err := pcService(t, f.DB.Conn).Submit(t.Context(), pcInput(a, n))
	authdb.Require(t, "tạo pending tổng hợp", err)
	return d
}

func pcDecisionCount(t *testing.T, f *authdb.Fixture) int {
	t.Helper()
	return catalogCount(t, f, `SELECT count(*) FROM notifications WHERE notification_type IN ('PROFILE_CHANGE_APPROVED','PROFILE_CHANGE_REJECTED')`)
}

func TestProfileChangeDecisionApproveRejectAndReplay(t *testing.T) {
	f := authdb.NewFixture(t)
	hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR", "ADMIN", "DIRECTOR"}})
	a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
	s := pcService(t, f.DB.Conn)
	in := pcInput(a, 1)
	in.Before["phone"] = nil
	in.Proposed["phone"] = pcText("12345")
	d, e := s.Submit(t.Context(), in)
	authdb.Require(t, "pending", e)
	d, e = s.Approve(t.Context(), pc.DetailQuery{ActorID: hr.UserID, RequestID: d.RequestID})
	authdb.Require(t, "approve thật", e)
	if d.Status != "APPROVED" || d.ResolvedBy == nil || *d.ResolvedBy != hr.UserID || d.ResolvedAt == nil || d.DecisionReason != nil || !reflect.DeepEqual(d.Proposed, d.Current) || d.Before["address"] != nil {
		t.Fatal("approve projection/NULL/kết luận sai")
	}
	if catalogCount(t, f, `SELECT count(*) FROM employees WHERE employee_id=$1 AND address=$2 AND phone='12345'`, a.EmployeeID, *in.Proposed["address"]) != 1 {
		t.Fatal("approve chưa ghi đủ patch official")
	}
	if catalogCount(t, f, `SELECT count(*) FROM notifications WHERE recipient_id=$1 AND resource_id=$2 AND resource_type='PROFILE_CHANGE_REQUEST' AND notification_type='PROFILE_CHANGE_APPROVED' AND message NOT LIKE '%12345%' AND message NOT LIKE '%Địa chỉ%'`, a.UserID, d.RequestID) != 1 {
		t.Fatal("notification quyết định sai recipient hoặc lộ dữ liệu patch")
	}
	for _, reject := range []bool{false, true} {
		if reject {
			_, e = s.Reject(t.Context(), pc.RejectInput{ActorID: hr.UserID, RequestID: d.RequestID, Reason: "Không xử lý lại"})
		} else {
			_, e = s.Approve(t.Context(), pc.DetailQuery{ActorID: hr.UserID, RequestID: d.RequestID})
		}
		pcError(t, e, pc.ErrAlreadyResolved)
	}
	replay, e := s.Submit(t.Context(), in)
	authdb.Require(t, "replay sau approve thật", e)
	if replay.RequestID != d.RequestID || replay.Status != "APPROVED" || pcDecisionCount(t, f) != 1 {
		t.Fatal("replay tạo tác động hoặc dựng pending cũ")
	}
	second := pcInput(a, 2)
	second.Before = pc.Values{"phone": pcText("12345")}
	second.Proposed = pc.Values{"phone": nil}
	d, e = s.Submit(t.Context(), second)
	authdb.Require(t, "pending xóa phone", e)
	pcExec(t, f, `UPDATE employees SET phone='official đã đổi' WHERE employee_id=$1`, a.EmployeeID)
	d, e = s.Reject(t.Context(), pc.RejectInput{ActorID: hr.UserID, RequestID: d.RequestID, Reason: " \u00a0Cần đối chiếu lại\u3000 "})
	authdb.Require(t, "reject vẫn được khi before stale", e)
	if d.Status != "REJECTED" || d.DecisionReason == nil || *d.DecisionReason != "Cần đối chiếu lại" || *d.Current["phone"] != "official đã đổi" || *d.ResolvedBy != hr.UserID {
		t.Fatal("reject sửa official hoặc lý do sai")
	}
	_, e = s.Approve(t.Context(), pc.DetailQuery{ActorID: hr.UserID, RequestID: d.RequestID})
	pcError(t, e, pc.ErrAlreadyResolved)
	_, e = s.Reject(t.Context(), pc.RejectInput{ActorID: hr.UserID, RequestID: d.RequestID, Reason: "Lần hai"})
	pcError(t, e, pc.ErrAlreadyResolved)
	if pcDecisionCount(t, f) != 2 || catalogCount(t, f, `SELECT count(*) FROM audit_events`) != 0 {
		t.Fatal("terminal bị xử lý lại hoặc tạo audit ngoài phạm vi")
	}
}

func TestProfileChangeDecisionStaleAndDisjointFields(t *testing.T) {
	f := authdb.NewFixture(t)
	hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR"}})
	a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
	s := pcService(t, f.DB.Conn)
	first, same := pcPending(t, f, a, 1), pcPending(t, f, a, 2)
	in := pcInput(a, 3)
	in.Before, in.Proposed = pc.Values{"phone": nil}, pc.Values{"phone": pcText("54321")}
	disjoint, e := s.Submit(t.Context(), in)
	authdb.Require(t, "pending khác trường", e)
	_, e = s.Approve(t.Context(), pc.DetailQuery{ActorID: hr.UserID, RequestID: first.RequestID})
	authdb.Require(t, "approve đầu", e)
	_, e = s.Approve(t.Context(), pc.DetailQuery{ActorID: hr.UserID, RequestID: same.RequestID})
	pcError(t, e, pc.ErrStale)
	if catalogCount(t, f, `SELECT count(*) FROM employee_profile_change_requests WHERE request_id=$1 AND status='PENDING' AND resolved_by IS NULL`, same.RequestID) != 1 || pcDecisionCount(t, f) != 1 {
		t.Fatal("stale đã đổi pending hoặc notify")
	}
	_, e = s.Approve(t.Context(), pc.DetailQuery{ActorID: hr.UserID, RequestID: disjoint.RequestID})
	authdb.Require(t, "khác trường không stale giả", e)
	if catalogCount(t, f, `SELECT count(*) FROM employees WHERE employee_id=$1 AND address IS NOT NULL AND phone='54321'`, a.EmployeeID) != 1 {
		t.Fatal("approve patch khác trường làm mất dữ liệu")
	}
	// NULL và chuỗi rỗng khác nhau, không trim/normalize dữ liệu official khi so.
	in = pcInput(a, 4)
	in.Before, in.Proposed = pc.Values{"phone": pcText("54321")}, pc.Values{"phone": nil}
	nullRequest, e := s.Submit(t.Context(), in)
	authdb.Require(t, "pending phone NULL", e)
	_, e = s.Approve(t.Context(), pc.DetailQuery{ActorID: hr.UserID, RequestID: nullRequest.RequestID})
	authdb.Require(t, "approve phone NULL", e)
	if catalogCount(t, f, `SELECT count(*) FROM employees WHERE employee_id=$1 AND phone IS NULL`, a.EmployeeID) != 1 {
		t.Fatal("NULL không được ghi đúng")
	}
	in = pcInput(a, 5)
	in.Before, in.Proposed = pc.Values{"phone": nil}, pc.Values{"phone": pcText("98765")}
	emptyRequest, e := s.Submit(t.Context(), in)
	authdb.Require(t, "pending phone từ NULL", e)
	pcExec(t, f, `UPDATE employees SET phone='' WHERE employee_id=$1`, a.EmployeeID)
	_, e = s.Approve(t.Context(), pc.DetailQuery{ActorID: hr.UserID, RequestID: emptyRequest.RequestID})
	pcError(t, e, pc.ErrStale)
	if catalogCount(t, f, `SELECT count(*) FROM employee_profile_change_requests WHERE request_id=$1 AND status='PENDING'`, emptyRequest.RequestID) != 1 || catalogCount(t, f, `SELECT count(*) FROM employees WHERE employee_id=$1 AND phone=''`, a.EmployeeID) != 1 {
		t.Fatal("NULL bị coi như chuỗi rỗng hoặc stale bị ghi đè")
	}
}

func TestProfileChangeDecisionChangedOwnerLinkageFailsClosed(t *testing.T) {
	f := authdb.NewFixture(t)
	hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR"}})
	a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
	if a.UserID == a.EmployeeID {
		t.Fatal("fixture phải phân biệt user_id và employee_id")
	}
	d := pcPending(t, f, a, 1)
	other := f.AddAccount(t, authdb.AccountOptions{Email: "other@example.invalid", WithoutEmployee: true})
	// Fault liên kết tổng hợp: request giữ requested_by cũ; không chuyển quyền/notify sang user mới.
	pcExec(t, f, `UPDATE employees SET user_id=$1 WHERE employee_id=$2`, other.UserID, a.EmployeeID)
	_, e := pcService(t, f.DB.Conn).Approve(t.Context(), pc.DetailQuery{ActorID: hr.UserID, RequestID: d.RequestID})
	pcError(t, e, pc.ErrPersistence)
	_, e = pcService(t, f.DB.Conn).Reject(t.Context(), pc.RejectInput{ActorID: hr.UserID, RequestID: d.RequestID, Reason: "Lý do"})
	pcError(t, e, pc.ErrPersistence)
	if pcDecisionCount(t, f) != 0 || catalogCount(t, f, `SELECT count(*) FROM employee_profile_change_requests WHERE request_id=$1 AND status='PENDING'`, d.RequestID) != 1 {
		t.Fatal("linkage sai vẫn quyết định hoặc notify")
	}
}

func TestProfileChangeDecisionTargetAndSelfRules(t *testing.T) {
	for _, mode := range []string{"inactive approve", "inactive reject", "locked approve", "locked reject", "limited target approve", "self approve", "self reject"} {
		t.Run(mode, func(t *testing.T) {
			f := authdb.NewFixture(t)
			hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR", "ADMIN"}})
			a := hr
			if !strings.HasPrefix(mode, "self") {
				a = f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
			}
			d := pcPending(t, f, a, 1)
			if strings.HasPrefix(mode, "inactive") {
				pcExec(t, f, `UPDATE employees SET employment_status='INACTIVE' WHERE employee_id=$1`, a.EmployeeID)
			}
			if strings.HasPrefix(mode, "locked") {
				pcExec(t, f, `UPDATE users SET status='LOCKED' WHERE user_id=$1`, a.UserID)
			}
			if strings.HasPrefix(mode, "limited target") {
				pcExec(t, f, `UPDATE users SET must_change_password=true WHERE user_id=$1`, a.UserID)
			}
			var e error
			if strings.HasSuffix(mode, "reject") {
				d, e = pcService(t, f.DB.Conn).Reject(t.Context(), pc.RejectInput{ActorID: hr.UserID, RequestID: d.RequestID, Reason: "Lý do tổng hợp"})
			} else {
				d, e = pcService(t, f.DB.Conn).Approve(t.Context(), pc.DetailQuery{ActorID: hr.UserID, RequestID: d.RequestID})
			}
			if mode == "inactive approve" {
				pcError(t, e, pc.ErrEmployeeInactive)
				if pcDecisionCount(t, f) != 0 || d.RequestID != 0 {
					t.Fatal("target inactive bị approve")
				}
			} else {
				authdb.Require(t, "quy tắc target/self", e)
				if pcDecisionCount(t, f) != 1 {
					t.Fatal("actor=owner hoặc user đích locked làm sai số thông báo")
				}
			}
		})
	}
}

func TestProfileChangeDecisionValidationAndGates(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
	hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR"}})
	admin := f.AddAccount(t, authdb.AccountOptions{Email: "admin@example.invalid", Roles: []string{"ADMIN", "DIRECTOR"}})
	d := pcPending(t, f, a, 1)
	s := pcService(t, f.DB.Conn)
	for _, actor := range []authdb.Account{a, admin} {
		_, e := s.Approve(t.Context(), pc.DetailQuery{ActorID: actor.UserID, RequestID: d.RequestID})
		pcError(t, e, pc.ErrForbidden)
		_, e = s.Reject(t.Context(), pc.RejectInput{ActorID: actor.UserID, RequestID: d.RequestID, Reason: "Lý do"})
		pcError(t, e, pc.ErrForbidden)
	}
	_, e := s.Approve(t.Context(), pc.DetailQuery{ActorID: hr.UserID, RequestID: 99999})
	pcError(t, e, pc.ErrNotFound)
	for _, reason := range []string{"", " \u00a0\u3000", strings.Repeat("ế", 2001), "bad\x00", string([]byte{0xff})} {
		_, e = s.Reject(t.Context(), pc.RejectInput{ActorID: hr.UserID, RequestID: d.RequestID, Reason: reason})
		pcError(t, e, pc.ErrValidation)
	}
	if pcDecisionCount(t, f) != 0 {
		t.Fatal("validation/gate lỗi đã notify")
	}
	_, e = s.Reject(t.Context(), pc.RejectInput{ActorID: hr.UserID, RequestID: d.RequestID, Reason: strings.Repeat("ế", 2000)})
	authdb.Require(t, "lý do Unicode đúng biên", e)
}

func TestProfileChangeDecisionFaultsAndCommitOutcomes(t *testing.T) {
	for _, reject := range []bool{false, true} {
		for _, mode := range []string{"notification", "deferred commit", "known rollback", "lost commit"} {
			t.Run(map[bool]string{false: "approve/", true: "reject/"}[reject]+mode, func(t *testing.T) {
				f := authdb.NewFixture(t)
				hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR"}})
				a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
				pending := pcPending(t, f, a, 1)
				var db accountBeginner = f.DB.Conn
				switch mode {
				case "notification", "deferred commit":
					pcExec(t, f, `CREATE FUNCTION pca_decision_fault() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic fault' USING ERRCODE='23514'; END $$`)
					if mode == "notification" {
						pcExec(t, f, `CREATE TRIGGER pca_decision_fault BEFORE INSERT ON notifications FOR EACH ROW EXECUTE FUNCTION pca_decision_fault()`)
					} else {
						pcExec(t, f, `CREATE CONSTRAINT TRIGGER pca_decision_fault AFTER INSERT ON notifications DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION pca_decision_fault()`)
					}
				case "known rollback":
					db = accountCommitRollbackDB{f.DB.Conn}
				case "lost commit":
					db = catalogHookDB{db: f.DB.Conn, commitLost: true}
				}
				var d pc.Detail
				var e error
				if reject {
					d, e = pcService(t, db).Reject(t.Context(), pc.RejectInput{ActorID: hr.UserID, RequestID: pending.RequestID, Reason: "Lý do tổng hợp"})
				} else {
					d, e = pcService(t, db).Approve(t.Context(), pc.DetailQuery{ActorID: hr.UserID, RequestID: pending.RequestID})
				}
				if d.RequestID != 0 {
					t.Fatal("fault trả success")
				}
				wantDecisions := 0
				wantState := "PENDING"
				if mode == "lost commit" {
					pcError(t, e, pc.ErrCommitUnknown)
					wantDecisions, wantState = 1, "APPROVED"
					if reject {
						wantState = "REJECTED"
					}
				} else {
					pcError(t, e, pc.ErrPersistence)
					if errors.Is(e, pc.ErrCommitUnknown) {
						t.Fatal("rollback xác nhận bị gọi unknown")
					}
				}
				if pcDecisionCount(t, f) != wantDecisions || catalogCount(t, f, `SELECT count(*) FROM employee_profile_change_requests WHERE request_id=$1 AND status=$2`, pending.RequestID, wantState) != 1 {
					t.Fatal("fault không giữ atomicity/quyết định đúng một lần")
				}
				wantOfficial := 0
				if mode == "lost commit" && !reject {
					wantOfficial = 1
				}
				if catalogCount(t, f, `SELECT count(*) FROM employees WHERE employee_id=$1 AND address IS NOT NULL`, a.EmployeeID) != wantOfficial {
					t.Fatal("official không atomic với kết luận")
				}
				if mode == "lost commit" {
					_, e = pcService(t, f.DB.Conn).Approve(t.Context(), pc.DetailQuery{ActorID: hr.UserID, RequestID: pending.RequestID})
					pcError(t, e, pc.ErrAlreadyResolved)
					read, err := pcService(t, f.DB.Conn).Read(t.Context(), pc.DetailQuery{ActorID: hr.UserID, RequestID: pending.RequestID})
					authdb.Require(t, "GET đối soát sau commit mất phản hồi", err)
					if read.Status != wantState || pcDecisionCount(t, f) != 1 {
						t.Fatal("lost commit tự retry hoặc không đối soát được")
					}
				}
			})
		}
	}
}
