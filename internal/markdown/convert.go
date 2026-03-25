package markdown

import (
	"io"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/strikethrough"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
	"golang.org/x/net/html/charset"
)

var htmlConverter = converter.NewConverter(
	converter.WithPlugins(
		base.NewBasePlugin(),
		commonmark.NewCommonmarkPlugin(),
		strikethrough.NewStrikethroughPlugin(),
		table.NewTablePlugin(),
	),
)

// ConvertHTML converte um [io.Reader] de HTML para Markdown. Content-Type deve
// ser informado uma vez que o conversor presume que o input é UTF-8.
func ConvertHTML(r io.Reader, contentType string) (string, error) {
	rd, err := charset.NewReader(r, contentType)
	if err != nil {
		return "", err
	}

	md, err := htmlConverter.ConvertReader(rd)
	if err != nil {
		return "", err
	}
	return string(md), nil
}
