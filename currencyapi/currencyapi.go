// URL: https://currencyapi.com/docs

package currencyapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/alex-cos/currency"
	"github.com/alex-cos/restc"
)

const (
	// APIURL is the official currency API Endpoint.
	APIURL = "https://api.currencyapi.com"

	// APIVERSION the endpoint version.
	APIVERSION = "v3"
)

// CurrencyAPI represents a currency API Client connection.
type CurrencyAPI struct {
	client *restc.Client
}

func New(apikey string) currency.Currency {
	return NewWithClientTimeout(apikey, http.DefaultClient, restc.DefaultTimeout)
}

func NewWithTimeout(apikey string, timeout time.Duration) currency.Currency {
	return NewWithClientTimeout(apikey, http.DefaultClient, timeout)
}

func NewWithClient(apikey string, httpClient *http.Client) currency.Currency {
	return NewWithClientTimeout(apikey, httpClient, restc.DefaultTimeout)
}

func NewWithClientTimeout(
	apikey string,
	httpClient *http.Client,
	timeout time.Duration,
) currency.Currency {
	client := restc.NewWithClient(APIURL+"/"+APIVERSION, httpClient)
	client.SetHeader("apikey", apikey)
	if timeout > 0 {
		client.SetTimeout(timeout)
	}
	return &CurrencyAPI{
		client: client,
	}
}

func (api *CurrencyAPI) Ping() error {
	_, err := api.query("status", nil, &StatusResponseAPI{})
	return err
}

func (api *CurrencyAPI) Latest(base string, symbols []string) (*currency.ResponseAPI, error) {
	r, err := api.query("latest", map[string]string{
		"base_currency": base,
		"currencies":    strings.Join(symbols, ","),
	}, &RatesResponseAPI{})
	if err != nil {
		return nil, err
	}
	latest := r.(*RatesResponseAPI)
	resp := currency.ResponseAPI{
		Base:  base,
		Date:  latest.Meta.LastUpdatedAt,
		Rates: map[string]float64{},
	}
	for symbol, val := range latest.Data {
		if isInArray(symbol, symbols) {
			resp.Rates[symbol] = val.Value
		}
	}

	return &resp, nil
}

func (api *CurrencyAPI) ForDate(
	datetime time.Time,
	base string,
	symbols []string,
) (*currency.ResponseAPI, error) {
	r, err := api.query("historical", map[string]string{
		"date":          datetime.Format("2006-01-02"),
		"base_currency": base,
		"currencies":    strings.Join(symbols, ","),
	}, &RatesResponseAPI{})
	if err != nil {
		return nil, err
	}
	latest := r.(*RatesResponseAPI)
	resp := currency.ResponseAPI{
		Base:  base,
		Date:  latest.Meta.LastUpdatedAt,
		Rates: map[string]float64{},
	}
	for symbol, val := range latest.Data {
		if isInArray(symbol, symbols) {
			resp.Rates[symbol] = val.Value
		}
	}

	return &resp, nil
}

func (api *CurrencyAPI) History(
	start,
	end time.Time,
	base string,
	symbols []string,
) (*currency.HistoryResponseAPI, error) {
	r, err := api.query("range", map[string]string{
		"accuracy":       "day",
		"datetime_start": FormatRFC3339(start),
		"datetime_end":   FormatRFC3339(end),
		"base_currency":  base,
		"currencies":     strings.Join(symbols, ","),
	}, &HistoryResponseAPI{})
	if err != nil {
		return nil, err
	}
	history := r.(*HistoryResponseAPI)
	resp := currency.HistoryResponseAPI{
		Base:  base,
		Date:  start.Format("2006-01-02"),
		Rates: map[string]map[string]float64{},
	}
	for _, data := range history.Data {
		datetime, err := ParseRFC3339(data.Datetime)
		if err != nil {
			continue
		}
		dt := datetime.Format("2006-01-02")
		if _, ok := resp.Rates[dt]; !ok {
			resp.Rates[dt] = map[string]float64{}
		}
		for symbol, val := range data.Currencies {
			if isInArray(symbol, symbols) {
				resp.Rates[dt][symbol] = val.Value
			}
		}
	}

	return &resp, nil
}

// Unexported functions

func (api *CurrencyAPI) query(endpoint string, params map[string]string, typ interface{}) (interface{}, error) {
	req := restc.Get(endpoint).
		SetResponseType(typ)
	if params != nil {
		req.SetQueryParams(params)
	}

	resp, err := api.client.Execute(req)
	if err != nil {
		return nil, err
	}

	return resp.Content(), nil
}

func FormatRFC3339(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

func ParseRFC3339(str string) (time.Time, error) {
	t, err := time.Parse("2006-01-02T15:04:05.000Z", str)
	if err != nil {
		err = currency.ErrGeneric("failed to parse datetime '" + str + "': " + err.Error())
	}
	return t, err
}

func isInArray(val string, array []string) bool {
	for _, item := range array {
		if val == item {
			return true
		}
	}
	return false
}
