package apod

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"time"
)

type Service struct {
	apiKey string
}

func NewService(apiKey string) *Service {
	return &Service{apiKey: apiKey}
}

type apodResponse struct {
	URL string `json:"url"`
}

const errorStatusCode = 400

func (as *Service) getAPODForDate(ctx context.Context, date time.Time) (*apodResponse, error) {
	urlA, err := url.ParseRequestURI("https://api.nasa.gov/planetary/apod")
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	values := urlA.Query()
	values.Set("api_key", as.apiKey)
	values.Set("date", date.Format(time.DateOnly))
	urlA.RawQuery = values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlA.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request %q: %w", req.RequestURI, err)
	}
	defer resp.Body.Close()

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode >= errorStatusCode {
		return nil, fmt.Errorf("status code %v, body %q", resp.StatusCode, string(bts))
	}

	var apodResp apodResponse

	err = json.Unmarshal(bts, &apodResp)
	if err != nil {
		return nil, fmt.Errorf("unmarshal %q: %w", string(bts), err)
	}

	return &apodResp, nil
}

func (as *Service) getFile(ctx context.Context, url string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("new request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("do request %q: %w", req.RequestURI, err)
	}
	defer resp.Body.Close()

	image, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	contentType := resp.Header.Get("Content-Type")
	ext, err := mime.ExtensionsByType(contentType)
	if err != nil {
		return nil, "", fmt.Errorf("extensions by type: %w", err)
	}

	return image, ext[0], nil
}

func (as *Service) GetImageForDate(ctx context.Context, date time.Time) ([]byte, string, error) {
	apod, err := as.getAPODForDate(ctx, date)
	if err != nil {
		return nil, "", err
	}

	image, ext, err := as.getFile(ctx, apod.URL)
	if err != nil {
		return nil, "", err
	}

	return image, ext, nil
}
