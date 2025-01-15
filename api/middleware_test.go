package api

import (
	"bank/token"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type mockTokenMaker struct {
    validToken string
}

func (m *mockTokenMaker) CreateToken(username string, duration time.Duration) (string, error) {
    return m.validToken, nil
}

func (m *mockTokenMaker) VerifyToken(tokenString string) (*token.Payload, error) {
    if tokenString == m.validToken {
        return &token.Payload{}, nil
    }
    return nil, fmt.Errorf("invalid token")
}

func TestAuthMiddleware(t *testing.T) {
    gin.SetMode(gin.TestMode)

    tests := []struct {
        name          string
        setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
        checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
    }{
        {
            name: "MissingAuthorizationHeader",
            setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
                // No authorization header
            },
            checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
                require.Equal(t, http.StatusUnauthorized, recorder.Code)
            },
        },
        {
            name: "InvalidAuthorizationHeaderFormat",
            setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
                request.Header.Set("Authorization", "InvalidFormat")
            },
            checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
                require.Equal(t, http.StatusUnauthorized, recorder.Code)
            },
        },
        {
            name: "InvalidToken",
            setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
                request.Header.Set("Authorization", "Bearer invalid-token")
            },
            checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
                require.Equal(t, http.StatusUnauthorized, recorder.Code)
            },
        },
        {
            name: "ValidToken",
            setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
                request.Header.Set("Authorization", "Bearer valid-token")
            },
            checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
                require.Equal(t, http.StatusOK, recorder.Code)
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tokenMaker := &mockTokenMaker{validToken: "valid-token"}

            router := gin.New()
            router.Use(authMiddleware(tokenMaker))
            router.GET("/", func(ctx *gin.Context) {
                ctx.JSON(http.StatusOK, gin.H{"message": "success"})
            })

            recorder := httptest.NewRecorder()
            request, err := http.NewRequest(http.MethodGet, "/", nil)
            require.NoError(t, err)

            tt.setupAuth(t, request, tokenMaker)
            router.ServeHTTP(recorder, request)
            tt.checkResponse(t, recorder)
        })
    }
}