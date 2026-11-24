package notification

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"
)

type stub struct {
	calls int
	page  Page
	err   error
}

func (s *stub) List(context.Context, ListQuery) (Page, error) { s.calls++; return s.page, s.err }
func (s *stub) MarkRead(context.Context, ReadCommand) error   { s.calls++; return s.err }
func TestValidationBoundariesAndOffset(t *testing.T) {
	r := &stub{}
	s := NewService(r)
	for _, q := range []ListQuery{{ActorID: 0}, {ActorID: 1, Filter: "other"}, {ActorID: 1, Page: -1}, {ActorID: 1, Page: math.MaxInt32 + 1}, {ActorID: 1, PageSize: 101}, {ActorID: 1, PageSize: -1}} {
		if _, e := s.List(t.Context(), q); e == nil {
			t.Fatal("input sai được nhận")
		}
	}
	for _, q := range []ReadCommand{{0, 1}, {1, 0}, {1, math.MaxInt32 + 1}} {
		if s.MarkRead(t.Context(), q) == nil {
			t.Fatal("ID sai được nhận")
		}
	}
	if r.calls != 0 {
		t.Fatal("validation lỗi đã gọi adapter")
	}
	q, e := ValidateList(ListQuery{ActorID: 1})
	if e != nil || q.Filter != "all" || q.Page != 1 || q.PageSize != 20 {
		t.Fatal("default sai")
	}
	q, e = ValidateList(ListQuery{ActorID: 1, Page: math.MaxInt32, PageSize: 100})
	if e != nil || q.Offset() != 214748364600 {
		t.Fatal("offset bị overflow")
	}
}
func TestProjectionFailsClosedAndErrorPreserved(t *testing.T) {
	r := &stub{page: Page{Page: 1, PageSize: 20, Total: 1, UnreadCount: 1, Items: []Item{{ID: 1, Type: "PROFILE_CHANGE_SUBMITTED", ResourceType: "PROFILE_CHANGE_REQUEST", ResourceID: 2, Message: "Có đề nghị hồ sơ.", CreatedAt: time.Now()}}}}
	s := NewService(r)
	if _, e := s.List(t.Context(), ListQuery{ActorID: 1}); e != nil {
		t.Fatal("projection hợp lệ bị chặn")
	}
	r.page.Items[0].Type = "TASK_EVENT"
	if _, e := s.List(t.Context(), ListQuery{ActorID: 1}); !errors.Is(e, ErrPersistence) {
		t.Fatal("loại ngoài phạm vi bị lộ")
	}
	r.err = errors.Join(ErrCommitUnknown, ErrUnavailable)
	if e := s.MarkRead(t.Context(), ReadCommand{1, 1}); !errors.Is(e, ErrCommitUnknown) || !errors.Is(e, ErrUnavailable) {
		t.Fatal("mất kết quả commit chưa rõ")
	}
}
