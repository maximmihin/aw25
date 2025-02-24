package dal

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/jackc/pgx/v5/pgxpool"
	models "github.com/maximmihin/aw25/internal/dal/modelsgen"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"net"
	"net/url"
	"strings"
	"testing"
	"time"
)

const (
	DbUser = "user_db_user"
	DbPass = "changeme"
	DbName = "user_db_name"
)

const (
	defPass  = "qwerty123"
	defCoins = 1000
)

func TestIntegration(t *testing.T) {

	pgConnStr, _ := RunDockerPostgres(t)
	PostgresPrepare(t, pgConnStr)

	ctx := t.Context()

	pgxPool, err := pgxpool.New(ctx, pgConnStr)
	require.NoError(t, err)
	t.Cleanup(pgxPool.Close)

	repo, err := New(ctx, pgxPool)
	require.NoError(t, err)

	t.Run("AddNewUser", func(t *testing.T) {
		t.Parallel()

		t.Run("double", func(t *testing.T) {
			t.Parallel()

			userName := NewEmail(t)

			user, err := repo.AddNewUser(ctx, models.CreateUserParams{
				Name:     userName,
				Password: defPass,
				Coins:    defCoins,
			})
			require.NoError(t, err)
			require.NotNil(t, user)

			sameUser, err := repo.AddNewUser(ctx, models.CreateUserParams{
				Name:     userName,
				Password: defPass,
				Coins:    defCoins,
			})
			require.ErrorIs(t, err, ErrUserAlreadyExists)
			require.Nil(t, sameUser)
		})

	})

	t.Run("Buy merch - minus coin add merch WITHOUT tx", func(t *testing.T) {
		t.Parallel()

		userName := NewEmail(t)
		merchPen := models.Merch{
			Slug:  "pen",
			Price: 10,
		}

		user, err := repo.AddNewUser(ctx, models.CreateUserParams{
			Name:     userName,
			Password: defPass,
			Coins:    defCoins,
		})
		require.NoError(t, err)
		require.NotNil(t, user)

		user, err = repo.MinusCoins(ctx, userName, merchPen.Price)
		require.NoError(t, err)
		require.NotNil(t, user)
		require.Equal(t, defCoins-merchPen.Price, user.Coins)

		userMerch, err := repo.AddMerchToUser(ctx, userName, merchPen.Slug)
		require.NoError(t, err)
		require.Equal(t, &models.MerchOwnership{
			UserName:  userName,
			MerchItem: merchPen.Slug,
			Quantity:  1,
		}, userMerch)

	})

	t.Run("Buy merch - minus coin plus merch WITH tx", func(t *testing.T) {
		t.Parallel()

		userName := NewEmail(t)
		merchPen := models.Merch{
			Slug:  "pen",
			Price: 10,
		}

		user, err := repo.AddNewUser(ctx, models.CreateUserParams{
			Name:     userName,
			Password: defPass,
			Coins:    defCoins,
		})
		require.NoError(t, err)
		require.NotNil(t, user)

		tx, err := repo.PgxPool.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		txRepo := repo.WithTx(tx)

		user, err = txRepo.MinusCoins(ctx, userName, merchPen.Price)
		require.NoError(t, err)
		require.NotNil(t, user)
		require.Equal(t, defCoins-merchPen.Price, user.Coins)

		userMerch, err := txRepo.AddMerchToUser(ctx, userName, merchPen.Slug)
		require.NoError(t, err)
		require.Equal(t, &models.MerchOwnership{
			UserName:  userName,
			MerchItem: merchPen.Slug,
			Quantity:  1,
		}, userMerch)

		require.NoError(t, tx.Commit(ctx))

	})

	t.Run("get info", func(t *testing.T) {
		userName := "user1_teste2e_info@ya.ru"

		q := `
INSERT INTO users (name, password, coins) VALUES ('user2_teste2e_info@ya.ru', 'qwerty123', 800);
INSERT INTO users (name, password, coins) VALUES ('user1_teste2e_info@ya.ru', 'qwerty123', 1030);

INSERT INTO merch_ownership (user_name, merch_item, quantity) VALUES ('user1_teste2e_info@ya.ru', 't-shirt', 2);
INSERT INTO merch_ownership (user_name, merch_item, quantity) VALUES ('user1_teste2e_info@ya.ru', 'pen', 1);

INSERT INTO coin_transfers (sender, recipient, amount) VALUES ('user1_teste2e_info@ya.ru', 'user2_teste2e_info@ya.ru', 100);
INSERT INTO coin_transfers (sender, recipient, amount) VALUES ('user2_teste2e_info@ya.ru', 'user1_teste2e_info@ya.ru', 200);
INSERT INTO coin_transfers (sender, recipient, amount) VALUES ('user1_teste2e_info@ya.ru', 'user2_teste2e_info@ya.ru', 300);
INSERT INTO coin_transfers (sender, recipient, amount) VALUES ('user2_teste2e_info@ya.ru', 'user1_teste2e_info@ya.ru', 400);

`
		_, err := repo.PgxPool.Exec(ctx, q)
		require.NoError(t, err)

		ui, err := repo.GetCompositeUserInfo(ctx, userName)
		require.NoError(t, err)
		require.NotNil(t, ui) // todo add good check

	})

	t.Run("get info with empty inventory", func(t *testing.T) {
		userName := "user11_teste2e_info@ya.ru"

		q := `
INSERT INTO users (name, password, coins) VALUES ('user22_teste2e_info@ya.ru', 'qwerty123', 800);
INSERT INTO users (name, password, coins) VALUES ('user11_teste2e_info@ya.ru', 'qwerty123', 1030);

-- INSERT INTO merch_ownership (user_name, merch_item, quantity) VALUES ('user11_teste2e_info@ya.ru', 't-shirt', 2);
-- INSERT INTO merch_ownership (user_name, merch_item, quantity) VALUES ('user11_teste2e_info@ya.ru', 'pen', 1);

INSERT INTO coin_transfers (sender, recipient, amount) VALUES ('user11_teste2e_info@ya.ru', 'user22_teste2e_info@ya.ru', 100);
INSERT INTO coin_transfers (sender, recipient, amount) VALUES ('user22_teste2e_info@ya.ru', 'user11_teste2e_info@ya.ru', 200);
INSERT INTO coin_transfers (sender, recipient, amount) VALUES ('user11_teste2e_info@ya.ru', 'user22_teste2e_info@ya.ru', 300);
INSERT INTO coin_transfers (sender, recipient, amount) VALUES ('user22_teste2e_info@ya.ru', 'user11_teste2e_info@ya.ru', 400);

`
		_, err := repo.PgxPool.Exec(ctx, q)
		require.NoError(t, err)

		ui, err := repo.GetCompositeUserInfo(ctx, userName)
		require.NoError(t, err)
		require.NotNil(t, ui) // todo add good check
	})

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

func withName(ctrName string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Name = ctrName
		return nil
	}
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

func PostgresPrepare(t *testing.T, connStr string) {

	mg, err := migrate.New(
		"file://migrations",
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

var yellow = "\033[33m"
var resetColor = "\033[0m"

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
