// URL: https://currencyapi.com/docs

package currencyapi

import (
	"context"
	"encoding/json"
	"reflect"

	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alex-cos/currency"
)

const (
	// APIURL is the official currency API Endpoint.
	APIURL = "https://api.currencyapi.com"

	// APIVERSION the endpoint version.
	APIVERSION = "v3"
)

// CurrencyAPI represents a currency API Client connection.
type CurrencyAPI struct {
	client  *http.Client
	timeout time.Duration
	apikey  string
}

func New(apikey string) currency.Currency {
	return NewWithClient(apikey, http.DefaultClient)
}

func NewWithTimeout(apikey string, timeout time.Duration) currency.Currency {
	return NewWithClient(apikey, &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       timeout,
	})
}

func NewWithClient(apikey string, httpClient *http.Client) currency.Currency {
	return &CurrencyAPI{
		client:  httpClient,
		timeout: httpClient.Timeout,
		apikey:  apikey,
	}
}

func NewWithClientTimeout(
	apikey string,
	httpClient *http.Client,
	timeout time.Duration,
) currency.Currency {
	return &CurrencyAPI{
		client:  httpClient,
		timeout: timeout,
		apikey:  apikey,
	}
}

func (api *CurrencyAPI) Ping() error {
	_, err := api.query("status", nil, &StatusResponseAPI{})
	if err != nil {
		return err
	}

	return nil
}

func (api *CurrencyAPI) Latest(base string, symbols []string) (*currency.ResponseAPI, error) {
	params := url.Values{
		"base_currency": {base},
		"currencies":    {strings.Join(symbols, ",")},
	}
	r, err := api.query("latest", params, &RatesResponseAPI{})
	if err != nil {
		return nil, err
	}
	latest := (r.(*RatesResponseAPI)) //nolint: forcetypeassert
	resp := currency.ResponseAPI{
		Base:  base,
		Date:  latest.Meta.LastUpdatedAt,
		Rates: map[string]float64{},
	}
	for symbol, val := range latest.Data {
		ok, _ := IsInArray(symbol, symbols)
		if ok {
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
	params := url.Values{
		"date":          {datetime.Format("2006-01-02")},
		"base_currency": {base},
		"currencies":    {strings.Join(symbols, ",")},
	}
	r, err := api.query("historical", params, &RatesResponseAPI{})
	if err != nil {
		return nil, err
	}
	latest := (r.(*RatesResponseAPI))
	resp := currency.ResponseAPI{
		Base:  base,
		Date:  latest.Meta.LastUpdatedAt,
		Rates: map[string]float64{},
	}
	for symbol, val := range latest.Data {
		ok, _ := IsInArray(symbol, symbols)
		if ok {
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
	params := url.Values{
		"accuracy":       {"day"},
		"datetime_start": {FormatRFC3339(start)},
		"datetime_end":   {FormatRFC3339(end)},
		"base_currency":  {base},
		"currencies":     {strings.Join(symbols, ",")},
	}
	r, err := api.query("range", params, &HistoryResponseAPI{})
	if err != nil {
		return nil, err
	}
	history := (r.(*HistoryResponseAPI))
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
		_, ok := resp.Rates[dt]
		if !ok {
			resp.Rates[dt] = map[string]float64{}
		}
		for symbol, val := range data.Currencies {
			ok, _ := IsInArray(symbol, symbols)
			if ok {
				resp.Rates[dt][symbol] = val.Value
			}
		}
	}

	return &resp, nil
}

// Unexported functions

func (api *CurrencyAPI) query(endpoint string, values url.Values, typ interface{}) (interface{}, error) {
	uri := fmt.Sprintf("%s/%s/%s", APIURL, APIVERSION, endpoint)
	resp, err := api.doRequest(http.MethodGet, uri, values, typ)

	return resp, err
}

func (api *CurrencyAPI) doRequest(method, reqURL string, values url.Values, typ interface{}) (interface{}, error) {
	ctx, cancel := api.getContext()
	if cancel != nil {
		defer cancel()
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, currency.ErrNewRequest(err)
	}
	if method == http.MethodGet {
		req.URL.RawQuery = values.Encode()
	}

	req.Header.Set("Apikey", api.apikey)

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, currency.ErrDoRequest(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, currency.ErrReadBody(method, reqURL, err)
	}

	mimeType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, currency.ErrParseContentType(method, reqURL, err)
	}
	if mimeType != "application/json" {
		return nil, currency.ErrUnsupportedMimeType(method, reqURL, mimeType)
	}

	if len(body) == 0 {
		return nil, currency.ErrEmptyBody(method, reqURL)
	}

	var jsonResp interface{}

	err = json.Unmarshal(body, &jsonResp)
	if err != nil {
		return nil, currency.ErrUnmarshal(method, reqURL, err)
	}

	if typ != nil {
		jsonData := typ
		err = json.Unmarshal(body, &jsonData)
		if err != nil {
			return nil, currency.ErrUnmarshal(method, reqURL, err)
		}
		return jsonData, nil
	}

	return jsonResp, currency.ErrEmptyType(method, reqURL)
}

func (api *CurrencyAPI) getContext() (context.Context, context.CancelFunc) {
	var cancel context.CancelFunc

	ctx := context.Background()
	if api.timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), api.timeout)
	}

	return ctx, cancel
}

func FormatRFC3339(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

func ParseRFC3339(str string) (time.Time, error) {
	t, err := time.Parse("2006-01-02T15:04:05.000Z", str)
	if err != nil {
		err = fmt.Errorf("failed to parse datetime '%s': %w", str, err)
	}
	return t, err
}

func IsInArray(val, array interface{}) (bool, int) {
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice, reflect.Array:
		s := reflect.ValueOf(array)
		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) {
				return true, i
			}
		}
	default:
		if reflect.DeepEqual(val, array) {
			return true, 0
		}
	}
	return false, -1
}
