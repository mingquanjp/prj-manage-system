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

func TestProjectionAcceptsProfileAndSkillTypes(t *testing.T) {
	now := time.Now()
	items := []Item{}
	for _, family := range []string{"PROFILE", "SKILL"} {
		for _, event := range []string{"SUBMITTED", "APPROVED", "REJECTED"} {
			items = append(items, Item{ID: len(items) + 1, Type: family + "_CHANGE_" + event, ResourceType: family + "_CHANGE_REQUEST", ResourceID: 7, Message: "Có cập nhật đề nghị.", CreatedAt: now})
		}
	}
	r := &stub{page: Page{Page: 1, PageSize: 20, Total: 6, UnreadCount: 6, Items: items}}
	p, err := NewService(r).List(t.Context(), ListQuery{ActorID: 1})
	if err != nil || len(p.Items) != 6 || p.Total != 6 || p.UnreadCount != 6 {
		t.Fatal("projection trộn hồ sơ và kỹ năng bị chặn")
	}
}

func TestProjectionRejectsInvalidTypeResourcePairs(t *testing.T) {
	for _, tc := range []struct{ notificationType, resourceType string }{
		{"PROFILE_CHANGE_SUBMITTED", "SKILL_CHANGE_REQUEST"},
		{"SKILL_CHANGE_APPROVED", "PROFILE_CHANGE_REQUEST"},
		{"SKILL_CHANGE_REJECTED", "EMPLOYEE"},
		{"SKILL_CHANGE_OTHER", "SKILL_CHANGE_REQUEST"},
		{"TASK_EVENT", "SKILL_CHANGE_REQUEST"},
		{"", "SKILL_CHANGE_REQUEST"},
		{"SKILL_CHANGE_SUBMITTED", ""},
	} {
		t.Run(tc.notificationType+"/"+tc.resourceType, func(t *testing.T) {
			r := &stub{page: Page{Page: 1, PageSize: 20, Total: 1, UnreadCount: 1, Items: []Item{{ID: 1, Type: tc.notificationType, ResourceType: tc.resourceType, ResourceID: 7, Message: "Có cập nhật đề nghị.", CreatedAt: time.Now()}}}}
			p, err := NewService(r).List(t.Context(), ListQuery{ActorID: 1})
			if !errors.Is(err, ErrPersistence) || len(p.Items) != 0 {
				t.Fatal("cặp loại và tài nguyên sai không được trả nội dung")
			}
		})
	}
}
