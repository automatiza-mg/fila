package soap

import (
	"encoding/xml"
	"fmt"
)

type Envelope[T any] struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    Body[T]  `xml:"Body"`
}

type Body[T any] struct {
	XMLName xml.Name `xml:"Body"`
	Content T
}

type Fault struct {
	XMLName xml.Name    `xml:"Fault"`
	Code    string      `xml:"faultcode"`
	Message string      `xml:"faultstring"`
	Detail  FaultDetail `xml:"detail"`
}

type FaultDetail struct {
	Items []FaultDetailItem `xml:"item"`
}

type FaultDetailItem struct {
	Key   string `xml:"key"`
	Value string `xml:"value"`
}

type Error struct {
	Status int
	Fault  Envelope[Fault]
}

func (e *Error) Error() string {
	return fmt.Sprintf("failed to execute action (%d): %s", e.Status, e.Fault.Body.Content.Message)
}

func NewError(status int, fault Envelope[Fault]) *Error {
	return &Error{
		Status: status,
		Fault:  fault,
	}
}
