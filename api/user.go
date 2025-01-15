package api

import (
	db "bank/db/sqlc"
	"bank/util"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// Create a new file called api/user.go. This file will contain the implementation of the user service.
// CreateUser creates a new user
type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum,min=4,max=32"`
	Password string `json:"password" binding:"required,min=6,max=256"`
	FullName string `json:"full_name" binding:"required,min=4,max=256"`
	Email    string `json:"email" binding:"required,email"`
}

type userResponse struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username: user.Username,
		FullName: user.FullName,
		Email:    user.Email,
	}
}
func (s *Server) createUser(ctx *gin.Context) (err error) {

	var req createUserRequest

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
	rsp := newUserResponse(user)

	ctx.JSON(http.StatusOK, rsp)
	return
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

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum,min=4,max=32"`
	Password string `json:"password" binding:"required,min=6,max=256"`
}

type loginUserResponse struct {
	AccessToken string       `json:"access_token"`
	User        userResponse `json:"user"`
}

func (s *Server) loginUser(ctx *gin.Context) (err error) {
	var req loginUserRequest
	if err := ctx.ShouldBind(&req); err != nil {
		return &ApiError{Status: http.StatusBadRequest, Err: fmt.Sprintf("login failed: %s", err.Error())}
	}
	user, err := s.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return &ApiError{Status: http.StatusNotFound, Err: "login failed, username does not exist"}
		}
		return &ApiError{Status: http.StatusInternalServerError, Err: "login failed, failed to get user"}
	}
	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		return &ApiError{Status: http.StatusUnauthorized, Err: "login failed, not authorized"}
	}
	accessToken, err := s.tokenMaker.CreateToken(user.Username, s.config.TokenDuration)
	if err != nil {
		return &ApiError{Status: http.StatusInternalServerError, Err: "login failed, failed to generate token"}
	}
	rsp := loginUserResponse{
		AccessToken: accessToken,
		User:        newUserResponse(user),
	}
	ctx.JSON(http.StatusOK, rsp)
	return
}
