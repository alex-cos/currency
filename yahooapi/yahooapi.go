package yahooapi

import (
	"context"
	"encoding/csv"
	"errors"

	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/alex-cos/currency"
)

const (
	// APIURL is the official yahoo API Endpoint.
	APIURL = "https://query1.finance.yahoo.com"

	// APIVERSION the endpoint version.
	APIVERSION = "v7"

	DATEFORMAT = "2006-01-02"

	DAY = 24 * time.Hour
)

// YahooAPI represents a yahoo API Client connection.
type YahooAPI struct {
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
	return &YahooAPI{
		client:  httpClient,
		timeout: httpClient.Timeout,
	}
}

func NewWithClientTimeout(
	httpClient *http.Client,
	timeout time.Duration,
) currency.Currency {
	return &YahooAPI{
		client:  httpClient,
		timeout: timeout,
	}
}

func (api *YahooAPI) Ping() error {
	return nil
}

func (api *YahooAPI) Latest(base string, symbols []string) (*currency.ResponseAPI, error) {
	end := time.Now().UTC().Truncate(DAY)
	start := end.AddDate(0, 0, -3)
	params := url.Values{
		"interval": {"1d"},
		"period1":  {strconv.FormatInt(start.Unix(), 10)},
		"period2":  {strconv.FormatInt(end.Unix(), 10)},
		"events":   {"history"},
	}
	date := start.Format(DATEFORMAT)
	rates := map[string]map[string]float64{}
	for _, symbol := range symbols {
		endpoint := fmt.Sprintf("finance/download/%s%s=X", base, symbol)
		data, err := api.query(endpoint, params)
		if err != nil {
			return nil, err
		}
		err = api.parseData(data, symbol, &rates)
		if err != nil {
			return nil, err
		}
	}
	rate := map[string]float64{}
	datetime := end
	for datetime.Unix() >= start.Unix() {
		_, ok := rates[datetime.Format(DATEFORMAT)]
		if ok {
			rate = rates[datetime.Format(DATEFORMAT)]
			break
		}
		datetime = datetime.AddDate(0, 0, -1)
	}

	return &currency.ResponseAPI{
		Base:  base,
		Date:  date,
		Rates: rate,
	}, nil
}

func (api *YahooAPI) ForDate(datetime time.Time, base string, symbols []string) (*currency.ResponseAPI, error) {
	params := url.Values{
		"interval": {"1d"},
		"period1":  {strconv.FormatInt(datetime.AddDate(0, 0, -5).Unix(), 10)},
		"period2":  {strconv.FormatInt(datetime.Unix(), 10)},
		"events":   {"history"},
	}
	date := datetime.Format(DATEFORMAT)
	rates := map[string]map[string]float64{}
	for _, symbol := range symbols {
		endpoint := fmt.Sprintf("finance/download/%s%s=X", base, symbol)
		data, err := api.query(endpoint, params)
		if err != nil {
			return nil, err
		}
		err = api.parseData(data, symbol, &rates)
		if err != nil {
			return nil, err
		}
	}
	day := datetime
	for day.Unix() > datetime.AddDate(0, 0, -3).Unix() {
		rates := rates[day.Format(DATEFORMAT)]
		if len(rates) > 0 {
			return &currency.ResponseAPI{
				Base:  base,
				Date:  date,
				Rates: rates,
			}, nil
		}
		day = day.AddDate(0, 0, -1)
	}

	return &currency.ResponseAPI{
		Base:  base,
		Date:  date,
		Rates: map[string]float64{},
	}, nil
}

func (api *YahooAPI) History(
	start time.Time,
	end time.Time,
	base string,
	symbols []string,
) (*currency.HistoryResponseAPI, error) {
	params := url.Values{
		"interval": {"1d"},
		"period1":  {strconv.FormatInt(start.Unix(), 10)},
		"period2":  {strconv.FormatInt(end.Unix(), 10)},
		"events":   {"history"},
	}
	resp := currency.HistoryResponseAPI{
		Base:  base,
		Date:  start.Format(DATEFORMAT),
		Rates: map[string]map[string]float64{},
	}
	for _, symbol := range symbols {
		endpoint := fmt.Sprintf("finance/download/%s%s=X", base, symbol)
		data, err := api.query(endpoint, params)
		if err != nil {
			return nil, err
		}
		err = api.parseData(data, symbol, &resp.Rates)
		if err != nil {
			return nil, err
		}
	}
	return &resp, nil
}

// Unexported functions

func (api *YahooAPI) parseData(data, symbol string, rates *map[string]map[string]float64) error {
	csvReader := csv.NewReader(strings.NewReader(data))
	for {
		rec, err := csvReader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if len(rec) < 5 {
			continue
		}
		date, err := time.Parse("2006-01-02", rec[0])
		if err != nil {
			continue
		}
		value, err := strconv.ParseFloat(rec[4], 64)
		if err != nil {
			continue
		}
		_, ok := (*rates)[date.Format(DATEFORMAT)]
		if !ok {
			(*rates)[date.Format(DATEFORMAT)] = map[string]float64{}
		}
		(*rates)[date.Format(DATEFORMAT)][symbol] = value
	}
	return nil
}

func (api *YahooAPI) query(endpoint string, values url.Values) (string, error) {
	uri := fmt.Sprintf("%s/%s/%s", APIURL, APIVERSION, endpoint)
	resp, err := api.doRequest(http.MethodGet, uri, values)

	return resp, err
}

func (api *YahooAPI) doRequest(method, reqURL string, values url.Values) (string, error) {
	ctx, cancel := api.getContext()
	if cancel != nil {
		defer cancel()
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return "", currency.ErrNewRequest(err)
	}
	req.URL.RawQuery = values.Encode()

	resp, err := api.client.Do(req)
	if err != nil {
		return "", currency.ErrDoRequest(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", currency.ErrReadBody(method, reqURL, err)
	}

	mimeType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return "", currency.ErrParseContentType(method, reqURL, err)
	}
	if mimeType != "text/csv" {
		return "", currency.ErrUnsupportedMimeType(method, reqURL, mimeType)
	}

	if len(body) == 0 {
		return "", currency.ErrEmptyBody(method, reqURL)
	}

	return string(body), nil
}

func (api *YahooAPI) getContext() (context.Context, context.CancelFunc) {
	var cancel context.CancelFunc

	ctx := context.Background()
	if api.timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), api.timeout)
	}

	return ctx, cancel
}
