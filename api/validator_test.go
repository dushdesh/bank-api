// FILEPATH: /Users/dushdesh/Documents/code/github.com/dushdesh/bank-api/api/validator_test.go

package api

import (
    "testing"

    "github.com/go-playground/validator/v10"
    "github.com/stretchr/testify/require"
    "bank/util"
)

type CurrencyTest struct {
    Currency string `validate:"currency"`
}

func TestValidCurrency(t *testing.T) {
    validate := validator.New()
    validate.RegisterValidation("currency", validCurrency)

    tests := []struct {
        name     string
        currency string
        valid    bool
    }{
        {
            name:     "ValidCurrencyUSD",
            currency: util.USD,
            valid:    true,
        },
        {
            name:     "ValidCurrencyEUR",
            currency: util.EUR,
            valid:    true,
        },
        {
            name:     "InvalidCurrency",
            currency: "INVALID",
            valid:    false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            testStruct := CurrencyTest{Currency: tt.currency}
            err := validate.Struct(testStruct)
            if tt.valid {
                require.NoError(t, err)
            } else {
                require.Error(t, err)
            }
        })
    }
}