package imagehandler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Dyleme/apod.git/pkg/models"
	"github.com/go-chi/chi/v5"
)

type Service interface {
	GetImageURLForDate(ctx context.Context, date time.Time) (string, error)
	GetAlbum(ctx context.Context) ([]models.AlbumRecord, error)
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

type albumRecordResponse struct {
	Date string `json:"date"`
	URL  string `json:"url"`
}

func (ih *Handler) GetForDate(w http.ResponseWriter, r *http.Request) {
	dateString := chi.URLParam(r, "date")

	date, err := time.Parse(time.DateOnly, dateString)
	if err != nil {
		responseError(w, err, http.StatusBadRequest)

		return
	}

	if date.After(time.Now().UTC()) {
		err = fmt.Errorf("provided date %q is in future", dateString)
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
	urls, err := ih.service.GetAlbum(r.Context())
	if err != nil {
		responseError(w, err, http.StatusInternalServerError)

		return
	}

	urlsResponse := make([]albumRecordResponse, 0, len(urls))

	fmt.Println(urlsResponse)
	for _, u := range urls {
		urlsResponse = append(urlsResponse, albumRecordResponse{URL: u.URL, Date: u.Date.Format(time.DateOnly)})
	}

	fmt.Println(urlsResponse)
	responseJSON(w, urlsResponse)
}
