package main

import (
	"context"
	"fmt"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maximmihin/aw25/internal/dal"
	handlers "github.com/maximmihin/aw25/internal/httpcontroller"
	slogfiber "github.com/samber/slog-fiber"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
)

type Config struct {
	PostgresConnString string

	JwtPrivateKey string
	JwtPublicKey  string

	HttpServiceHost string
	HttpServicePort string

	LogLevel string
}

func main() {
	cfg := Config{
		PostgresConnString: fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_PASSWORD"),
			os.Getenv("POSTGRES_DB"),
			os.Getenv("POSTGRES_HOST"),
			os.Getenv("POSTGRES_PORT")),

		JwtPrivateKey: os.Getenv("JWT_PRIVATE_KEY"),
		JwtPublicKey:  os.Getenv("JWT_PUBLIC_KEY"),

		HttpServiceHost: os.Getenv("HTTP_SERVICE_HOST"),
		HttpServicePort: os.Getenv("HTTP_SERVICE_PORT"),

		LogLevel: os.Getenv("LOG_LEVEL"),
	}

	err := Run(cfg)
	if err != nil {
		fmt.Println(err.Error())
	}
}

// log.Info("started listen") is required for waiting func in e2e tests with Docker - add it if reassign
var merchShopHttpServiceFiberOnListenFunc fiber.OnListenHandler

var merchShopHttpServiceLogger *slog.Logger

func Run(cfg Config) error {

	ctx := context.TODO()

	logLevel, err := strconv.Atoi(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("log level need to be valid int")
	}

	if merchShopHttpServiceLogger == nil {
		merchShopHttpServiceLogger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.Level(logLevel),
			}))
	}
	log := merchShopHttpServiceLogger.With(
		slog.String("service", "merch_shop"),
		slog.String("app", "http_server"),
	)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler:          handlers.ErrorHandler,
	})

	app.Use(slogfiber.NewWithConfig(log, slogfiber.Config{
		DefaultLevel:       slog.Level(logLevel),
		ClientErrorLevel:   slog.LevelDebug,
		ServerErrorLevel:   slog.LevelDebug,
		WithUserAgent:      true,
		WithRequestID:      true,
		WithRequestBody:    true,
		WithRequestHeader:  true,
		WithResponseBody:   true,
		WithResponseHeader: true,
	}))
	app.Use(recover.New())
	app.Use(requestid.New())

	dbPool, err := pgxpool.New(ctx, cfg.PostgresConnString)
	if err != nil {
		return err
	}
	defer dbPool.Close()

	db, err := dal.New(ctx, dbPool)
	if err != nil {
		return err
	}

	h := handlers.Handlers{
		Dal:           db,
		Logger:        log,
		JWTPrivateKey: cfg.JwtPrivateKey,
	}

	api := app.Group("/api")
	api.Post("/auth", h.Auth)

	secured := api.Group("")
	secured.Use(jwtware.New(jwtware.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return handlers.ResErr(c, 401, "problems with token")
		},
		SigningKey: jwtware.SigningKey{Key: []byte(cfg.JwtPublicKey)},
	}))
	secured.Get("/buy/:item", h.BuyMerch)
	secured.Post("/sendCoin", h.CoinTransfer)
	secured.Get("/info", h.Info)

	if merchShopHttpServiceFiberOnListenFunc != nil {
		app.Hooks().OnListen(merchShopHttpServiceFiberOnListenFunc)
	} else {
		app.Hooks().OnListen(func(data fiber.ListenData) error {
			log.Info("started listen")
			return nil
		})
	}

	gc := make(chan os.Signal, 1)
	signal.Notify(gc, os.Interrupt)

	errC := make(chan error, 1)

	go func() {
		err = app.Listen(net.JoinHostPort(cfg.HttpServiceHost, cfg.HttpServicePort))
		if err != nil {
			log.Error(err.Error())
			errC <- err
		}
	}()

	select {
	case <-errC:
	case <-gc:
	}

	errDown := app.Shutdown()
	if errDown != nil {
		log.Error(errDown.Error())
	}
	log.Info("gracefully shutdown")
	return err

}
