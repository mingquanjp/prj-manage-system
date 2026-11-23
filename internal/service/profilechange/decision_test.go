package profilechange

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func (f *fakeRepo) Approve(context.Context, DetailQuery) (Detail, error) {
	f.called = true
	return f.detail, f.err
}
func (f *fakeRepo) Reject(_ context.Context, in RejectInput) (Detail, error) {
	f.called = true
	if in.Reason != "Lý do tổng hợp" {
		return Detail{}, ErrValidation
	}
	return f.detail, f.err
}

func TestDecisionValidation(t *testing.T) {
	for _, reason := range []string{"", "\u00a0\u3000", "x\x00", string([]byte{0xff}), strings.Repeat("ế", 2001)} {
		f := &fakeRepo{}
		_, err := NewService(f).Reject(t.Context(), RejectInput{101, 1, reason})
		if !errors.Is(err, ErrValidation) || f.called {
			t.Fatal("lý do sai phải dừng trước repository")
		}
	}
	for _, q := range []DetailQuery{{0, 1}, {101, 0}, {101, -1}} {
		f := &fakeRepo{}
		_, err := NewService(f).Approve(t.Context(), q)
		if err == nil || f.called {
			t.Fatal("ID sai gọi repository")
		}
	}
	for _, n := range []int{1, 2000} {
		in, err := PrepareReject(RejectInput{101, 1, "\u3000" + strings.Repeat("ế", n) + "\u00a0"})
		if err != nil || in.Reason != strings.Repeat("ế", n) {
			t.Fatal("trim/code point sai")
		}
	}
	in := RejectInput{101, 1, "Lý do riêng tư"}
	if strings.Contains(fmt.Sprintf("%#v", in), in.Reason) {
		t.Fatal("fmt lộ reason")
	}
	if _, err := json.Marshal(in); err == nil {
		t.Fatal("không xuất input nội bộ")
	}
}

func TestDecisionServiceOutcomes(t *testing.T) {
	q := DetailQuery{101, 1}
	now := time.Now()
	actor := 101
	before := Values{"address": nil}
	proposed := Values{"address": text("Địa chỉ tổng hợp")}
	d := Detail{Summary: Summary{RequestID: 1, EmployeeID: 1001, EmployeeCode: "EMP000001", Fullname: "Fixture", Status: "APPROVED", ChangedFields: []string{"address"}, RequestedAt: now, ResolvedAt: &now}, Before: before, Proposed: proposed, Current: proposed, RequestedBy: 202, ResolvedBy: &actor}
	f := &fakeRepo{detail: d}
	s := NewService(f)
	if _, err := s.Approve(t.Context(), q); err != nil {
		t.Fatal("approve projection hợp lệ")
	}
	for _, mutate := range []func(*Detail){
		func(v *Detail) { v.RequestID = 2 }, func(v *Detail) { v.Status = "PENDING" },
		func(v *Detail) { v.ResolvedBy = textInt(202) }, func(v *Detail) { v.Current = before },
	} {
		f.detail = d
		mutate(&f.detail)
		if _, err := s.Approve(t.Context(), q); !errors.Is(err, ErrPersistence) {
			t.Fatal("projection sai không được pass")
		}
	}
	f.detail = d
	f.detail.Status = "REJECTED"
	f.detail.Current = before
	f.detail.DecisionReason = text("Lý do tổng hợp")
	if _, err := s.Reject(t.Context(), RejectInput{101, 1, " \u00a0Lý do tổng hợp\n"}); err != nil {
		t.Fatal("reject normalize/projection hợp lệ")
	}
	f.detail.DecisionReason = text("Lý do khác")
	if _, err := s.Reject(t.Context(), RejectInput{101, 1, "Lý do tổng hợp"}); !errors.Is(err, ErrPersistence) {
		t.Fatal("reason response sai")
	}
	for _, want := range []error{ErrAlreadyResolved, ErrEmployeeInactive, ErrStale, ErrForbidden, errors.Join(ErrCommitUnknown, ErrUnavailable)} {
		f.err = want
		for _, call := range []func() (Detail, error){func() (Detail, error) { return s.Approve(t.Context(), q) }, func() (Detail, error) { return s.Reject(t.Context(), RejectInput{101, 1, "Lý do tổng hợp"}) }} {
			got, err := call()
			if !errors.Is(err, want) || got.RequestID != 0 {
				t.Fatal("nuốt typed error hoặc trả success khi lỗi")
			}
		}
	}
}
func textInt(n int) *int { return &n }
