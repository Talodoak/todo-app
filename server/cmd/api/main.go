package main

import (
	"context"
	"github.com/Talodoak/todo-app/internal/service"
	repository2 "github.com/Talodoak/todo-app/internal/storage"
	"github.com/Talodoak/todo-app/internal/storage/postgres"
	"github.com/Talodoak/todo-app/internal/transport/rest"
	"github.com/Talodoak/todo-app/internal/transport/rest/handler"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
)

func goDotEnvVariable(key string) string {
	err := godotenv.Load(".env")

	if err != nil {
		logrus.Error("Error loading .env file", err)
	}

	return os.Getenv(key)
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	logrus.SetFormatter(new(logrus.JSONFormatter))
	if err := initConfig(); err != nil {
		logrus.Fatalf("error initializing configs: %s", err)
	}

	db, err := postgres.NewPostgresDB(postgres.Config{
		Host:     goDotEnvVariable("POSTGRES_HOST"),
		Port:     goDotEnvVariable("POSTGRES_PORT"),
		Username: goDotEnvVariable("POSTGRES_USERNAME"),
		DBName:   goDotEnvVariable("POSTGRES_DATABASENAME"),
		SSLMode:  goDotEnvVariable("POSTGRES_SSL_MODE"),
		Password: goDotEnvVariable("POSTGRES_PASSWORD"),
	})
	if err != nil {
		logrus.Fatalf("failed to initialize db: %s", err.Error())
	}

	repos := repository2.NewRepository(db)
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)

	server := new(rest.Server)
	go func() {
		if servErr := server.Run(goDotEnvVariable("APP_PORT"), handlers.InitRoutes()); servErr != nil {
			logrus.Fatalf("error occured while running http server: %s", servErr.Error())
		}
	}()

	logrus.Printf("TodoApp Started at port %s", goDotEnvVariable("APP_PORT"))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Print("TodoApp Shutting Down")

	if err := server.Shutdown(context.Background()); err != nil {
		logrus.Errorf("error occured on server shutting down: %s", err.Error())
	}

	if err := db.Close(); err != nil {
		logrus.Errorf("error occured on db connection close: %s", err.Error())
	}

}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
