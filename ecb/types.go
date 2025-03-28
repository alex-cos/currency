package ecb

import (
	"encoding/xml"
)

type ECBResponse struct {
	XMLName xml.Name
	Days    []ECBDay `xml:"Cube>Cube"`
}

type ECBDay struct {
	Date       string          `xml:"time,attr"`
	Currencies []ECBCurrencies `xml:"Cube"`
}

type ECBCurrencies struct {
	Code  string  `xml:"currency,attr"`
	Value float64 `xml:"rate,attr"`
}
