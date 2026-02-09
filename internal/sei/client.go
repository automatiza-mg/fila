package sei

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/automatiza-mg/fila/internal/soap"
)

type Parametros[T any] struct {
	Items []T `xml:"item"`
}

type Items[T any] struct {
	Items []T `xml:"item" json:"items"`
}

type Client struct {
	cfg  *Config
	http *http.Client
}

func NewClient(cfg *Config) *Client {
	return &Client{
		cfg:  cfg,
		http: http.DefaultClient,
	}
}

func makeSoapError(status int, r io.Reader) error {
	var fault soap.Envelope[soap.Fault]
	err := xml.NewDecoder(r).Decode(&fault)
	if err != nil {
		return err
	}
	return soap.NewError(status, fault)
}

func doReq[Req any, Res any](ctx context.Context, client *http.Client, url string, req Req) (*Res, error) {
	body, err := xml.Marshal(soap.Envelope[Req]{
		Body: soap.Body[Req]{
			Content: req,
		},
	})
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(xml.Header+string(body)))
	if err != nil {
		return nil, err
	}

	res, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, makeSoapError(res.StatusCode, res.Body)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(b))

	var resp soap.Envelope[Res]
	err = xml.NewDecoder(strings.NewReader(string(b))).Decode(&resp)
	return &resp.Body.Content, err
}
