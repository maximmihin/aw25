package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	slogmulti "github.com/samber/slog-multi"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	. "github.com/maximmihin/aw25/cmd/httpserv/internal/testclient"
	"github.com/maximmihin/aw25/internal/repo"
	models "github.com/maximmihin/aw25/internal/repo/modelsgen"
)

const (
	DbUser = "merchstore_db_user"
	DbPass = "changeme"
	DbName = "merchstore_db_name"

	jwtPublicKey  = "access_secret"
	jwtPrivateKey = jwtPublicKey

	httpServiceHost = "localhost"
	httpServicePort = "8888"
	logLevel        = "-4" // slog.LevelDebug

	logFile = "test.log.json"
)

const (
	defPass = "qwerty123"
)

var globalTestModeIsDev = false

func TestE2E(t *testing.T) {

	require.NoError(t, os.Setenv("E2E_DEV_MODE", "true")) // TODO

	globalTestModeIsDev = os.Getenv("E2E_DEV_MODE") == "true"

	pgConnStr := RunPostgresContainer(t)
	PostgresPrepare(t, pgConnStr)

	if globalTestModeIsDev {
		UserHttpServerRun(t, pgConnStr)
	} else {
		// TODO
		// run in container
		t.Fatal("real e2e not implemented")
	}

	tcl := NewTestClient(t, httpServiceHost, httpServicePort)
	trepo, err := repo.New(context.TODO(), pgConnStr)
	require.NoError(t, err)

	t.Run("auth", func(t *testing.T) {
		t.Parallel()

		t.Run("green", func(t *testing.T) {
			t.Parallel()

			t.Run("new user", func(t *testing.T) {
				t.Parallel()

				res := tcl.Auth(t, AuthParams{
					Username: NewEmail(t),
					Password: defPass,
				})
				require.Equal(t, 200, res.StatusCode())

			})

			t.Run("repeated login", func(t *testing.T) {
				t.Parallel()

				mail := NewEmail(t)

				res := tcl.Auth(t, AuthParams{
					Username: mail,
					Password: defPass,
				})
				require.Equal(t, 200, res.StatusCode())

				res2 := tcl.Auth(t, AuthParams{
					Username: mail,
					Password: defPass,
				})
				require.Equal(t, 200, res2.StatusCode())

			})
		})

		t.Run("red", func(t *testing.T) {
			t.Parallel()

			t.Run("wrong pass", func(t *testing.T) {
				t.Parallel()

				mail := NewEmail(t)

				res := tcl.Auth(t, AuthParams{
					Username: mail,
					Password: defPass,
				})
				require.Equal(t, 200, res.StatusCode())

				res2 := tcl.Auth(t, AuthParams{
					Username: mail,
					Password: "someWrongPassword",
				})
				require.Equal(t, 401, res2.StatusCode())

			})
		})

	})

	t.Run("buy", func(t *testing.T) {
		t.Parallel()

		t.Run("green", func(t *testing.T) {
			t.Parallel()

			res := tcl.Auth(t, AuthParams{
				Password: defPass,
				Username: NewEmail(t),
			})
			require.Equal(t, 200, res.StatusCode())

			res2 := tcl.BuyMerch(t, BuyMerchParams{
				Auth:      *res.JSON200.Token,
				MerchItem: "t-shirt",
			})
			require.Equal(t, 200, res2.StatusCode())
		})

		t.Run("empty jwt", func(t *testing.T) {
			t.Parallel()

			res := tcl.Auth(t, AuthParams{
				Password: defPass,
				Username: NewEmail(t),
			})
			require.Equal(t, 200, res.StatusCode())

			res2 := tcl.BuyMerch(t, BuyMerchParams{
				Auth:      "",
				MerchItem: "t-shirt",
			})
			require.Equal(t, 401, res2.StatusCode())
		})

		t.Run("not enough coins", func(t *testing.T) {
			t.Parallel()

			userName := NewEmail(t)
			var userToken string

			res := tcl.Auth(t, AuthParams{
				Password: defPass,
				Username: userName,
			})
			require.Equal(t, 200, res.StatusCode())
			userToken = *res.JSON200.Token

			res2 := tcl.BuyMerch(t, BuyMerchParams{
				Auth:      userToken,
				MerchItem: "pink-hoody",
			})
			require.Equal(t, 200, res2.StatusCode())

			res3 := tcl.BuyMerch(t, BuyMerchParams{
				Auth:      userToken,
				MerchItem: "pink-hoody",
			})
			require.Equal(t, 200, res3.StatusCode())

			res4 := tcl.BuyMerch(t, BuyMerchParams{
				Auth:      userToken,
				MerchItem: "pink-hoody",
			})
			require.Equal(t, 400, res4.StatusCode())

		})

		t.Run("unavailable merch item", func(t *testing.T) {
			t.Parallel()

			res := tcl.Auth(t, AuthParams{
				Password: defPass,
				Username: NewEmail(t),
			})
			require.Equal(t, 200, res.StatusCode())

			res2 := tcl.BuyMerch(t, BuyMerchParams{
				Auth:      *res.JSON200.Token,
				MerchItem: "some-unavailable-merch-item",
			})
			require.Equal(t, 400, res2.StatusCode())
		})
	})

	t.Run("send coins", func(t *testing.T) { // TODO try send to themself
		t.Parallel()

		user1 := NewEmail(t, "one")
		user2 := NewEmail(t, "two")

		res := tcl.Auth(t, AuthParams{
			Username: user1,
			Password: defPass,
		})
		require.Equal(t, 200, res.StatusCode())

		res2 := tcl.Auth(t, AuthParams{
			Username: user2,
			Password: defPass,
		})
		require.Equal(t, 200, res2.StatusCode())

		res3 := tcl.SendCoins(t, SendCoinParams{
			Auth:   *res.JSON200.Token,
			Amount: 10,
			ToUser: user2,
		})
		require.Equal(t, 200, res3.StatusCode())

		u1, err := trepo.Queries.GetUserByName(context.TODO(), user1)
		require.NoError(t, err)
		require.Equal(t, models.User{
			Name:     user1,
			Password: defPass,
			Coins:    990,
		}, u1)

		u2, err := trepo.Queries.GetUserByName(context.TODO(), user2)
		require.NoError(t, err)
		require.Equal(t, models.User{
			Name:     user2,
			Password: defPass,
			Coins:    1010,
		}, u2)

	})

	t.Run("info", func(t *testing.T) {
		t.Parallel()

		t.Run("green", func(t *testing.T) {
			user1 := NewEmail(t, "user1")
			user2 := NewEmail(t, "user2")

			var user1AuthToken string
			var user2AuthToken string

			// register two user
			{
				res := tcl.Auth(t, AuthParams{
					Username: user1,
					Password: defPass,
				})
				require.Equal(t, 200, res.StatusCode())
				user1AuthToken = *res.JSON200.Token

				res2 := tcl.Auth(t, AuthParams{
					Username: user2,
					Password: defPass,
				})
				require.Equal(t, 200, res2.StatusCode())
				user2AuthToken = *res2.JSON200.Token
			}

			// transfer coins back and forth
			{
				res3 := tcl.SendCoins(t, SendCoinParams{
					Auth:   user1AuthToken,
					Amount: 100,
					ToUser: user2,
				})
				require.Equal(t, 200, res3.StatusCode())

				res4 := tcl.SendCoins(t, SendCoinParams{
					Auth:   user2AuthToken,
					Amount: 200,
					ToUser: user1,
				})
				require.Equal(t, 200, res4.StatusCode())

				res5 := tcl.SendCoins(t, SendCoinParams{
					Auth:   user1AuthToken,
					Amount: 300,
					ToUser: user2,
				})
				require.Equal(t, 200, res5.StatusCode())

				res6 := tcl.SendCoins(t, SendCoinParams{
					Auth:   user2AuthToken,
					Amount: 400,
					ToUser: user1,
				})
				require.Equal(t, 200, res6.StatusCode())
			}

			// user1 buy two t-shirt and one pen
			{
				res7 := tcl.BuyMerch(t, BuyMerchParams{
					Auth:      user1AuthToken,
					MerchItem: "t-shirt",
				})
				require.Equal(t, 200, res7.StatusCode())

				res8 := tcl.BuyMerch(t, BuyMerchParams{
					Auth:      user1AuthToken,
					MerchItem: "t-shirt",
				})
				require.Equal(t, 200, res8.StatusCode())

				res9 := tcl.BuyMerch(t, BuyMerchParams{
					Auth:      user1AuthToken,
					MerchItem: "pen",
				})
				require.Equal(t, 200, res9.StatusCode())
			}

			res10 := tcl.Info(t, InfoParams{JwtToken: user1AuthToken})
			require.Equal(t, 200, res10.StatusCode())
			require.JSONEq(t, fmt.Sprintf(`{
				"coins": 1030,
				"inventory": [
					{ "type": "pen", "quantity": 1 },
					{ "type": "t-shirt", "quantity": 2 }
				],
				"coinHistory": {
					"received": [
						{ "fromUser": "%[1]s", "amount": 200 },
						{ "fromUser": "%[1]s", "amount": 400 }
					],
					"sent": [
						{ "toUser": "%[1]s",	"amount": 100 },
						{ "toUser": "%[1]s",	"amount": 300 }
					]
				}
			}`, user2), string(res10.Body))
		})

		t.Run("clean user", func(t *testing.T) {

			user := NewEmail(t, "user")
			var userToken string

			res := tcl.Auth(t, AuthParams{
				Username: user,
				Password: defPass,
			})
			require.Equal(t, 200, res.StatusCode())
			userToken = *res.JSON200.Token

			res10 := tcl.Info(t, InfoParams{JwtToken: userToken})
			require.Equal(t, 200, res10.StatusCode())
		})

		t.Run("bad http method with good path and jwt", func(t *testing.T) {

			user := NewEmail(t, "user")
			var userToken string

			res := tcl.Auth(t, AuthParams{
				Username: user,
				Password: defPass,
			})
			require.Equal(t, 200, res.StatusCode())
			userToken = *res.JSON200.Token

			req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%s", httpServiceHost, httpServicePort), nil)
			require.NoError(t, err)

			req.Header.Add("Authorization", "Bearer "+userToken)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			t.Cleanup(func() {
				resp.Body.Close()
			})

			res10 := tcl.Info(t, InfoParams{JwtToken: userToken})
			require.Equal(t, 200, res10.StatusCode())
		})

	})

}

