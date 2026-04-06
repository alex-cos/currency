package yahooapi

import (
	"encoding/csv"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alex-cos/currency"
	"github.com/alex-cos/restc"
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
	client *restc.Client
}

func New() currency.Currency {
	return NewWithClientTimeout(http.DefaultClient, restc.DefaultTimeout)
}

func NewWithTimeout(timeout time.Duration) currency.Currency {
	return NewWithClientTimeout(http.DefaultClient, timeout)
}

func NewWithClient(httpClient *http.Client) currency.Currency {
	return NewWithClientTimeout(httpClient, restc.DefaultTimeout)
}

func NewWithClientTimeout(
	httpClient *http.Client,
	timeout time.Duration,
) currency.Currency {
	client := restc.NewWithClient(APIURL+"/"+APIVERSION, httpClient)
	if timeout > 0 {
		client.SetTimeout(timeout)
	}
	client.SetRedirectPolicy(restc.NoRedirect)
	return &YahooAPI{client: client}
}

func (api *YahooAPI) Ping() error {
	return nil
}

func (api *YahooAPI) Latest(base string, symbols []string) (*currency.ResponseAPI, error) {
	end := time.Now().UTC().Truncate(DAY)
	start := end.AddDate(0, 0, -3)
	date := start.Format(DATEFORMAT)
	rates := map[string]map[string]float64{}
	for _, symbol := range symbols {
		endpoint := "finance/download/" + base + symbol + "=X"
		data, err := api.query(endpoint, map[string]string{
			"interval": "1d",
			"period1":  strconv.FormatInt(start.Unix(), 10),
			"period2":  strconv.FormatInt(end.Unix(), 10),
			"events":   "history",
		})
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
		if r, ok := rates[datetime.Format(DATEFORMAT)]; ok {
			rate = r
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
	date := datetime.Format(DATEFORMAT)
	rates := map[string]map[string]float64{}
	for _, symbol := range symbols {
		endpoint := "finance/download/" + base + symbol + "=X"
		data, err := api.query(endpoint, map[string]string{
			"interval": "1d",
			"period1":  strconv.FormatInt(datetime.AddDate(0, 0, -5).Unix(), 10),
			"period2":  strconv.FormatInt(datetime.Unix(), 10),
			"events":   "history",
		})
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
		if r := rates[day.Format(DATEFORMAT)]; len(r) > 0 {
			return &currency.ResponseAPI{
				Base:  base,
				Date:  date,
				Rates: r,
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
	resp := currency.HistoryResponseAPI{
		Base:  base,
		Date:  start.Format(DATEFORMAT),
		Rates: map[string]map[string]float64{},
	}
	for _, symbol := range symbols {
		endpoint := "finance/download/" + base + symbol + "=X"
		data, err := api.query(endpoint, map[string]string{
			"interval": "1d",
			"period1":  strconv.FormatInt(start.Unix(), 10),
			"period2":  strconv.FormatInt(end.Unix(), 10),
			"events":   "history",
		})
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
		if _, ok := (*rates)[date.Format(DATEFORMAT)]; !ok {
			(*rates)[date.Format(DATEFORMAT)] = map[string]float64{}
		}
		(*rates)[date.Format(DATEFORMAT)][symbol] = value
	}
	return nil
}

func (api *YahooAPI) query(endpoint string, params map[string]string) (string, error) {
	req := restc.Get(endpoint)
	if params != nil {
		req.SetQueryParams(params)
	}

	resp, err := api.client.Execute(req)
	if err != nil {
		return "", err
	}

	if len(resp.Bytes()) == 0 {
		return "", currency.ErrEmptyBody(restc.MethodGet, endpoint)
	}

	return resp.String(), nil
}
