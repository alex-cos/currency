package currency

import (
	"math/rand"
	"time"

	"github.com/stretchr/testify/mock"
)

const day = 24 * time.Hour

type Mock struct {
	mock.Mock
}

func NewMock() Currency {
	mockAPI := &Mock{}

	mockAPI.On("Ping").Return(nil)
	mockAPI.On("Latest", mock.Anything, mock.Anything).Return(nil, nil)
	mockAPI.On("ForDate", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	mockAPI.On("History", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	return mockAPI
}

func (mock *Mock) Ping() error {
	args := mock.Called()

	return args.Error(0)
}

func (mock *Mock) Latest(base string, symbols []string) (*ResponseAPI, error) {
	args := mock.Called(base, symbols)
	resp := args.Get(0)
	err := args.Error(1)
	if resp == nil && err == nil {
		resp = response(base, symbols)
	}

	return resp.(*ResponseAPI), err
}

func (mock *Mock) ForDate(date time.Time, base string, symbols []string) (*ResponseAPI, error) {
	args := mock.Called(date, base, symbols)
	resp := args.Get(0)
	err := args.Error(1)
	if resp == nil && err == nil {
		resp = response(base, symbols)
	}

	return resp.(*ResponseAPI), err
}

func (mock *Mock) History(start, end time.Time, base string, symbols []string) (*HistoryResponseAPI, error) {
	args := mock.Called(start, end, base, symbols)
	resp := args.Get(0)
	err := args.Error(1)
	if resp == nil && err == nil {
		resp = history(start, end, base, symbols)
	}

	return resp.(*HistoryResponseAPI), err
}

// Unexported function

func generateRandom() float64 {
	return (rand.Float64() * 0.04) + 0.98 // nolint: gosec
}

func response(base string, symbols []string) *ResponseAPI {
	rates := map[string]float64{}
	for _, symbol := range symbols {
		rates[symbol] = generateRandom()
	}

	return &ResponseAPI{
		Base:  base,
		Date:  time.Now().UTC().Truncate(day).Format(time.RFC3339),
		Rates: rates,
	}
}

func history(start, end time.Time, base string, symbols []string) *HistoryResponseAPI {
	rates := map[string]map[string]float64{}
	date := start.UTC().Truncate(day)
	for date.Unix() < end.Unix() {
		strdate := date.Format("2006-01-02")
		rates[strdate] = map[string]float64{}
		for _, symbol := range symbols {
			rates[strdate][symbol] = generateRandom()
		}
		date = date.AddDate(0, 0, 1)
	}

	return &HistoryResponseAPI{
		Base:  base,
		Date:  start.Format(time.RFC3339),
		Rates: rates,
	}
}
