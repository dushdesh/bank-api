package api

import (
	db "bank/db/sqlc"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type createAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,oneof=USD EUR CAD INR"`
}

func (server *Server) createAccount(ctx *gin.Context) (err error) {
	var req createAccountRequest
	if err = ctx.ShouldBind(&req); err != nil {
		return &ApiError{Status: http.StatusBadRequest, Err: err.Error()}
	}

	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: req.Currency,
		Balance:  0,
	}

	accountId, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return &ApiError{Status: http.StatusConflict, Err: pqErr.Error()}
			case "foreign_key_violation":
				return &ApiError{Status: http.StatusNotFound, Err: pqErr.Error()}
			}
		}
		return &ApiError{Status: http.StatusInternalServerError, Err: err.Error()}
	}

	ctx.JSON(http.StatusOK, gin.H{"accountID": accountId})
	return
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAccount(ctx *gin.Context) (err error) {
	var req getAccountRequest
	if err = ctx.ShouldBindUri(&req); err != nil {
		return &ApiError{Status: http.StatusBadRequest, Err: err.Error()}
	}

	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &ApiError{Status: http.StatusNotFound, Err: err.Error()}
		}
		return &ApiError{Status: http.StatusInternalServerError, Err: err.Error()}
	}

	ctx.JSON(http.StatusOK, account)
	return
}

type listAccountsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=1"`
}

func (server *Server) listAccounts(ctx *gin.Context) (err error) {
	var req listAccountsRequest
	if err = ctx.ShouldBindQuery(&req); err != nil {
		return &ApiError{Status: http.StatusBadRequest, Err: err.Error()}
	}

	arg := db.ListAccountsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	accounts, err := server.store.ListAccounts(ctx, arg)
	if err != nil {
		return &ApiError{Status: http.StatusInternalServerError, Err: err.Error()}
	}

	ctx.JSON(http.StatusOK, accounts)
	return
}
