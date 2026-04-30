package docintel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type AzureDocIntel struct {
	endpoint string
	apiKey   string
	http     *http.Client
}

func NewAzureDocIntel(cfg *Config) *AzureDocIntel {
	return &AzureDocIntel{
		endpoint: cfg.AzureURL,
		apiKey:   cfg.AzureApiKey,
		http:     http.DefaultClient,
	}
}

type AnalyzeOperation struct {
	Status        string        `json:"status"`
	AnalyzeResult AnalyzeResult `json:"analyzeResult"`
}

type AnalyzeResult struct {
	ApiVersion      string `json:"apiVersion"`
	ModelID         string `json:"modelId"`
	Content         string `json:"content"`
	StringIndexType string `json:"stringIndexType"`
	ContentFormat   string `json:"contentFormat"`
}

// ExtractText extrai o texto (em formato markdown) de um [io.Reader] usando a API da Azure Document Intelligence.
//
// Referência: https://learn.microsoft.com/en-us/rest/api/aiservices/document-models/analyze-document-from-stream?view=rest-aiservices-v4.0%20(2024-11-30)&tabs=HTTP
func (a *AzureDocIntel) ExtractText(ctx context.Context, r io.Reader, contentType string) (string, error) {
	q := make(url.Values)
	q.Set("locale", "pt-BR")
	q.Set("api-version", "2024-11-30")
	q.Set("outputContentFormat", "markdown")

	endpoint := strings.TrimSuffix(a.endpoint, "/")
	url := fmt.Sprintf("%s/documentintelligence/documentModels/%s:analyze?%s", endpoint, "prebuilt-layout", q.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, r)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Ocp-Apim-Subscription-Key", a.apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	// Lê o corpo da requisição e retorna o erro caso status seja diferente de 202.
	if res.StatusCode != http.StatusAccepted {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("azure returned unexpected status: %d (%s)", res.StatusCode, string(b))
	}

	operationLocation := res.Header.Get("Operation-Location")
	return a.poolResult(ctx, operationLocation)
}

func (a *AzureDocIntel) poolResult(ctx context.Context, location string) (string, error) {
	poolCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	ticker := time.NewTicker(1500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-poolCtx.Done():
			return "", fmt.Errorf("polling timed out: %w", poolCtx.Err())
		case <-ticker.C:
			op, err := a.getOperationStatus(poolCtx, location)
			if err != nil {
				return "", err
			}

			switch op.Status {
			case "succeeded":
				return op.AnalyzeResult.Content, nil
			case "failed":
				return "", errors.New("failed to analyze document")
			case "running", "notStarted":
				continue
			default:
				return "", fmt.Errorf("unexpected status: %s", op.Status)
			}
		}
	}
}

func (a *AzureDocIntel) getOperationStatus(ctx context.Context, location string) (*AnalyzeOperation, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Ocp-Apim-Subscription-Key", a.apiKey)

	res, err := a.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		var op AnalyzeOperation
		err := json.NewDecoder(res.Body).Decode(&op)
		if err != nil {
			return nil, err
		}
		return &op, nil
	default:
		return nil, fmt.Errorf("failed to get result: %d", res.StatusCode)
	}
}
