package ecb_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/alex-cos/currency/ecb"
	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	t.Parallel()

	api := ecb.New()

	err := api.Ping()
	assert.NoError(t, err)
}

func TestLastest(t *testing.T) {
	t.Parallel()

	api := ecb.New()

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
	assert.NotNil(t, resp)
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

	api := ecb.New()

	resp, err := api.ForDate(time.Date(2023, time.June, 22, 0, 0, 0, 0, time.UTC), "USD", []string{"EUR", "GBP", "JPY"})
	assert.NoError(t, err)
	if !testing.Short() {
		fmt.Printf("resp = %+v\n", resp)
	}
	assert.NotNil(t, resp)
	assert.Len(t, resp.Rates, 3)
	value, ok := resp.Rates["EUR"]
	assert.True(t, ok)
	assert.InDelta(t, 0.91033, value, 0.00001)
	value, ok = resp.Rates["GBP"]
	assert.True(t, ok)
	assert.InDelta(t, 0.78393, value, 0.00001)
	value, ok = resp.Rates["JPY"]
	assert.True(t, ok)
	assert.InDelta(t, 142.057350, value, 0.00001)

	resp, err = api.ForDate(time.Date(2023, time.January, 25, 0, 0, 0, 0, time.UTC), "USD", []string{"EUR"})
	assert.NoError(t, err)
	if !testing.Short() {
		fmt.Printf("resp = %+v\n", resp)
	}
	assert.NotNil(t, resp)
	assert.Len(t, resp.Rates, 1)
	value, ok = resp.Rates["EUR"]
	assert.True(t, ok)
	assert.InDelta(t, 0.91928, value, 0.00001)

	date := time.Date(2024, time.March, 31, 0, 0, 0, 0, time.UTC)
	resp, err = api.ForDate(date, "USD", []string{"USD", "EUR"})
	assert.NoError(t, err)
	assert.Len(t, resp.Rates, 2)
	value, ok = resp.Rates["EUR"]
	assert.True(t, ok)
	assert.InDelta(t, 0.92498, value, 0.00001)

	resp, err = api.ForDate(time.Now().UTC().AddDate(0, 0, -60), "USD", []string{"EUR", "GBP", "JPY"})
	assert.NoError(t, err)
	if !testing.Short() {
		fmt.Printf("resp = %+v\n", resp)
	}
	assert.NotNil(t, resp)
	assert.Len(t, resp.Rates, 3)
}

func TestHistory(t *testing.T) {
	t.Parallel()

	api := ecb.New()

	resp, err := api.History(
		time.Date(2023, time.February, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2023, time.June, 22, 0, 0, 0, 0, time.UTC),
		"USD", []string{"EUR", "GBP", "CHF", "CNY", "CAD", "JPY"})
	assert.NoError(t, err)
	if !testing.Short() {
		fmt.Printf("resp = %+v\n", resp)
	}
	assert.NotNil(t, resp.Rates)
	assert.Len(t, resp.Rates, 142)
	for _, rates := range resp.Rates {
		assert.Len(t, rates, 6)
		value, ok := rates["EUR"]
		assert.True(t, ok)
		assert.Greater(t, value, 0.0)
		value, ok = rates["GBP"]
		assert.True(t, ok)
		assert.Greater(t, value, 0.0)
	}
}
