package main

import (
	"context"
	"log"
	"os"

	"github.com/Dyleme/apod.git/pkg/apod-service"
	"github.com/Dyleme/apod.git/pkg/database/postgres"
	"github.com/Dyleme/apod.git/pkg/handler"
	"github.com/Dyleme/apod.git/pkg/handler/imagehandler"
	"github.com/Dyleme/apod.git/pkg/repository"
	"github.com/Dyleme/apod.git/pkg/server"
	"github.com/Dyleme/apod.git/pkg/service"
	"github.com/Dyleme/apod.git/pkg/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"
)

func main() {
	apodService := initAPOD()

	stor, err := initMinio()
	if err != nil {
		log.Fatal(err)
	}

	repo, err := initRepository()
	if err != nil {
		log.Fatal(err)
	}

	imageService := service.New(apodService, repo, stor)
	imageHandler := imagehandler.New(imageService)
	hand := handler.New(imageHandler)

	appPort := os.Getenv("APP_PORT")
	serv := server.New(appPort, hand.InitRouters())

	err = serv.Run(context.Background())
	if err != nil {
		logrus.Fatal("error on server", err)
	}
}

func initAPOD() *apod.Service {
	return apod.NewService(os.Getenv("NASA_API_KEY"))
}

func initMinio() (*storage.Minio, error) {
	minioConfig, err := storage.InitConfig()
	if err != nil {
		return nil, err
	}

	stor, err := storage.NewMinioStorage(*minioConfig)
	if err != nil {
		return nil, err
	}

	return stor, nil
}

func initRepository() (*repository.Repository, error) {
	dbCfg := postgres.InitConfig()

	db, err := postgres.NewDB(dbCfg)
	if err != nil {
		return nil, err
	}

	repo, err := repository.New(db)
	if err != nil {
		return nil, err
	}

	return repo, nil
}
