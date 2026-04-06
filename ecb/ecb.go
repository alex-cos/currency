package ecb

import (
	"bytes"
	"encoding/xml"
	"errors"
	"strings"

	"fmt"
	"net/http"
	"time"

	"github.com/alex-cos/currency"
	"github.com/alex-cos/restc"
)

const (
	// APIURL is the official European Central Bank API Endpoint.
	APIURL = "https://www.ecb.europa.eu"

	DATEFORMAT = "2006-01-02"
)

// ECBAPI represents a ECB API Client connection.
type ECBAPI struct {
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
	client := restc.NewWithClient(APIURL, httpClient)
	if timeout > 0 {
		client.SetTimeout(timeout)
	}
	client.SetRedirectPolicy(restc.NoRedirect)
	client.SetRetryCount(1)
	return &ECBAPI{client: client}
}

func (api *ECBAPI) Ping() error {
	return nil
}

func (api *ECBAPI) Latest(base string, symbols []string) (*currency.ResponseAPI, error) {
	var response ECBResponse

	req := restc.Get("stats/eurofxref/eurofxref-daily.xml").
		SetHeader("Accept", restc.TypeApplicationXML)
	resp, err := api.client.Execute(req)
	if err != nil {
		return nil, err
	}

	if err := xml.NewDecoder(bytes.NewReader(resp.Bytes())).Decode(&response); err != nil {
		return nil, err
	}

	if len(response.Days) < 1 {
		return nil, errors.New("invalid response length")
	}

	date := response.Days[0].Date
	rates, err := parseRates(base, symbols, response.Days[0].Currencies)
	if err != nil {
		return nil, err
	}

	return &currency.ResponseAPI{
		Base:  base,
		Date:  date,
		Rates: rates,
	}, nil
}

func (api *ECBAPI) ForDate(datetime time.Time, base string, symbols []string) (*currency.ResponseAPI, error) {
	var (
		response ECBResponse
		endpoint = "stats/eurofxref/eurofxref-hist.xml"
	)

	if datetime.UTC().Truncate(24*time.Hour).Unix() >= time.Now().AddDate(0, 0, -88).Unix() {
		endpoint = "stats/eurofxref/eurofxref-hist-90d.xml"
	}

	req := restc.Get(endpoint).
		SetHeader("Accept", restc.TypeApplicationXML)

	resp, err := api.client.Execute(req)
	if err != nil {
		return nil, err
	}

	if err := xml.NewDecoder(bytes.NewReader(resp.Bytes())).Decode(&response); err != nil {
		return nil, err
	}

	date := datetime.Format(DATEFORMAT)

	ok, currencies := findFirstValidDay(datetime, response.Days)
	if ok {
		rates, err := parseRates(base, symbols, currencies)
		if err != nil {
			return nil, err
		}

		return &currency.ResponseAPI{
			Base:  base,
			Date:  date,
			Rates: rates,
		}, nil
	}

	return &currency.ResponseAPI{
		Base:  base,
		Date:  date,
		Rates: map[string]float64{},
	}, nil
}

func (api *ECBAPI) History(
	start time.Time,
	end time.Time,
	base string,
	symbols []string,
) (*currency.HistoryResponseAPI, error) {
	var response ECBResponse

	req := restc.Get("stats/eurofxref/eurofxref-hist.xml").
		SetHeader("Accept", restc.TypeApplicationXML)
	resp, err := api.client.Execute(req)
	if err != nil {
		return nil, err
	}

	if err := xml.NewDecoder(bytes.NewReader(resp.Bytes())).Decode(&response); err != nil {
		return nil, err
	}

	global := map[string]map[string]float64{}
	date := start
	for date.Unix() <= end.Unix() {
		ok, currencies := findFirstValidDay(date, response.Days)
		if ok {
			rates, err := parseRates(base, symbols, currencies)
			if err != nil {
				return nil, err
			}
			global[date.Format(DATEFORMAT)] = rates
		}
		date = date.AddDate(0, 0, 1)
	}

	return &currency.HistoryResponseAPI{
		Base:  base,
		Date:  start.Format(DATEFORMAT),
		Rates: global,
	}, nil
}

// Unexported functions

func findFirstValidDay(datetime time.Time, days []ECBDay) (bool, []ECBCurrencies) {
	date := datetime
	for date.Unix() > datetime.AddDate(0, 0, -5).Unix() {
		ok, currencies := findDay(date, days)
		if ok {
			return true, currencies
		}
		date = date.AddDate(0, 0, -1)
	}

	return false, []ECBCurrencies{}
}

func findDay(datetime time.Time, days []ECBDay) (bool, []ECBCurrencies) {
	day := datetime.Format(DATEFORMAT)
	for _, entry := range days {
		if day == entry.Date {
			return true, entry.Currencies
		}
	}

	return false, []ECBCurrencies{}
}

func parseRates(base string, symbols []string, currencies []ECBCurrencies) (map[string]float64, error) {
	rates := map[string]float64{}

	convert := 1.0
	if strings.ToUpper(base) != "EUR" {
		var ok bool
		ok, convert = findRate(base, currencies)
		if !ok {
			return rates, fmt.Errorf("can't find base code '%s' in returned list", base)
		}
	}

	if isInSymbols("EUR", symbols) {
		rates["EUR"] = 1 / convert
	}
	for _, entry := range currencies {
		code := strings.ToUpper(entry.Code)
		if isInSymbols(code, symbols) {
			rates[code] = entry.Value / convert
		}
	}

	return rates, nil
}

func findRate(code string, currencies []ECBCurrencies) (bool, float64) {
	for _, entry := range currencies {
		if strings.EqualFold(code, entry.Code) {
			return true, entry.Value
		}
	}

	return false, 1.0
}

func isInSymbols(code string, symbols []string) bool {
	for _, symbol := range symbols {
		if strings.EqualFold(symbol, code) {
			return true
		}
	}

	return false
}
