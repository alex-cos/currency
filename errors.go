package currency

import (
	"fmt"
)

func ErrNewRequest(err error) error {
	return fmt.Errorf("failed to build HTTP request: %w", err)
}

func ErrDoRequest(err error) error {
	return fmt.Errorf("failed to execute HTTP request: %w", err)
}

func ErrReadBody(method, url string, err error) error {
	return fmt.Errorf("failed to Read HTTP response body '%s:%s': %w",
		method, url, err)
}

func ErrParseContentType(method, url string, err error) error {
	return fmt.Errorf("failed to parse HTTP content type '%s:%s': %w",
		method, url, err)
}

func ErrEmptyBody(method, url string) error {
	return fmt.Errorf("unexpected empty body response '%s:%s'", method, url)
}

func ErrUnmarshal(method, url string, err error) error {
	return fmt.Errorf("error when unmarshaling body '%s:%s': %w",
		method, url, err)
}

func ErrUnsupportedMimeType(method, url, mimeType string) error {
	return fmt.Errorf("unsupported mime type '%s:%s': %s",
		method, url, mimeType)
}

func ErrEmptyType(method, url string) error {
	return fmt.Errorf("empty predifined type '%s:%s'", method, url)
}

func ErrGeneric(msg string) error {
	return fmt.Errorf("returned an error: %s", msg)
}

func ErrUnexpectedError(method, url string) error {
	return fmt.Errorf("API returned an unexpected error '%s:%s'", method, url)
}

func ErrFailedNotification(method, url string) error {
	return fmt.Errorf("API return a failed notifiation '%s:%s'", method, url)
}
