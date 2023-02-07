package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	maxHeaderBytes = 1 << 20
	readTimeout    = 10 * time.Second
	writeTimeout   = 10 * time.Second

	timeForGracefulShutdown = 5 * time.Second
)

// Server is a struct which handles the requests.
type Server struct {
	*http.Server
}

func New(port string, handler http.Handler) Server {
	return Server{&http.Server{
		Addr:           ":" + port,
		Handler:        handler,
		MaxHeaderBytes: maxHeaderBytes,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
	}}
}

func catchOSInterrupt(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
		logrus.Info("os interruption call")
		cancel()
	}()
}

// After Run method Server starts to listen port and response to  the reqeusts.
// Run function provide the abitility of the gracefule shutdown.
func (s *Server) Run(ctx context.Context) error {
	logrus.Info("start server")
	ctx, cancel := context.WithCancel(ctx)

	catchOSInterrupt(cancel)

	servError := make(chan error, 1)

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			servError <- fmt.Errorf("listen: %w", err)
		}
	}()

	select {
	case err := <-servError:
		return err
	case <-ctx.Done():
		ctxShutDown, cancel := context.WithTimeout(context.Background(), timeForGracefulShutdown)
		defer cancel()

		logrus.Info("start graceful shutdown")
		if err := s.Shutdown(ctxShutDown); err != nil { //nolint: contextcheck // create new context for graceful shutdown
			return fmt.Errorf("shutdown: %w", err)
		}
		logrus.Info("graceful shutdown ends")
	}

	return nil
}
