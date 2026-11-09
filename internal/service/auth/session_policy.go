package auth

import "time"

// evaluateSession xác định danh tính và quyền hiệu lực tại thời điểm được truyền vào.
func evaluateSession(snapshot SessionSnapshot, now time.Time) (Principal, error) {
	if !snapshot.UserActive || !snapshot.EmployeeActive || snapshot.RevokedAt != nil || !now.Before(snapshot.ExpiresAt) {
		return Principal{}, ErrUnauthenticated
	}

	principal := snapshot.Principal
	principal.SystemRoles = []string{}
	if !principal.MustChangePassword {
		principal.SystemRoles = append(principal.SystemRoles, snapshot.Principal.SystemRoles...)
	}
	return principal, nil
}
