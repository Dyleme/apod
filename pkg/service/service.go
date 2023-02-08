package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Dyleme/apod.git/pkg/model"
)

const (
	imageBucket = "images"
)

type APODer interface {
	GetImageForDate(ctx context.Context, date time.Time) (img []byte, extension string, err error)
}

type Repository interface {
	AddImage(ctx context.Context, date time.Time, path string) error
	FetchImagePath(ctx context.Context, date time.Time) (string, error)
	FetchAllImagePaths(ctx context.Context) ([]string, error)
}

type Storager interface {
	UploadFile(ctx context.Context, bucket, filename string, data []byte) (url string, err error)
}

type Service struct {
	repo       Repository
	downloader downloaders
}

func New(apod APODer, repo Repository, storage Storager) *Service {
	return &Service{
		repo: repo,
		downloader: downloaders{
			mx:      sync.Mutex{},
			waiters: make(map[time.Time][]chan<- error),
			apod:    apod,
			storage: storage,
			repo:    repo,
		},
	}
}

func (s *Service) GetImageURLForDate(ctx context.Context, date time.Time) (string, error) {
	var (
		url string
		err error
	)

	url, err = s.repo.FetchImagePath(ctx, date)
	if err == nil { // eq nil
		return url, nil
	}

	if errors.Is(err, model.ErrImageNotExists) {
		err := s.downloadImage(ctx, date)
		if err != nil {
			return "", err
		}

		url, err := s.repo.FetchImagePath(ctx, date)
		if err != nil {
			return "", err
		}

		return url, nil
	}

	return "", fmt.Errorf("fetch image path: %w", err)
}

// downloadImage is function which downloads image from apod and uploads it to the storage.
// If it is called concurrently, only one operation of downloading and saving is performed.
func (s *Service) downloadImage(ctx context.Context, date time.Time) error {
	waiter := make(chan error)

	s.downloader.registerDownloadWaiter(ctx, date, waiter)

	resErr := <-waiter
	if resErr != nil {
		return resErr
	}

	return nil
}

func (s *Service) GetAlbumURLs(ctx context.Context) ([]string, error) {
	urls, err := s.repo.FetchAllImagePaths(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch all images: %w", err)
	}

	return urls, nil
}
