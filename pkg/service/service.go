package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Dyleme/apod.git/pkg/model"
	retry "github.com/avast/retry-go/v4"
	"github.com/google/uuid"
)

const (
	imageBucket = "images"
)

type APOD interface {
	GetImageForDate(ctx context.Context, date time.Time) ([]byte, string, error)
}

type Repository interface {
	AddPendingImage(ctx context.Context, date time.Time) error
	SetImagePath(ctx context.Context, date time.Time, path string) error
	FetchImagePath(ctx context.Context, date time.Time) (string, error)
	FetchAllImagePaths(ctx context.Context) ([]string, error)
}

type Storager interface {
	UploadFile(ctx context.Context, bucket, filename string, data []byte) (string, error)
}

type Service struct {
	apod    APOD
	repo    Repository
	storage Storager
}

func New(apod APOD, repo Repository, storage Storager) *Service {
	return &Service{
		apod:    apod,
		repo:    repo,
		storage: storage,
	}
}

func (s *Service) GetImageURLForDate(ctx context.Context, date time.Time) (string, error) {
	var (
		url string
		err error
	)

	url, err = s.repo.FetchImagePath(ctx, date)
	if err != nil {
		if errors.Is(err, model.ErrImageNotExists) {
			url, err = s.downloadImage(ctx, date)
			if err != nil {
				return "", fmt.Errorf("download image %v: %w", date, err)
			}

			return url, nil
		}

		if errors.Is(err, model.ErrPendingImage) { // Image is already downloading. Wait until download competes
			url, err = s.checkImageStatus(ctx, date)
			if err != nil {
				return "", fmt.Errorf("check image status %v: %w", date, err)
			}

			return url, nil
		}

		return "", fmt.Errorf("fetch image path: %w", err)
	}

	return url, nil
}

func (s *Service) checkImageStatus(ctx context.Context, date time.Time) (string, error) {
	const attemptsAmount = 5

	var (
		path string
		err  error
	)

	err = retry.Do(func() error {
		path, err = s.repo.FetchImagePath(ctx, date)

		return fmt.Errorf("fetch image path %v: %w", date, err)
	},
		retry.Attempts(attemptsAmount),
		retry.RetryIf(func(err error) bool {
			return errors.Is(err, model.ErrPendingImage)
		}),
	)
	if err != nil {
		return "", fmt.Errorf("in retry: %w", err)
	}

	return path, nil
}

func (s *Service) downloadImage(ctx context.Context, date time.Time) (string, error) {
	err := s.repo.AddPendingImage(ctx, date)
	if err != nil {
		return "", fmt.Errorf("set pending image %v: %w", date, err)
	}

	image, filetype, err := s.apod.GetImageForDate(ctx, date)
	if err != nil {
		return "", fmt.Errorf("get image from date %v: %w", date, err)
	}

	filename := uuid.NewString() + filetype

	path, err := s.storage.UploadFile(ctx, imageBucket, filename, image)
	if err != nil {
		return "", fmt.Errorf("upload file bucket[%q], filename[%q], len(image)[%v]:%w", imageBucket, filename, len(image), err)
	}

	err = s.repo.SetImagePath(ctx, date, path)
	if err != nil {
		return "", fmt.Errorf("set image url %q: %w", path, err)
	}

	return path, nil
}