package image

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Service interface {
	GetImageURLForDate(ctx context.Context, date time.Time) (string, error)
	GetAlbumURLs(ctx context.Context) ([]string, error)
}

type Handler struct {
	service Service
}

func New(imageService Service) *Handler {
	return &Handler{service: imageService}
}

type urlResponse struct {
	URL string `json:"url"`
}

func (ih *Handler) GetForDate(w http.ResponseWriter, r *http.Request) {
	dateString := chi.URLParam(r, "date")

	date, err := time.Parse(time.DateOnly, dateString)
	if err != nil {
		responseError(w, err, http.StatusBadRequest)

		return
	}

	url, err := ih.service.GetImageURLForDate(r.Context(), date)
	if err != nil {
		responseError(w, err, http.StatusInternalServerError)

		return
	}

	responseJSON(w, urlResponse{URL: url})
}

func (ih *Handler) GetAlbumImages(w http.ResponseWriter, r *http.Request) {
	urls, err := ih.service.GetAlbumURLs(r.Context())
	if err != nil {
		responseError(w, err, http.StatusInternalServerError)

		return
	}

	urlsResponse := make([]urlResponse, 0, len(urls))

	for _, u := range urls {
		urlsResponse = append(urlsResponse, urlResponse{URL: u})
	}

	responseJSON(w, urlsResponse)
}
