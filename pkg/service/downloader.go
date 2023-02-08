package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type downloaders struct {
	mx      sync.Mutex
	waiters map[time.Time][]chan<- error
	apod    APODer
	storage Storager
	repo    Repository
}

func (d *downloaders) download(ctx context.Context, date time.Time) {
	image, ext, err := d.apod.GetImageForDate(ctx, date)
	if err != nil {
		d.response(fmt.Errorf("get image from date %v: %w", date, err), date)
		return
	}

	filename := uuid.NewString() + ext

	path, err := d.storage.UploadFile(ctx, imageBucket, filename, image)
	if err != nil {
		d.response(fmt.Errorf("upload file bucket[%q], filename[%q], len(image)[%v]:%w", imageBucket, filename, len(image), err), date)
		return
	}

	err = d.repo.AddImage(ctx, date, path)
	if err != nil {
		d.response(fmt.Errorf("set image url %q: %w", path, err), date)
		return
	}

	d.response(nil, date)
}

func (ds *downloaders) response(err error, date time.Time) {
	ds.mx.Lock()
	defer ds.mx.Unlock()

	for _, w := range ds.waiters[date] {
		w <- err
	}
	delete(ds.waiters, date)
}

// registerDownloadWaiter function is used to start downloading and saving images.
// After image saving completes waiter will receive downloading error.
func (d *downloaders) registerDownloadWaiter(ctx context.Context, date time.Time, waiter chan<- error) {
	d.mx.Lock()
	defer d.mx.Unlock()

	if _, ok := d.waiters[date]; !ok {
		d.waiters[date] = make([]chan<- error, 0, 2)
		go d.download(ctx, date)
	}

	d.waiters[date] = append(d.waiters[date], waiter)
}
