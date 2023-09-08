package api

import (
	db "bank/db/sqlc"
	"bank/util"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// Create a new file called api/user.go. This file will contain the implementation of the user service.
// CreateUser creates a new user
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,alphanum,min=4,max=32"`
	Password string `json:"password" binding:"required,min=6,max=256"`
	FullName string `json:"full_name" binding:"required,min=4,max=256"`
	Email    string `json:"email" binding:"required,email"`
}

type UserResponse struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

func (s *Server) createUser(ctx *gin.Context) (err error) {

	fmt.Println("In Create User")
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
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return &ApiError{Status: http.StatusConflict, Err: err.Error()}
			}
		}
		return &ApiError{Status: http.StatusInternalServerError, Err: "error creating user"}
	}

	// Return the user
	rsp := UserResponse{	// Create a new UserResponse struct
		Username: user.Username,
		FullName: user.FullName,
		Email:    user.Email,
	}
	
	ctx.JSON(http.StatusOK, rsp)
	return
}

type LoginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum,min=4,max=32"`
	Password string `json:"password" binding:"required,min=6,max=256"`
}

// GetUser returns the given user
func (s *Server) GetUser(ctx *gin.Context) (err error) {
	username := ctx.Param("username")

	user, err := s.store.GetUser(ctx, username)
	if err != nil {
		return &ApiError{Status: http.StatusInternalServerError, Err: "error getting user"}
	}

	ctx.JSON(http.StatusOK, user)
	return
}