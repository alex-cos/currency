// URL: https://exchangerate.host/#/docs

package exchangerates

import (
	"context"
	"encoding/json"

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
	// APIURL is the official exchangerates API Endpoint.
	APIURL = "https://api.exchangerate.host"
)

// ExchangeRatesAPI represents a exchangerates API Client connection.
type ExchangeRatesAPI struct {
	client  *http.Client
	timeout time.Duration
}

func New() currency.Currency {
	return NewWithClient(http.DefaultClient)
}

func NewWithTimeout(timeout time.Duration) currency.Currency {
	return NewWithClient(&http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       timeout,
	})
}

func NewWithClient(httpClient *http.Client) currency.Currency {
	return &ExchangeRatesAPI{
		client:  httpClient,
		timeout: httpClient.Timeout,
	}
}

func NewWithClientTimeout(
	httpClient *http.Client,
	timeout time.Duration,
) currency.Currency {
	return &ExchangeRatesAPI{
		client:  httpClient,
		timeout: timeout,
	}
}

func (api *ExchangeRatesAPI) Ping() error {
	_, err := api.query("latest", nil, &ResponseAPI{})
	if err != nil {
		return err
	}

	return nil
}

func (api *ExchangeRatesAPI) Latest(base string, symbols []string) (*currency.ResponseAPI, error) {
	params := url.Values{
		"base":    {base},
		"symbols": {strings.Join(symbols, ",")},
		"format":  {"json"},
		"source":  {"ecb"},
	}
	resp, err := api.query("latest", params, &currency.ResponseAPI{})
	if err != nil {
		return nil, err
	}

	return resp.(*currency.ResponseAPI), nil
}

func (api *ExchangeRatesAPI) ForDate(
	datetime time.Time,
	base string,
	symbols []string,
) (*currency.ResponseAPI, error) {
	params := url.Values{
		"base":    {base},
		"symbols": {strings.Join(symbols, ",")},
		"format":  {"json"},
		"source":  {"ecb"},
	}
	resp, err := api.query(datetime.Format("2006-01-02"), params, &ResponseAPI{})
	if err != nil {
		return nil, err
	}

	return resp.(*currency.ResponseAPI), nil
}

func (api *ExchangeRatesAPI) History(
	start,
	end time.Time,
	base string,
	symbols []string,
) (*currency.HistoryResponseAPI, error) {
	params := url.Values{
		"base":       {base},
		"symbols":    {strings.Join(symbols, ",")},
		"start_date": {start.Format("2006-01-02")},
		"end_date":   {end.Format("2006-01-02")},
		"format":     {"json"},
		"source":     {"ecb"},
	}
	resp, err := api.query("timeseries", params, &HistoryResponseAPI{})
	if err != nil {
		return nil, err
	}

	return resp.(*currency.HistoryResponseAPI), nil
}

// Unexported functions

func (api *ExchangeRatesAPI) query(endpoint string, values url.Values, typ interface{}) (interface{}, error) {
	uri := fmt.Sprintf("%s/%s", APIURL, endpoint)
	resp, err := api.doRequest(http.MethodGet, uri, values, nil, typ)

	return resp, err
}

func (api *ExchangeRatesAPI) doRequest(method, reqURL string, values url.Values,
	headers map[string]string, typ interface{}) (interface{}, error) {
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

	for key, value := range headers {
		req.Header.Add(key, value)
	}

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

	err = parseJSONError(jsonResp, method, reqURL)
	if err != nil {
		return nil, err
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

func (api *ExchangeRatesAPI) getContext() (context.Context, context.CancelFunc) {
	var cancel context.CancelFunc

	ctx := context.Background()
	if api.timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), api.timeout)
	}

	return ctx, cancel
}

func parseJSONError(jsonResp interface{}, method, reqURL string) error {
	respMap, ok := jsonResp.(map[string]interface{})
	if !ok {
		return nil
	}
	if msg, ok := respMap["error"]; ok {
		if message, ok := msg.(string); ok {
			return currency.ErrGeneric(message)
		}
		return currency.ErrUnexpectedError(method, reqURL)
	}
	if msg, ok := respMap["success"]; ok {
		if success, ok := msg.(bool); ok && !success {
			return currency.ErrFailedNotification(method, reqURL)
		}
	}

	return nil
}
