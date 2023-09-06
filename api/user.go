package api

import (
	db "bank/db/sqlc"
	"bank/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Create a new file called api/user.go. This file will contain the implementation of the user service.
// CreateUser creates a new user
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,alphanum,min=4,max=32"`
	Password string `json:"password" binding:"required,min=6,max=256"`
	FullName string `json:"full_name" binding:"required,min=4,max=256"`
	Email    string `json:"email" binding:"required,email"`
}

func (s *Server) CreateUser(ctx *gin.Context) (err error) {
	var req CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return &ApiError{Status: http.StatusBadRequest, Err: err.Error()}
	}

	// Hash the user password
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		return &ApiError{Status: http.StatusInternalServerError, Err: "error hashing password"}
	}

	// Create a new user
	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}
	user, err := s.store.CreateUser(ctx, arg)
	if err != nil {
		return &ApiError{Status: http.StatusInternalServerError, Err: "error creating user"}
	}

	// Return the user
	ctx.JSON(http.StatusOK, user)
	return
}
