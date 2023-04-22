package api

import (
	db "bank/db/sqlc"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type createTransferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,oneof=USD EUR INR CAD"`
}

func (server *Server) createTransfer(ctx *gin.Context) (err error) {
	var req createTransferRequest
	if err = ctx.ShouldBind(&req); err != nil {
		return &ApiError{Status: http.StatusBadRequest, Err: err.Error()}
	}

	validFromCh := make(chan error)
	validToCh := make(chan error)

	go server.validAccount(ctx, req.FromAccountID, req.Currency, validFromCh)
	go server.validAccount(ctx, req.ToAccountID, req.Currency, validToCh)

	validFrom, validTo := <-validFromCh, <-validToCh

	if validFrom != nil {
		return validFrom
	}

	if validTo != nil {
		return validTo
	}

	arg := db.TransferTxParms{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	transferId, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		return &ApiError{Status: http.StatusInternalServerError, Err: err.Error()}
	}

	ctx.JSON(http.StatusOK, gin.H{"transferID": transferId})
	return
}

func (server *Server) validAccount(ctx *gin.Context, accountId int64, currency string, valid chan<- error) {
	account, err := server.store.GetAccount(ctx, accountId)
	if err != nil {
		if err == sql.ErrNoRows {
			valid <- &ApiError{Status: http.StatusNotFound, Err: fmt.Sprintf("Account ID: %d not found", accountId)}
			return
		}
		valid <- &ApiError{Status: http.StatusInternalServerError, Err: err.Error()}
		return
	}
	if account.Currency != currency {
		valid <- &ApiError{Status: http.StatusBadRequest, Err: fmt.Sprintf("Currency %s not supported on account ID %d", currency, accountId)}
		return
	}
	valid <- err
}
