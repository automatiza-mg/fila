package markdown

import (
	"mime"
	"strings"
)

// IsHTML reporta se o Content-Type informado é text/html.
func IsHTML(contentType string) bool {
	mediaType, _, _ := mime.ParseMediaType(contentType)
	return mediaType == "text/html" || strings.HasSuffix(mediaType, "+html")
}
