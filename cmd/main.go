package main

import (
	"context"
	"log"
	"os"

	"github.com/Dyleme/apod.git/pkg/apod-service"
	"github.com/Dyleme/apod.git/pkg/database"
	"github.com/Dyleme/apod.git/pkg/handler"
	image "github.com/Dyleme/apod.git/pkg/handler/images_handler"
	"github.com/Dyleme/apod.git/pkg/server"
	"github.com/Dyleme/apod.git/pkg/service"
	"github.com/Dyleme/apod.git/pkg/storage"
	_ "github.com/jackc/pgx/v5"
)

func main() {
	apodService := apod.NewService(os.Getenv("NASA_API_KEY"))
	storConfig, err := storage.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	stor, err := storage.NewMinioStorage(*storConfig)
	if err != nil {
		log.Fatal(err)
	}
	dbCfg := database.InitConfig()
	db, err := database.New(dbCfg)
	if err != nil {
		log.Fatal(err)
	}

	repo, err := database.NewRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	serv := service.New(apodService, repo, stor)

	imageHandler := image.New(serv)
	hand := handler.New(imageHandler)

	s := server.New("8080", hand.InitRouters())
	err = s.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
