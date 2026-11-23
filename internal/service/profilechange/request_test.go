package profilechange

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

func text(s string) *string { return &s }
func input() SubmitInput {
	return SubmitInput{ActorID: 101, EmployeeID: 1001, Key: "abcdef01-1234-4234-8234-123456789abc", Before: Values{"address": nil}, Proposed: Values{"address": text("Địa chỉ tổng hợp")}}
}
func TestCanonicalTypedPayload(t *testing.T) {
	a := input()
	a.Before["phone"] = nil
	a.Proposed["phone"] = text("123")
	x, err := PrepareSubmit(a)
	if err != nil {
		t.Fatal("input hợp lệ")
	}
	b := input()
	b.Key = strings.ToUpper(b.Key)
	b.Before = Values{"phone": nil, "address": nil}
	b.Proposed = Values{"phone": text(" 123\u00a0"), "address": text("\u3000Địa chỉ tổng hợp\n")}
	b.Explanation = text("\u00a0\t")
	y, err := PrepareSubmit(b)
	if err != nil || x.Hash != y.Hash || y.Input.Key != a.Key {
		t.Fatal("thứ tự map/case key/normalize không canonical")
	}
	for _, edit := range []func(*SubmitInput){
		func(v *SubmitInput) { v.EmployeeID++ }, func(v *SubmitInput) { v.Before["phone"] = text("") },
		func(v *SubmitInput) { v.Explanation = text("Lý do khác") }, func(v *SubmitInput) { delete(v.Before, "phone"); delete(v.Proposed, "phone") },
	} {
		v := input()
		v.Before["phone"] = nil
		v.Proposed["phone"] = text("123")
		edit(&v)
		z, e := PrepareSubmit(v)
		if e != nil || z.Hash == x.Hash {
			t.Fatal("hash mất ID/null/giải trình/tập khóa")
		}
	}
	// Lệnh giữ bản sao payload, không bị thay đổi sau khi caller sửa map/pointer.
	*a.Proposed["phone"] = "modified"
	delete(a.Before, "address")
	if *x.Input.Proposed["phone"] != "123" || len(x.Input.Before) != 2 {
		t.Fatal("command giữ alias caller")
	}
	// Không tự NFC, code point khác nhau vẫn có nghĩa khác nhau.
	v := input()
	v.Proposed["address"] = text("é")
	one, _ := PrepareSubmit(v)
	v.Proposed["address"] = text("e\u0301")
	two, _ := PrepareSubmit(v)
	if one.Hash == two.Hash {
		t.Fatal("không được NFC ngầm")
	}
}
func TestSubmitValidation(t *testing.T) {
	for _, edit := range []func(*SubmitInput){
		func(v *SubmitInput) { v.ActorID = 0 }, func(v *SubmitInput) { v.EmployeeID = 0 }, func(v *SubmitInput) { v.EmployeeID = math.MaxInt32 + 1 },
		func(v *SubmitInput) { v.Key = "00000000-0000-1000-8000-000000000000" }, func(v *SubmitInput) { v.Key = "abcdef01-1234-4234-7234-123456789abc" }, func(v *SubmitInput) { v.Key += " " },
		func(v *SubmitInput) { v.Before = nil }, func(v *SubmitInput) { v.Proposed["role"] = text("HR") }, func(v *SubmitInput) { v.Before["phone"] = nil },
		func(v *SubmitInput) { v.Proposed["address"] = nil }, func(v *SubmitInput) { v.Proposed["address"] = text("\u00a0\u3000") }, func(v *SubmitInput) { v.Proposed["address"] = text("x\x00") },
		func(v *SubmitInput) { v.Proposed["address"] = text(string([]byte{0xff})) }, func(v *SubmitInput) { v.Explanation = text("x\x00") }, func(v *SubmitInput) { v.Explanation = text(strings.Repeat("ế", 2001)) },
		func(v *SubmitInput) { v.Proposed["address"] = text("same"); v.Before["address"] = text("same") },
	} {
		v := input()
		edit(&v)
		if _, err := PrepareSubmit(v); err == nil {
			t.Fatal("input sai được chấp nhận")
		}
	}
	for field, limit := range map[string]int{"fullname": 100, "phone": 20, "address": 2000} {
		v := input()
		v.Before = Values{field: text("legacy")}
		v.Proposed = Values{field: text(strings.Repeat("ế", limit))}
		if _, err := PrepareSubmit(v); err != nil {
			t.Fatal("biên Unicode code point hợp lệ")
		}
		v.Proposed[field] = text(strings.Repeat("ế", limit+1))
		if _, err := PrepareSubmit(v); err == nil {
			t.Fatal("vượt code point")
		}
	}
	v := input()
	v.Before = Values{"phone": text("  legacy  ")}
	v.Proposed = Values{"phone": nil}
	if _, err := PrepareSubmit(v); err != nil {
		t.Fatal("phone NULL là xóa")
	}
	v = input()
	v.Before["address"] = text(strings.Repeat("x", 3000))
	if _, err := PrepareSubmit(v); err != nil {
		t.Fatal("before legacy không áp giới hạn mới")
	}
}
func TestStoredValuesAndRedaction(t *testing.T) {
	v := Values{"phone": nil, "address": text("Giá trị tổng hợp")}
	b, err := EncodeStoredValues(v)
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecodeStoredValues(b, false)
	if err != nil || !reflect.DeepEqual(out, v) {
		t.Fatal("stored values mất NULL/keys")
	}
	for _, raw := range []string{`null`, `[]`, `{}`, `{"schema_version":2,"address":null}`, `{"schema_version":1,"role":"HR"}`, `{"schema_version":1,"phone":123}`, `{"schema_version":1,"phone":true}`, `{"schema_version":1,"fullname":null}`} {
		if _, err := DecodeStoredValues([]byte(raw), false); !errors.Is(err, ErrPersistence) {
			t.Fatal("JSONB hỏng không được phát ra")
		}
	}
	cmd, _ := PrepareSubmit(input())
	if b, err := json.Marshal(cmd); err == nil || strings.Contains(string(b), "address") {
		t.Fatal("không xuất command nội bộ")
	}
	if strings.Contains(fmt.Sprintf("%#v", cmd), "Địa chỉ") {
		t.Fatal("fmt rò payload")
	}
}

