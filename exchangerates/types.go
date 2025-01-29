package exchangerates

import "fmt"

// ResponseAPI defines the basic response for the API.
type ResponseAPI struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

func (item *ResponseAPI) String() string {
	return fmt.Sprintf("%+v", *item)
}

// HistoryResponseAPI defines the response for History returned call.
type HistoryResponseAPI struct {
	Base  string                        `json:"base"`
	Date  string                        `json:"date"`
	Rates map[string]map[string]float64 `json:"rates"`
}

func (item *HistoryResponseAPI) String() string {
	return fmt.Sprintf("%+v", *item)
}
