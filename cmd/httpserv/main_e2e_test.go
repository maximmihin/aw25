package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/jackc/pgx/v5/pgconn"
	slogmulti "github.com/samber/slog-multi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	. "github.com/maximmihin/aw25/cmd/httpserv/internal/testclient"
	. "github.com/maximmihin/aw25/cmd/httpserv/internal/testdal"
	models "github.com/maximmihin/aw25/internal/dal/modelsgen"
)

// app config
const (
	DbUser = "merchstore_db_user"
	DbPass = "merchstore_db_pass"
	DbName = "merchstore_db_name"

	jwtPublicKey  = "access_secret"
	jwtPrivateKey = jwtPublicKey // TODO make async

	logLevel = "-4" // slog.LevelDebug

	logFile = "test.log.json"
)

// test defaults
const (
	defPass = "qwerty123"
)

// target server types
const (
	// run app in goroutine
	embed = iota
	// run app in a separate docker container
	docker
	// tests will run with an external app on url
	external
)

type TargetServer struct {
	Type int
	Url  string
}

func ParseServerConfig(t *testing.T, confStr string) TargetServer {
	t.Helper()
	tokens := strings.SplitN(confStr, ":", 2)
	serverType := tokens[0]

	switch serverType {
	case "embed", "":
		t.Log("tests will work with embed httpserv")
		return TargetServer{
			Type: embed,
		}
	case "docker":
		t.Log("tests will work with httpserv in docker")
		return TargetServer{
			Type: docker,
		}
	case "http", "https":
		_, err := url.Parse(confStr)
		require.NoError(t, err, "invalid extended url")

		t.Logf("tests will work with external httpserv on %s", confStr)
		return TargetServer{
			Type: external,
			Url:  confStr,
		}
	}
	t.Fatalf("unavailable server type: %s; available: \"empty\" or \"embed\" (default), \"docker\" and extended server (url)", serverType)
	return TargetServer{}
}

func TestE2E(t *testing.T) {

	//t.Setenv("E2E_TESTING_SERVER", "docker")
	//
	//t.Setenv("E2E_TESTING_SERVER", "http://localhost:8080")
	//t.Setenv("E2E_TESTING_DB", "postgres://merch_store_db_user:changeme@localhost:5432/merch_store_db_name?sslmode=disable")
	//
	//t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	var dockerBridgeDbConnStr string
	dbConnStr := os.Getenv("E2E_TESTING_DB")
	if dbConnStr == "" {
		dbConnStr, dockerBridgeDbConnStr = RunDockerPostgres(t)
	}
	PostgresPrepare(t, dbConnStr)
	tdb := NewTestDal(t, dbConnStr)

	serverConfig := ParseServerConfig(t, os.Getenv("E2E_TESTING_SERVER"))
	switch serverConfig.Type {
	case embed:
		serverConfig.Url = RunEmbedHttpServ(t, dbConnStr)
	case docker:
		serverConfig.Url = RunDockerHttpServ(t, dockerBridgeDbConnStr)
	case external:
		// already set in ParseServerConfig()
	}
	tcl := NewTestClient(t, serverConfig.Url)

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

		t.Run("green", func(t *testing.T) {
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

			u1, err := tdb.Queries.GetUserByName(context.TODO(), user1)
			require.NoError(t, err)
			require.Equal(t, models.User{
				Name:     user1,
				Password: defPass,
				Coins:    990,
			}, u1)

			u2, err := tdb.Queries.GetUserByName(context.TODO(), user2)
			require.NoError(t, err)
			require.Equal(t, models.User{
				Name:     user2,
				Password: defPass,
				Coins:    1010,
			}, u2)
		})

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

			req, err := http.NewRequest("POST", serverConfig.Url, nil)
			require.NoError(t, err)

			req.Header.Add("Authorization", "Bearer "+userToken)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			t.Cleanup(func() {
				assert.NoError(t, resp.Body.Close())
			})

			res10 := tcl.Info(t, InfoParams{JwtToken: userToken})
			require.Equal(t, 200, res10.StatusCode())
		})

	})

}

