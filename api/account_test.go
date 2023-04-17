package api

import (
	mockdb "bank/db/mock"
	db "bank/db/sqlc"
	"bank/util"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestAccountApi(t *testing.T) {
	account := randomAccount()

	ctrl := gomock.NewController(t)
	ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	// Build stubs
	store.EXPECT().
		GetAccount(gomock.Any(), gomock.Eq(account.ID)).
		Times(1).
		Return(account, nil)

	// Start a test server and send request
	server := NewServer(store)
	recorder := httptest.NewRecorder()

	// Generate request
	url := fmt.Sprintf("/accounts/%d", account.ID)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	// Setup the response recorder to collect API responses
	server.router.ServeHTTP(recorder, request)

	// Check response http status code
	require.Equal(t, http.StatusOK, recorder.Code)
}

func randomAccount() db.Account {
	return db.Account{
		ID: util.RandomInt(1, 1000),
		Owner: util.RandomOwner(),
		Currency: util.RandomCurrency(),
		Balance: util.RandomAmount(),
	}
}
