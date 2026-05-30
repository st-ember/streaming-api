package handler

type SignupRequest struct {
	Username string `json:"user_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	RoleName string `json:"role"`
}
