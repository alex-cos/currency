// URL: https://exchangerate.host/#/docs

package exchangerates_test

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/alex-cos/currency/exchangerates"
	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	t.Parallel()

	api := exchangerates.NewWithTimeout(5 * time.Second)

	err := api.Ping()
	assert.NoError(t, err)
}

func TestLastest(t *testing.T) {
	t.Parallel()

	api := exchangerates.New()

	resp, err := api.Latest("USD", []string{"EUR", "GBP"})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
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

	api := exchangerates.New()

	resp, err := api.ForDate(time.Date(2019, time.June, 22, 0, 0, 0, 0, time.UTC), "USD", []string{"EUR", "GBP", "JPY"})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
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

	resp, err = api.ForDate(time.Date(2019, time.December, 25, 0, 0, 0, 0, time.UTC), "USD", []string{"EUR"})
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

	api := exchangerates.New()

	resp, err := api.History(
		time.Date(2018, time.February, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2018, time.June, 22, 0, 0, 0, 0, time.UTC),
		"USD", []string{"EUR", "GBP"})
	assert.NoError(t, err)
	if !testing.Short() {
		fmt.Printf("resp = %+v\n", resp)
	}
	assert.Len(t, resp.Rates, 142)
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

func TestHistory2(t *testing.T) {
	t.Parallel()

	api := exchangerates.New()

	resp, err := api.History(
		time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.January, 11, 0, 0, 0, 0, time.UTC),
		"USD", []string{"EUR", "GBP", "CHF", "CNY", "JPY"})
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Rates, "response should not be empty")

	dates := make([]string, len(resp.Rates))
	i := 0
	for k := range resp.Rates {
		dates[i] = k
		i++
	}
	if !testing.Short() {
		sort.Strings(dates)
		for _, date := range dates {
			rates := resp.Rates[date]
			for symbol, rate := range rates {
				fmt.Printf("%s\t%s 00:00:00\t%s\t%.6f\n", date, date, symbol, rate)
			}
			fmt.Printf("%s\t%s 00:00:00\t%s\t%.6f\n", date, date, "USD", 1.0)
		}
	}
}
