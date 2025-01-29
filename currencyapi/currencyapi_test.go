package currencyapi_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/alex-cos/currency/currencyapi"
	"github.com/stretchr/testify/assert"
)

var APIKey = os.Getenv("TEST_CURRENCY_API_KEY")

func TestPing(t *testing.T) {
	t.Parallel()

	api := currencyapi.New(APIKey)

	err := api.Ping()
	assert.NoError(t, err)
}

func TestLastest(t *testing.T) {
	t.Parallel()

	api := currencyapi.New(APIKey)

	resp, err := api.Latest("USD", []string{"EUR"})
	assert.NoError(t, err)
	if !testing.Short() {
		fmt.Printf("resp = %+v\n", resp)
	}

	resp, err = api.Latest("USD", []string{"EUR", "GBP"})
	assert.NoError(t, err)
	if !testing.Short() {
		fmt.Printf("resp = %+v\n", resp)
	}
	assert.Len(t, resp.Rates, 2)
	value, ok := resp.Rates["EUR"]
	assert.True(t, ok)
	assert.Greater(t, value, 0.0)
	value, ok = resp.Rates["GBP"]
	assert.True(t, ok)
	assert.Greater(t, value, 0.0)
}

func TestForDate(t *testing.T) {
	t.Parallel()

	api := currencyapi.New(APIKey)

	resp, err := api.ForDate(time.Date(2023, time.June, 22, 0, 0, 0, 0, time.UTC), "USD", []string{"EUR", "GBP", "JPY"})
	assert.NoError(t, err)
	if !testing.Short() {
		fmt.Printf("resp = %+v\n", resp)
	}
	assert.Len(t, resp.Rates, 3)
	value, ok := resp.Rates["EUR"]
	assert.True(t, ok)
	assert.Greater(t, value, 0.0)
	value, ok = resp.Rates["GBP"]
	assert.True(t, ok)
	assert.Greater(t, value, 0.0)
	value, ok = resp.Rates["JPY"]
	assert.True(t, ok)
	assert.Greater(t, value, 0.0)

	resp, err = api.ForDate(time.Date(2022, time.December, 25, 0, 0, 0, 0, time.UTC), "USD", []string{"EUR"})
	assert.NoError(t, err)
	if !testing.Short() {
		fmt.Printf("resp = %+v\n", resp)
	}
	assert.Len(t, resp.Rates, 1)
	value, ok = resp.Rates["EUR"]
	assert.True(t, ok)
	assert.Greater(t, value, 0.0)
}

func TestHistory(t *testing.T) {
	t.Parallel()

	api := currencyapi.New(APIKey)

	resp, err := api.History(
		time.Date(2023, time.February, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2023, time.June, 22, 0, 0, 0, 0, time.UTC),
		"USD", []string{"EUR", "GBP"})
	assert.NoError(t, err)
	if !testing.Short() {
		fmt.Printf("resp = %+v\n", resp)
	}
	assert.Len(t, resp.Rates, 102)
	for _, rates := range resp.Rates {
		assert.Len(t, rates, 2)
		value, ok := rates["EUR"]
		assert.True(t, ok)
		assert.Greater(t, value, 0.0)
		value, ok = rates["GBP"]
		assert.True(t, ok)
		assert.Greater(t, value, 0.0)
	}
}
