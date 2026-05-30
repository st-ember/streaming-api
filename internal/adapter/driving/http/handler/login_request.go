package handler

type LoginRequest struct {
	Key      string `json:"key"`
	Password string `json:"password"`
}
