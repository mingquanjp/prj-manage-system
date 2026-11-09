package auth

import (
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"
)

func TestEvaluateSessionValidity(t *testing.T) {
	expiresAt := time.Date(2026, time.November, 9, 17, 0, 0, 0, time.UTC)
	beforeExpiry := expiresAt.Add(-time.Nanosecond)
	tests := []struct {
		name             string
		now              time.Time
		revokedAt        *time.Time
		userInactive     bool
		employeeInactive bool
		wantRejected     bool
	}{
		{name: "trước hạn", now: beforeExpiry},
		{name: "đúng hạn", now: expiresAt, wantRejected: true},
		{name: "sau hạn", now: expiresAt.Add(time.Nanosecond), wantRejected: true},
		{name: "đã thu hồi", now: beforeExpiry, revokedAt: &beforeExpiry, wantRejected: true},
		{name: "tài khoản ngừng hoạt động", now: beforeExpiry, userInactive: true, wantRejected: true},
		{name: "hồ sơ ngừng hoạt động", now: beforeExpiry, employeeInactive: true, wantRejected: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot := SessionSnapshot{
				Principal:      Principal{UserID: 12, EmployeeID: 24, SystemRoles: []string{"ADMIN"}},
				CreatedAt:      expiresAt.Add(-8 * time.Hour),
				ExpiresAt:      expiresAt,
				RevokedAt:      tt.revokedAt,
				UserActive:     !tt.userInactive,
				EmployeeActive: !tt.employeeInactive,
			}
			principal, err := evaluateSession(snapshot, tt.now)
			if tt.wantRejected {
				if !errors.Is(err, ErrUnauthenticated) || !reflect.DeepEqual(principal, Principal{}) {
					t.Fatal("phiên không hợp lệ phải trả ErrUnauthenticated và principal rỗng")
				}
				return
			}
			if err != nil || !reflect.DeepEqual(principal, snapshot.Principal) {
				t.Fatal("phiên còn hạn phải trả đúng principal")
			}
			if _, err := evaluateSession(snapshot, expiresAt); !errors.Is(err, ErrUnauthenticated) {
				t.Fatal("kiểm tra phiên thành công không được kéo dài hạn đã lưu")
			}
		})
	}
}

func TestEvaluateSessionEffectiveRoles(t *testing.T) {
	now := time.Date(2026, time.November, 9, 16, 0, 0, 0, time.UTC)
	tests := []struct {
		name               string
		storedRoles        []string
		mustChangePassword bool
		wantRoles          []string
	}{
		{name: "nhân viên không có quyền đặc biệt", wantRoles: []string{}},
		{name: "kiêm nhiệm", storedRoles: []string{"ADMIN", "HR"}, wantRoles: []string{"ADMIN", "HR"}},
		{name: "phiên hạn chế", storedRoles: []string{"DIRECTOR"}, mustChangePassword: true, wantRoles: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot := SessionSnapshot{
				Principal: Principal{
					UserID:             12,
					EmployeeID:         24,
					EmployeeCode:       "E024",
					Fullname:           "Nhân viên kiểm thử",
					SystemRoles:        tt.storedRoles,
					MustChangePassword: tt.mustChangePassword,
				},
				ExpiresAt:      now.Add(time.Hour),
				UserActive:     true,
				EmployeeActive: true,
			}
			want := snapshot.Principal
			want.SystemRoles = tt.wantRoles
			originalRoles := slices.Clone(snapshot.Principal.SystemRoles)
			principal, err := evaluateSession(snapshot, now)
			if err != nil || !reflect.DeepEqual(principal, want) {
				t.Fatal("principal phải giữ danh tính, cờ buộc đổi và trả đúng quyền hiệu lực")
			}
			if len(principal.SystemRoles) > 0 {
				principal.SystemRoles[0] = "DIRECTOR"
			}
			if !reflect.DeepEqual(snapshot.Principal.SystemRoles, originalRoles) {
				t.Fatal("quyền hiệu lực trả về không được làm thay đổi quyền trong snapshot")
			}
		})
	}
}