func RunDockerHttpServ(t *testing.T, dbConnStr string) string {
	httpSrvCtrName := fmt.Sprintf("merchStore_%s_HttpServ", NameTestInSnakeCase(t))

	pgConfig, err := pgconn.ParseConfig(dbConnStr)
	require.NoError(t, err)

	ctrReq := testcontainers.ContainerRequest{
		Name: httpSrvCtrName,

		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "../../",
			Dockerfile: "cmd/httpserv/Dockerfile",
			KeepImage:  true,
		},
		Env: map[string]string{
			"POSTGRES_USER":     pgConfig.User,
			"POSTGRES_PASSWORD": pgConfig.Password,
			"POSTGRES_DB":       pgConfig.Database,

			"POSTGRES_HOST": pgConfig.Host,
			"POSTGRES_PORT": strconv.Itoa(int(pgConfig.Port)),

			"JWT_PRIVATE_KEY": jwtPublicKey,
			"JWT_PUBLIC_KEY":  jwtPrivateKey,

			"HTTP_SERVICE_HOST": "0.0.0.0",
			"HTTP_SERVICE_PORT": "8080",

			"LOG_LEVEL": logLevel,
		},

		WaitingFor:   wait.ForLog("started listen"),
		ExposedPorts: []string{"8080/tcp"},
	}

	if pgConfig.Host == "localhost" {
		ctrReq.HostAccessPorts = []int{int(pgConfig.Port)}
	}

	container, err := testcontainers.GenericContainer(context.TODO(), testcontainers.GenericContainerRequest{
		ContainerRequest: ctrReq,
		Started:          true,
	})
	require.NoError(t, err)
	require.NotNil(t, container)

	p, err := container.MappedPort(context.TODO(), "8080/tcp")
	require.NoError(t, err)

	return fmt.Sprintf("http://localhost:%s", p.Port())
}

func RunEmbedHttpServ(t *testing.T, connStr string) string {

	logger := InitLogger(t)

	merchShopHttpServiceLogger = logger

	var runningPort string
	var fiberStartedListen = make(chan struct{}, 1)
	merchShopHttpServiceFiberOnListenFunc = func(listenData fiber.ListenData) error {
		defer close(fiberStartedListen)

		runningPort = listenData.Port

		return nil
	}

	cErr := make(chan error, 1)

	go func() {
		cErr <- Run(Config{
			PostgresConnString: connStr,

			HttpServiceHost: "localhost",
			HttpServicePort: "0", // choose random free port

			JwtPrivateKey: jwtPrivateKey,
			JwtPublicKey:  jwtPublicKey,

			LogLevel: logLevel,
		})
	}()
	select {
	case err := <-cErr:
		if err != nil {
			t.Fatal("fail to start app" + err.Error())
		}
	case <-fiberStartedListen:
		t.Log("fiber server started listen")
	}
	return fmt.Sprintf("http://localhost:%s", runningPort)
}

func handleFknPanicTestContainers(t *testing.T, some any) {
	if some == nil {
		return
	}
	err, ok := some.(error)
	if !ok {
		t.Fatal("testcontainers doesnt run for some reason: ", some)
	}
	if err.Error() == "rootless Docker not found" {
		t.Fatal("need running docker on host for run e2e tests - testcontainers will create and launch the necessary containers automatically")
	}
	t.Fatal("testcontainers fail to connect to docker on host: " + err.Error())
}

func RunDockerPostgres(t *testing.T) (localhostConnStr, dockerBridgeConnStr string) {
	ctx := t.Context()

	pgCtrName := fmt.Sprintf("merchStore_%s_Postgres", NameTestInSnakeCase(t))

	ctrRunOpts := []testcontainers.ContainerCustomizer{
		postgres.WithDatabase(DbName),
		postgres.WithUsername(DbUser),
		postgres.WithPassword(DbPass),

		withName(pgCtrName),

		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10 * time.Second)),
	}

	defer func() {
		handleFknPanicTestContainers(t, recover())
	}()
	t.Logf("starting %s...", pgCtrName)
	pgContainer, err := postgres.Run(ctx,
		"postgres:17.2-alpine3.21",
		ctrRunOpts...,
	)
	// TODO add cleanup terminate container
	require.NoError(t, err)
	t.Logf("started %s", pgCtrName)

	localhostConnStr, err = pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	ctrBridgeIP, err := pgContainer.ContainerIP(t.Context()) // not robust code
	// it is assumed that the container is running only in the default docker network (bridge)
	require.NoError(t, err)

	// create connString for internal network (docker bridge)
	publicPgUrl, err := url.Parse(localhostConnStr)
	require.NoError(t, err)
	internalPgUrl := *publicPgUrl
	internalPgUrl.Host = net.JoinHostPort(ctrBridgeIP, "5432")
	dockerBridgeConnStr = internalPgUrl.String()

	t.Logf("postgres available on: \n - %s\n - %s (default docker bridge network)",
		publicPgUrl.Host, internalPgUrl.Host)

	return localhostConnStr, dockerBridgeConnStr
}

func PostgresPrepare(t *testing.T, pgConnString string) {

	mg, err := migrate.New(
		"file://../../internal/dal/migrations",
		pgConnString,
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

func withName(ctrName string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Name = ctrName
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

func NameTestInSnakeCase(t *testing.T) string {
	return strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
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
