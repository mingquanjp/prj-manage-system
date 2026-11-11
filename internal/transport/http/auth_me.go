package http

import (
	"net/http"

	"ai-staffing/internal/service/auth"
)

type authIdentityResponse struct {
	UserID             int      `json:"user_id"`
	EmployeeID         int      `json:"employee_id"`
	EmployeeCode       string   `json:"employee_code"`
	Fullname           string   `json:"fullname"`
	SystemRoles        []string `json:"system_roles"`
	MustChangePassword bool     `json:"must_change_password"`
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromRequest(r)
	if !ok {
		writeAuthSessionError(w, r, auth.ErrUnauthenticated)
		return
	}

	writeAuthJSON(w, http.StatusOK, newAuthIdentityResponse(session.principal))
}

func newAuthIdentityResponse(principal auth.Principal) authIdentityResponse {
	return authIdentityResponse{
		UserID:             principal.UserID,
		EmployeeID:         principal.EmployeeID,
		EmployeeCode:       principal.EmployeeCode,
		Fullname:           principal.Fullname,
		SystemRoles:        principal.SystemRoles,
		MustChangePassword: principal.MustChangePassword,
	}
}
