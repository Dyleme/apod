package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Dyleme/apod.git/pkg/model"
	retry "github.com/avast/retry-go/v4"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	imageBucket = "images"
)

type APODer interface {
	GetImageForDate(ctx context.Context, date time.Time) (img []byte, extension string, err error)
}

type Repository interface {
	AddPendingImage(ctx context.Context, date time.Time) error
	SetImagePath(ctx context.Context, date time.Time, path string) error
	FetchImagePath(ctx context.Context, date time.Time) (string, error)
	FetchAllImagePaths(ctx context.Context) ([]string, error)
	DeleteImage(ctx context.Context, date time.Time) error
}

type Storager interface {
	UploadFile(ctx context.Context, bucket, filename string, data []byte) (url string, err error)
}

type Service struct {
	apod    APODer
	repo    Repository
	storage Storager
}

func New(apod APODer, repo Repository, storage Storager) *Service {
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
	fmt.Println(url, err)
	if err != nil {
		if errors.Is(err, model.ErrImageNotExists) {
			logrus.Infof("image %v not exist", date)
			url, err = s.downloadImage(ctx, date)
			if err != nil {
				delErr := s.repo.DeleteImage(ctx, date)
				if delErr != nil {
					err = fmt.Errorf("download image: %w (delete image %w)", err, delErr)
					logrus.Error(err)
					return "", err
				}

				err = fmt.Errorf("download image %v: %w", date, err)
				logrus.Error(err)
				return "", err
			}

			return url, nil
		}

		if errors.Is(err, model.ErrPendingImage) { // Image is already downloading. Wait until download competes
			logrus.Infof("image is pending %v", date)
			url, err = s.checkImageStatus(ctx, date)
			if err != nil {
				err = fmt.Errorf("check image status %v: %w", date, err)
				logrus.Error(err)
				return "", err
			}

			return url, nil
		}

		logrus.Error(err)
		return "", fmt.Errorf("fetch image path: %w", err)
	}

	return url, nil
}

func (s *Service) checkImageStatus(ctx context.Context, date time.Time) (string, error) {
	const (
		attemptsAmount = 50
		startDelay     = 250 * time.Millisecond
	)

	var (
		path string
		err  error
	)

	err = retry.Do(func() error {
		path, err = s.repo.FetchImagePath(ctx, date)
		if err != nil {
			return fmt.Errorf("fetch image path %v: %w", date, err)
		}

		return nil
	},
		retry.Delay(startDelay),
		retry.DelayType(retry.FixedDelay),
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

	image, ext, err := s.apod.GetImageForDate(ctx, date)
	if err != nil {
		return "", fmt.Errorf("get image from date %v: %w", date, err)
	}

	filename := uuid.NewString() + ext

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

func (s *Service) GetAlbumURLs(ctx context.Context) ([]string, error) {
	urls, err := s.repo.FetchAllImagePaths(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch all images: %w", err)
	}

	return urls, nil
}
