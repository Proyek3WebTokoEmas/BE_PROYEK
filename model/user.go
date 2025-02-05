package model

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Verified bool   `json:"verified"`
	Role      string `json:"role"`
}