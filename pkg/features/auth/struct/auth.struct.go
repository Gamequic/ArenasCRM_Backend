package authstruct

import "github.com/golang-jwt/jwt"

type TokenStruct struct {
	jwt.StandardClaims
	Username  string `json:"username"`
	Email     string `json:"email"`
	Id        int    `json:"id"`
	SessionID string `json:"session_id"` // Agregamos SessionID
}

type LogIn struct {
	Email    string `validate:"required,email" json:"email"`
	Password string `validate:"required,min=8" json:"password"`
}

type Session struct {
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Token     string `json:"token"`
	CreateAt  string `json:"create_at,omitempty"`
	SessionID string `json:"session_id"`
}

type PasswordResetRequest struct {
	Email string `validate:"required,email" json:"email"`
}

type PasswordReset struct {
	Token       string `validate:"required" json:"token"`
	NewPassword string `validate:"required,min=8" json:"new_password"`
}
