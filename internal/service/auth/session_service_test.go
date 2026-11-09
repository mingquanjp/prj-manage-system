package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"reflect"
	"testing"
	"time"
)

type fakeSessionRepository struct {
	read   func(context.Context, SessionHash) (SessionSnapshot, bool, error)
	revoke func(context.Context, SessionHash) error
}

func (f fakeSessionRepository) ReadSession(ctx context.Context, hash SessionHash) (SessionSnapshot, bool, error) {
	return f.read(ctx, hash)
}

func (f fakeSessionRepository) RevokeCurrentSession(ctx context.Context, hash SessionHash) error {
	return f.revoke(ctx, hash)
}

type sessionTestClock func() time.Time

func (c sessionTestClock) Now() time.Time { return c() }

func serviceTestSnapshot(expiresAt time.Time) SessionSnapshot {
	return SessionSnapshot{
		Principal: Principal{
			UserID:       12,
			EmployeeID:   24,
			EmployeeCode: "E024",
			Fullname:     "Nhân viên kiểm thử",
			SystemRoles:  []string{"ADMIN"},
		},
		CreatedAt:      expiresAt.Add(-8 * time.Hour),
		ExpiresAt:      expiresAt,
		UserActive:     true,
		EmployeeActive: true,
	}
}

func TestAuthenticateReadsCurrentSnapshot(t *testing.T) {
	now := time.Date(2026, time.November, 10, 16, 0, 0, 0, time.UTC)
	snapshot := serviceTestSnapshot(now.Add(time.Hour))
	secret := SessionSecret{0x01, 0x80, 0xff}
	wantHash := sha256.Sum256(secret[:])
	ctx := t.Context()
	reads := 0
	service := NewService(fakeSessionRepository{
		read: func(gotCtx context.Context, hash SessionHash) (SessionSnapshot, bool, error) {
			if gotCtx != ctx || hash != wantHash {
				t.Fatal("repository phải nhận đúng context và hash của phiên")
			}
			reads++
			return snapshot, true, nil
		},
	}, sessionTestClock(func() time.Time { return now }))

	principal, err := service.Authenticate(ctx, secret)
	if err != nil || !reflect.DeepEqual(principal, snapshot.Principal) {
		t.Fatalf("lần đọc đầu phải trả đúng principal, lỗi: %v", err)
	}
	snapshot.Principal.SystemRoles = []string{"HR"}
	principal, err = service.Authenticate(ctx, secret)
	if err != nil || !reflect.DeepEqual(principal, snapshot.Principal) {
		t.Fatalf("lần đọc tiếp theo phải trả quyền hiện hành, lỗi: %v", err)
	}
	snapshot.Principal.MustChangePassword = true
	want := snapshot.Principal
	want.SystemRoles = []string{}
	principal, err = service.Authenticate(ctx, secret)
	if err != nil || !reflect.DeepEqual(principal, want) {
		t.Fatalf("lần đọc tiếp theo phải áp dụng cờ buộc đổi hiện hành, lỗi: %v", err)
	}
	if reads != 3 {
		t.Fatalf("mỗi lần xác thực phải đọc một snapshot, nhận %d lần đọc", reads)
	}
}

func TestAuthenticateRejectsInvalidSessionAndPreservesErrors(t *testing.T) {
	expiresAt := time.Date(2026, time.November, 10, 17, 0, 0, 0, time.UTC)
	beforeExpiry := expiresAt.Add(-time.Second)
	tests := []struct {
		name             string
		missing          bool
		readErr          error
		revoked          bool
		userInactive     bool
		employeeInactive bool
		expiresWhileRead bool
		wantErr          error
	}{
		{name: "không tìm thấy phiên", missing: true, wantErr: ErrUnauthenticated},
		{name: "hết hạn trong lúc đọc", expiresWhileRead: true, wantErr: ErrUnauthenticated},
		{name: "phiên đã thu hồi", revoked: true, wantErr: ErrUnauthenticated},
		{name: "tài khoản ngừng hoạt động", userInactive: true, wantErr: ErrUnauthenticated},
		{name: "hồ sơ ngừng hoạt động", employeeInactive: true, wantErr: ErrUnauthenticated},
		{name: "kho dữ liệu không khả dụng", missing: true, readErr: ErrUnavailable, wantErr: ErrUnavailable},
		{name: "lỗi lưu trữ", missing: true, readErr: ErrPersistence, wantErr: ErrPersistence},
		{name: "snapshot hợp lệ trả kèm lỗi lưu trữ", readErr: ErrPersistence, wantErr: ErrPersistence},
		{name: "request bị hủy", missing: true, readErr: context.Canceled, wantErr: context.Canceled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := beforeExpiry
			snapshot := serviceTestSnapshot(expiresAt)
			snapshot.UserActive = !tt.userInactive
			snapshot.EmployeeActive = !tt.employeeInactive
			if tt.revoked {
				snapshot.RevokedAt = &beforeExpiry
			}
			service := NewService(fakeSessionRepository{
				read: func(context.Context, SessionHash) (SessionSnapshot, bool, error) {
					if tt.expiresWhileRead {
						now = expiresAt
					}
					return snapshot, !tt.missing, tt.readErr
				},
			}, sessionTestClock(func() time.Time { return now }))
			principal, err := service.Authenticate(t.Context(), SessionSecret{0x01})
			if !errors.Is(err, tt.wantErr) || !reflect.DeepEqual(principal, Principal{}) {
				t.Fatalf("cần principal rỗng và lỗi %v, nhận lỗi %v", tt.wantErr, err)
			}
		})
	}
}

