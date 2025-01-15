package api

import (
	db "bank/db/sqlc"
	"bank/token"
	"bank/util"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Server serves HTTP requests for banking service
type Server struct {
	store      db.Store
	tokenMaker token.Maker
	config     util.Config
	router     *gin.Engine
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TokenKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		store:      store,
		tokenMaker: tokenMaker,
		config:     config,
	}

	// Add routes
	server.setupRouter()

	// Custom validator engine
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}
	
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()
	
	
	// Public routes
    router.POST("/users", makeGinHandlerFunc(server.createUser))
    router.POST("/login", makeGinHandlerFunc(server.loginUser))

    // Protected routes
    protected := router.Group("/")
    protected.Use(authMiddleware(server.tokenMaker))
    {
        // Accounts
        protected.POST("/accounts", makeGinHandlerFunc(server.createAccount))
        protected.GET("/accounts/:id", makeGinHandlerFunc(server.getAccount))
        protected.GET("/accounts", makeGinHandlerFunc(server.listAccounts))

        // Transfers
        protected.POST("/transfer", makeGinHandlerFunc(server.createTransfer))
    }
	
	server.router = router
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