func UserHttpServerRun(t *testing.T, connStr string) {

	logger := InitLogger(t)

	merchShopHttpServiceLogger = logger

	var fiberStartedListen = make(chan struct{}, 1)
	merchShopHttpServiceFiberOnListenFunc = func(listenData fiber.ListenData) error {
		close(fiberStartedListen)
		return nil
	}

	cErr := make(chan error, 1)

	go func() {
		err := Run(Config{
			PostgresConnString: connStr,

			JwtPrivateKey: jwtPrivateKey,
			JwtPublicKey:  jwtPublicKey,

			HttpServiceHost: httpServiceHost,
			HttpServicePort: httpServicePort,

			LogLevel: logLevel,
		})
		if err != nil {
			cErr <- err
			return
		}
	}()
	select {
	case e := <-cErr:
		t.Fatal("fail to start app" + e.Error())
	case <-fiberStartedListen:
		t.Log("fiber server started listen")
	}
}

func RunPostgresContainer(t *testing.T) string {
	ctx := context.Background()

	pgCtrName := fmt.Sprintf("merchStore_%s_Postgres", NameTestInSnakeCase(t))

	ctrRunOpts := []testcontainers.ContainerCustomizer{
		postgres.WithDatabase(DbName),
		postgres.WithUsername(DbUser),
		postgres.WithPassword(DbPass),

		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10 * time.Second)),
	}

	if globalTestModeIsDev {
		require.NoError(t, os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true"))
		ctrRunOpts = append(ctrRunOpts, withStayAliveAs(pgCtrName))
	}

	pgContainer, err := postgres.Run(ctx,
		"postgres:17.2-alpine3.21",
		ctrRunOpts...,
	)
	require.NoError(t, err)
	t.Logf("started %s", pgCtrName)

	// TODO add info about container port in log
	// TODO add info was new container created or reuse old
	connStr := pgContainer.MustConnectionString(ctx, "sslmode=disable")
	t.Cleanup(func() {
		t.Logf("%sdon't forget: docker rm -f %s%s", yellow, pgCtrName, resetColor)
	})
	return connStr
}

