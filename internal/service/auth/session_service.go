package auth

import (
	"context"
	"crypto/sha256"
)

// SessionSecret là bí mật phiên đã giải mã tại biên HTTP.
type SessionSecret [32]byte

type sessionRepository interface {
	ReadSession(context.Context, SessionHash) (SessionSnapshot, bool, error)
	RevokeCurrentSession(context.Context, SessionHash) error
}

type Service struct {
	repository sessionRepository
	clock      Clock
}

func NewService(repository sessionRepository, clock Clock) *Service {
	return &Service{repository: repository, clock: clock}
}

// Authenticate kiểm tra phiên từ snapshot hiện hành và trả danh tính hiệu lực.
func (s *Service) Authenticate(ctx context.Context, secret SessionSecret) (Principal, error) {
	snapshot, found, err := s.repository.ReadSession(ctx, sha256.Sum256(secret[:]))
	if err != nil {
		return Principal{}, err
	}
	if !found {
		return Principal{}, ErrUnauthenticated
	}
	return evaluateSession(snapshot, s.clock.Now())
}

// Logout giao việc kiểm lại hiệu lực và thu hồi đúng phiên cho repository.
func (s *Service) Logout(ctx context.Context, secret SessionSecret) error {
	return s.repository.RevokeCurrentSession(ctx, sha256.Sum256(secret[:]))
}
