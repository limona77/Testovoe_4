package app

import (
	"zadanie-6105/config"
	"zadanie-6105/internal/controller"
	"zadanie-6105/internal/repository"
	"zadanie-6105/internal/service"
	"zadanie-6105/pkg/postgres"
	"zadanie-6105/slogger"

	"github.com/gofiber/fiber/v2"

	"github.com/gookit/slog"
)

func Run() {
	slogger.SetLogger()

	slog.Info("init config")
	cfg := config.NewConfig()
	slog.Info("config ok")

	slog.Info("connecting to postgres")
	db := postgres.New(cfg.PostgresConn)
	defer db.Close()
	slog.Info("connect to postgres ok")

	slog.Info("init repositories")
	repositories := repository.NewRepositories(db)

	slog.Info("init services")
	deps := service.ServicesDeps{
		Repository: repositories,
	}

	services := service.NewServices(deps)

	fiberConfig := fiber.Config{}
	app := fiber.New(fiberConfig)

	controller.NewRouter(app, services)

	slog.Info("starting fiber server")
	slog.Fatal(app.Listen(cfg.ServerAddress))
}