func TestLogoutRevokesOnlyCallingSession(t *testing.T) {
	now := time.Date(2026, time.November, 10, 16, 0, 0, 0, time.UTC)
	current, other := SessionSecret{0x01}, SessionSecret{0x02}
	currentHash, otherHash := sha256.Sum256(current[:]), sha256.Sum256(other[:])
	sessions := map[SessionHash]SessionSnapshot{
		currentHash: serviceTestSnapshot(now.Add(time.Hour)),
		otherHash:   serviceTestSnapshot(now.Add(time.Hour)),
	}
	ctx := t.Context()
	revokes := 0
	service := NewService(fakeSessionRepository{
		read: func(_ context.Context, hash SessionHash) (SessionSnapshot, bool, error) {
			snapshot, found := sessions[hash]
			return snapshot, found, nil
		},
		revoke: func(gotCtx context.Context, hash SessionHash) error {
			if gotCtx != ctx || hash != currentHash {
				t.Fatal("port thu hồi phải nhận đúng context và hash của phiên đang gọi")
			}
			revokes++
			snapshot, found := sessions[hash]
			if !found || snapshot.RevokedAt != nil {
				return ErrUnauthenticated
			}
			snapshot.RevokedAt = &now
			sessions[hash] = snapshot
			return nil
		},
	}, sessionTestClock(func() time.Time { return now }))

	if err := service.Logout(ctx, current); err != nil || revokes != 1 {
		t.Fatalf("cần đúng một lần thu hồi thành công, nhận %d lần, lỗi %v", revokes, err)
	}
	if _, err := service.Authenticate(ctx, current); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("phiên vừa thu hồi phải bị từ chối, nhận lỗi %v", err)
	}
	if _, err := service.Authenticate(ctx, other); err != nil {
		t.Fatalf("phiên khác của cùng người dùng phải còn hiệu lực, nhận lỗi %v", err)
	}
	if err := service.Logout(ctx, current); !errors.Is(err, ErrUnauthenticated) || revokes != 2 {
		t.Fatalf("logout lại cần đúng một lần gọi port và lỗi phiên, nhận %d lần tổng, lỗi %v", revokes, err)
	}
}

func TestLogoutPreservesRepositoryErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "phiên không hợp lệ", err: ErrUnauthenticated},
		{name: "kho dữ liệu không khả dụng", err: ErrUnavailable},
		{name: "lỗi lưu trữ", err: ErrPersistence},
		{name: "request quá hạn", err: context.DeadlineExceeded},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(fakeSessionRepository{
				read: func(context.Context, SessionHash) (SessionSnapshot, bool, error) {
					t.Fatal("logout phải giao việc kiểm và thu hồi cho cùng một port")
					return SessionSnapshot{}, false, nil
				},
				revoke: func(context.Context, SessionHash) error { return tt.err },
			}, sessionTestClock(func() time.Time {
				t.Fatal("logout phải dùng thời điểm kiểm lại bên trong repository")
				return time.Time{}
			}))
			if err := service.Logout(t.Context(), SessionSecret{0x01}); !errors.Is(err, tt.err) {
				t.Fatalf("cần giữ lỗi %v, nhận %v", tt.err, err)
			}
		})
	}
}
