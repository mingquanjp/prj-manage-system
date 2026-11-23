package profilechange

import (
	"context"
	"log/slog"
	"reflect"
	"strings"
	"unicode/utf8"
)

// RejectInput chỉ nhận actor từ nguồn server; reason không dùng làm log/DTO.
type RejectInput struct {
	ActorID, RequestID int
	Reason             string
}

func (RejectInput) String() string               { return "[profile decision redacted]" }
func (v RejectInput) GoString() string           { return v.String() }
func (RejectInput) LogValue() slog.Value         { return slog.StringValue("[profile decision redacted]") }
func (RejectInput) MarshalJSON() ([]byte, error) { return nil, ErrPersistence }

func PrepareReject(in RejectInput) (RejectInput, error) {
	if err := ValidateDetailQuery(DetailQuery{in.ActorID, in.RequestID}); err != nil {
		return RejectInput{}, err
	}
	if !validText(in.Reason) {
		return RejectInput{}, ErrValidation
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if n := utf8.RuneCountInString(in.Reason); n < 1 || n > 2000 {
		return RejectInput{}, ErrValidation
	}
	return in, nil
}

func (s *Service) Approve(ctx context.Context, q DetailQuery) (Detail, error) {
	if err := ValidateDetailQuery(q); err != nil {
		return Detail{}, err
	}
	d, err := s.repo.Approve(ctx, q)
	if err != nil {
		return Detail{}, err
	}
	if !validDecisionResult(d, q, "APPROVED") || !reflect.DeepEqual(d.Current, d.Proposed) {
		return Detail{}, ErrPersistence
	}
	return d, nil
}

func (s *Service) Reject(ctx context.Context, in RejectInput) (Detail, error) {
	in, err := PrepareReject(in)
	if err != nil {
		return Detail{}, err
	}
	d, err := s.repo.Reject(ctx, in)
	if err != nil {
		return Detail{}, err
	}
	if !validDecisionResult(d, DetailQuery{in.ActorID, in.RequestID}, "REJECTED") || d.DecisionReason == nil || *d.DecisionReason != in.Reason {
		return Detail{}, ErrPersistence
	}
	return d, nil
}

func validDecisionResult(d Detail, q DetailQuery, status string) bool {
	return ValidateDetail(d) == nil && d.RequestID == q.RequestID && d.Status == status && d.ResolvedBy != nil && *d.ResolvedBy == q.ActorID
}
