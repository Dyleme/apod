package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Handler is a struct which has service interfaces.
type Handler struct {
	imagesHandler ImagesHandler
}

// This constructor initialize Handler's fields with provided arguments.
func New(imagesHandler ImagesHandler) *Handler {
	return &Handler{
		imagesHandler: imagesHandler,
	}
}

type ImagesHandler interface {
	GetForDate(w http.ResponseWriter, r *http.Request)
	GetAlbumImages(w http.ResponseWriter, r *http.Request)
}

// InitRouters() method is used to initialize all endopoints with the routers.
func (h *Handler) InitRouters() *chi.Mux {
	r := chi.NewRouter()

	r.Get("/images/{date}", h.imagesHandler.GetForDate)
	r.Get("/images", h.imagesHandler.GetAlbumImages)

	return r
}
