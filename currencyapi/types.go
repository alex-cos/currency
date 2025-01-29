package currencyapi

import "fmt"

type StatusResponseAPI struct {
	AccountID int64 `json:"account_id"`
	Quotas    map[string]struct {
		Total     int64 `json:"total"`
		Used      int64 `json:"used"`
		Remaining int64 `json:"remaining"`
	} `json:"quotas"`
}

func (item *StatusResponseAPI) String() string {
	return fmt.Sprintf("%+v", *item)
}

type RatesResponseAPI struct {
	Meta struct {
		LastUpdatedAt string `json:"last_updated_at"`
	} `json:"meta"`
	Data map[string]struct {
		Code  string  `json:"code"`
		Value float64 `json:"value"`
	} `json:"data"`
}

func (item *RatesResponseAPI) String() string {
	return fmt.Sprintf("%+v", *item)
}

type HistoryResponseAPI struct {
	Data []struct {
		Datetime   string `json:"datetime"`
		Currencies map[string]struct {
			Code  string  `json:"code"`
			Value float64 `json:"value"`
		} `json:"currencies"`
	} `json:"data"`
}

func (item *HistoryResponseAPI) String() string {
	return fmt.Sprintf("%+v", *item)
}
