package apod

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Apod struct {
	apiKey string
}

func NewService(apiKey string) *Apod {
	return &Apod{apiKey: apiKey}
}

type apodResponse struct {
	URL string `json:"url"`
}

func (as *Apod) getApodForDate(ctx context.Context, date time.Time) (*apodResponse, error) {
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

	var apodResp apodResponse

	err = json.Unmarshal(bts, &apodResp)
	if err != nil {
		return nil, fmt.Errorf("unmarshal %q: %w", string(bts), err)
	}

	return &apodResp, nil
}

func (as *Apod) getImage(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request %q: %w", req.RequestURI, err)
	}
	defer resp.Body.Close()

	image, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func getFiletype(url string) string {
	pointIndex := strings.LastIndex(url, ".")
	if pointIndex == -1 || pointIndex+1 >= len(url) {
		return ""
	}

	return url[pointIndex:]
}

func (as *Apod) GetImageForDate(ctx context.Context, date time.Time) ([]byte, string, error) {
	apod, err := as.getApodForDate(ctx, date)
	if err != nil {
		return nil, "", err
	}

	image, err := as.getImage(ctx, apod.URL)
	if err != nil {
		return nil, "", err
	}

	filetype := getFiletype(apod.URL)

	return image, filetype, nil
}