type fakeRepo struct {
	called bool
	detail Detail
	err    error
}

func (f *fakeRepo) Submit(context.Context, SubmitCommand) (Detail, error) {
	f.called = true
	return f.detail, f.err
}
func (f *fakeRepo) Read(context.Context, DetailQuery) (Detail, error) {
	f.called = true
	return f.detail, f.err
}
func (f *fakeRepo) List(context.Context, ListQuery) (Page, error) {
	f.called = true
	return Page{}, f.err
}
func TestServiceValidationAndOutcomes(t *testing.T) {
	f := &fakeRepo{}
	s := NewService(f)
	v := input()
	v.Proposed = nil
	if _, err := s.Submit(t.Context(), v); err == nil || f.called {
		t.Fatal("input sai phải dừng trước repo")
	}
	for _, want := range []error{ErrKeyExpired, ErrKeyReused, ErrStale, errors.Join(ErrCommitUnknown, ErrUnavailable)} {
		f.err = want
		d, e := s.Submit(t.Context(), input())
		if !errors.Is(e, want) || d.RequestID != 0 {
			t.Fatal("typed outcome bị nuốt")
		}
	}
	f.err = nil
	f.detail = Detail{}
	if _, err := s.Submit(t.Context(), input()); !errors.Is(err, ErrPersistence) {
		t.Fatal("projection hỏng không được pass")
	}
	v = input()
	now := time.Now()
	f.detail = Detail{Summary: Summary{RequestID: 1, EmployeeID: v.EmployeeID, EmployeeCode: "EMP000001", Fullname: "Fixture", Status: "PENDING", ChangedFields: []string{"address"}, RequestedAt: now}, Before: v.Before, Proposed: v.Proposed, Current: v.Before, RequestedBy: v.ActorID}
	if _, err := s.Submit(t.Context(), v); err != nil {
		t.Fatal("projection hợp lệ")
	}
	f.detail.ResolvedBy = new(int)
	if _, err := s.Read(t.Context(), DetailQuery{101, 1}); !errors.Is(err, ErrPersistence) {
		t.Fatal("pending có resolver phải lỗi")
	}
	q, e := ValidateList(ListQuery{ActorID: 101})
	if e != nil || q.Scope != "self" || q.Page != 1 || q.PageSize != 20 {
		t.Fatal("default list")
	}
	for _, q := range []ListQuery{{ActorID: 101, Scope: "all"}, {ActorID: 101, Status: "UNKNOWN"}, {ActorID: 101, Page: -1}, {ActorID: 101, PageSize: 101}} {
		if _, err := ValidateList(q); err == nil {
			t.Fatal("query sai")
		}
	}
}
