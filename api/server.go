package api

import (
	db "bank/db/sqlc"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Server serves HTTP requests for banking service
type Server struct {
	store  db.Store
	router *gin.Engine
}

func NewServer(store db.Store) *Server {
	server := &Server{
		store: store,
	}
	router := gin.Default()

	// Add routes
	server.router = router

	// Custom validator engine
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	// Accounts
	router.POST("/users", makeGinHandlerFunc(server.createUser))
	router.POST("/accounts", makeGinHandlerFunc(server.createAccount))
	router.GET("/accounts/:id", makeGinHandlerFunc(server.getAccount))
	router.GET("/accounts", makeGinHandlerFunc(server.listAccounts))

	// Transfers
	router.POST("/transfer", makeGinHandlerFunc(server.createTransfer))
	return server
}

// The HandlerFuncs should return the error of this type embedding the http status code and error
type ApiError struct {
	Err    string `json:"error"`
	Status int    `json:"status"`
}

func (e ApiError) Error() string {
	return e.Err
}

type apiFunc func(*gin.Context) error

// Wraps the HandlerFuncs to handle the errors returned in one place
func makeGinHandlerFunc(f apiFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := f(ctx); err != nil {
			if e, ok := err.(*ApiError); ok { // check to see if the error is ok type ApiError with customer https satus otherwise respond with InternalServer error
				ctx.JSON(e.Status, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