func PostgresPrepare(t *testing.T, connStr string) {

	mg, err := migrate.New(
		"file://../../internal/repo/migrations",
		connStr,
	)
	require.NoError(t, err, "fail create migrate instance")

	err = mg.Down()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatal("fail migrate down: " + err.Error())
	}
	t.Log("postgres down old db")

	err = mg.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatal("fail migrate up: " + err.Error())
	}
	t.Log("postgres migration up")
}

var yellow = "\033[33m"
var resetColor = "\033[0m"

func NameTestInSnakeCase(t *testing.T) string {
	return strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
}

func withStayAliveAs(ctrName string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Name = ctrName
		req.Reuse = true
		return nil
	}
}

func createLogFile(t *testing.T) io.Writer {
	file, err := os.Create(logFile)
	require.NoError(t, err, "fail to create file "+logFile)

	t.Cleanup(func() {
		err = file.Close()
		if err != nil {
			t.Logf("fail to close file %s: %s", logFile, err.Error())
		}
	})
	return file
}

func InitLogger(t *testing.T) *slog.Logger {
	file := createLogFile(t)

	slogOpts := slog.HandlerOptions{
		Level: slog.LevelDebug,
		//AddSource: false,
	}

	tmp := slog.New(
		slogmulti.Fanout(
			//tint.NewHandler(os.Stdout, &tint.Options{
			//	Level: slog.LevelDebug,
			//}),
			slog.NewJSONHandler(os.Stdout, &slogOpts),
			slog.NewJSONHandler(file, &slogOpts)))

	return tmp
}

func NewEmail(t *testing.T, prefixes ...string) string {

	userName := NameTestInSnakeCase(t)
	if len(prefixes) > 0 {
		p := strings.Join(prefixes, "_")
		if p != "" {
			userName = fmt.Sprintf("%s_%s", p, userName)
		}
	}
	return userName + "@ya.ru"
}
