// Package notification cung cấp thao tác thông báo chính chủ, độc lập quyền tài nguyên.
package notification

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrValidation             = errors.New("notification: validation")
	ErrUnauthenticated        = errors.New("notification: unauthenticated")
	ErrPasswordChangeRequired = errors.New("notification: password change required")
	ErrNotFound               = errors.New("notification: not found")
	ErrPersistence            = errors.New("notification: persistence")
	ErrUnavailable            = errors.New("notification: unavailable")
	ErrCommitUnknown          = errors.New("notification: commit unknown")
)

type Item struct {
	ID                 int
	Type, ResourceType string
	ResourceID         int
	Message            string
	CreatedAt          time.Time
	ReadAt             *time.Time
}
type ListQuery struct {
	ActorID        int
	Filter         string
	Page, PageSize int
}

func (q ListQuery) Offset() int64 { return (int64(q.Page) - 1) * int64(q.PageSize) }

type ReadCommand struct{ ActorID, NotificationID int }
type Page struct {
	Items              []Item
	Page, PageSize     int
	Total, UnreadCount int64
}
type Repository interface {
	List(context.Context, ListQuery) (Page, error)
	MarkRead(context.Context, ReadCommand) error
}
type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func validID(id int) bool                 { return id > 0 && int64(id) <= math.MaxInt32 }
func ValidateList(q ListQuery) (ListQuery, error) {
	if !validID(q.ActorID) {
		return q, ErrUnauthenticated
	}
	if q.Filter == "" {
		q.Filter = "all"
	}
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 20
	}
	if (q.Filter != "all" && q.Filter != "unread") || !validID(q.Page) || q.PageSize < 1 || q.PageSize > 100 {
		return q, ErrValidation
	}
	return q, nil
}
func ValidateRead(q ReadCommand) error {
	if !validID(q.ActorID) {
		return ErrUnauthenticated
	}
	if !validID(q.NotificationID) {
		return ErrValidation
	}
	return nil
}
func (s *Service) List(ctx context.Context, q ListQuery) (Page, error) {
	q, err := ValidateList(q)
	if err != nil {
		return Page{}, err
	}
	p, err := s.repo.List(ctx, q)
	if err != nil {
		return Page{}, err
	}
	if p.Page != q.Page || p.PageSize != q.PageSize || p.Total < 0 || p.UnreadCount < 0 || len(p.Items) > q.PageSize || int64(len(p.Items)) > p.Total || (q.Filter == "unread" && p.Total != p.UnreadCount) {
		return Page{}, ErrPersistence
	}
	for _, n := range p.Items {
		if !validID(n.ID) || !validID(n.ResourceID) || !validResourcePair(n.Type, n.ResourceType) || !utf8.ValidString(n.Message) || strings.ContainsRune(n.Message, 0) || utf8.RuneCountInString(n.Message) < 1 || utf8.RuneCountInString(n.Message) > 200 || n.CreatedAt.IsZero() || (n.ReadAt != nil && n.ReadAt.Before(n.CreatedAt)) || (q.Filter == "unread" && n.ReadAt != nil) {
			return Page{}, ErrPersistence
		}
	}
	if p.Items == nil {
		p.Items = []Item{}
	}
	return p, nil
}

func validResourcePair(notificationType, resourceType string) bool {
	switch notificationType {
	case "PROFILE_CHANGE_SUBMITTED", "PROFILE_CHANGE_APPROVED", "PROFILE_CHANGE_REJECTED":
		return resourceType == "PROFILE_CHANGE_REQUEST"
	case "SKILL_CHANGE_SUBMITTED", "SKILL_CHANGE_APPROVED", "SKILL_CHANGE_REJECTED":
		return resourceType == "SKILL_CHANGE_REQUEST"
	default:
		return false
	}
}

func (s *Service) MarkRead(ctx context.Context, q ReadCommand) error {
	if err := ValidateRead(q); err != nil {
		return err
	}
	return s.repo.MarkRead(ctx, q)
}
