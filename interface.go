package currency

import "time"

// Currency abstract Exchange rates interface API.
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
