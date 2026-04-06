# Currency

A Go library providing a unified interface to multiple exchange rate APIs.

## Features

- **Single interface** for 4 exchange rate providers
- **Latest rates**, **historical rates**, and **date-specific rates**
- **Built-in mock** for testing your applications
- **Configurable HTTP client** and timeouts

## Supported Providers

| Provider | Package | Auth | Format |
| ---------- | --------- | ------ | -------- |
| [European Central Bank](https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml) | `ecb` | None | XML |
| [Yahoo Finance](https://finance.yahoo.com/) | `yahooapi` | None (deprecated) | CSV |
| [CurrencyAPI](https://currencyapi.com/docs) | `currencyapi` | API Key | JSON |
| [ExchangeRate.host](https://exchangerate.host/#/docs) | `exchangerates` | None | JSON |

## Installation

```bash
go get github.com/alex-cos/currency
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/alex-cos/currency/ecb"
)

func main() {
    api := ecb.New()

    // Get latest rates
    resp, err := api.Latest("USD", []string{"EUR", "GBP", "JPY"})
    if err != nil {
        panic(err)
    }
    fmt.Printf("1 USD = %.4f EUR\n", resp.Rates["EUR"])
}
```

## Usage

### European Central Bank (ECB)

No API key required. Base currency is always EUR on the ECB side — the library handles cross-conversion.

```go
import "github.com/alex-cos/currency/ecb"

api := ecb.New()
// With custom timeout
api := ecb.NewWithTimeout(10 * time.Second)
// With custom HTTP client
api := ecb.NewWithClient(&http.Client{...})
```

### Yahoo Finance

> **Deprecated / Broken** — Yahoo Finance now requires authentication (cookies + CRUMB token) for its download endpoint.
> This provider will fail with HTTP 401 errors unless credentials are provided. Consider using another provider.

No API key required. Fetches historical CSV data from Yahoo Finance.

```go
import "github.com/alex-cos/currency/yahooapi"

api := yahooapi.New()
```

### CurrencyAPI

Requires an API key from [currencyapi.com](https://currencyapi.com/).

```go
import "github.com/alex-cos/currency/currencyapi"

api := currencyapi.New("your-api-key")
// With custom timeout
api := currencyapi.NewWithTimeout("your-api-key", 10*time.Second)
// With custom HTTP client
api := currencyapi.NewWithClient("your-api-key", &http.Client{...})
```

### ExchangeRate

No API key required. Uses ECB as data source.

```go
import "github.com/alex-cos/currency/exchangerates"

api := exchangerates.New()
// With custom timeout
api := exchangerates.NewWithTimeout(10 * time.Second)
```

## API

All providers implement the `currency.Currency` interface:

```go
type Currency interface {
    // Ping performs a simple request to check if the remote service is up.
    Ping() error

    // Latest returns the latest conversion rates for the given symbols.
    Latest(base string, symbols []string) (*ResponseAPI, error)

    // ForDate returns the conversion rates for the specified datetime.
    ForDate(datetime time.Time, base string, symbols []string) (*ResponseAPI, error)

    // History returns the conversion rates for the given symbols between start and end time.
    History(start time.Time, end time.Time, base string, symbols []string) (*HistoryResponseAPI, error)
}
```

### Response Types

```go
type ResponseAPI struct {
    Base  string             `json:"base"`
    Date  string             `json:"date"`
    Rates map[string]float64 `json:"rates"`
}

type HistoryResponseAPI struct {
    Base  string                        `json:"base"`
    Date  string                        `json:"date"`
    Rates map[string]map[string]float64 `json:"rates"`
}
```

## Testing

### Run all tests

```bash
go test ./...
```

### Run tests with verbose output

```bash
go test -v ./...
```

### Run short tests (skip verbose output)

```bash
go test -short ./...
```

### CurrencyAPI tests

Requires setting the `TEST_CURRENCY_API_KEY` environment variable:

```bash
TEST_CURRENCY_API_KEY=your-key go test ./currencyapi/...
```

## Using the Mock

The library includes a mock implementation for testing your own code:

```go
import "github.com/alex-cos/currency"

func TestMyService(t *testing.T) {
    mock := currency.NewMock()

    // Use mock in your service
    svc := NewService(mock)

    // ... your test assertions
}
```

## Requirements

- Go 1.25+
