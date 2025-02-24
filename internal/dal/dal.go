package dal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	models "github.com/maximmihin/aw25/internal/dal/modelsgen"
)

const (
	pingRetry = 5
)

type ExtDBTX interface {
	models.DBTX

	Begin(ctx context.Context) (pgx.Tx, error)
}

type Dal struct {
	PgxPool ExtDBTX
	Queries *models.Queries
	// TODO add timeouts
}

func New(ctx context.Context, pool *pgxpool.Pool) (*Dal, error) {

	var err error
	for i := 1; i < pingRetry+1; i++ {
		err = pool.Ping(ctx)
		if err != nil {
			time.Sleep(time.Duration(i) * time.Second)
			continue
		}
		break
	}

	if err != nil {
		return nil, err
	}

	return &Dal{
		PgxPool: pool,
		Queries: models.New(pool),
	}, nil
}

func (r Dal) WithTx(tx pgx.Tx) *Dal {
	return &Dal{
		PgxPool: tx,
		Queries: r.Queries.WithTx(tx),
	}
}

var ErrInternal = errors.New("dal internal error")

func (r Dal) GetUserByName(ctx context.Context, name string) (*models.User, error) {
	user, err := r.Queries.GetUserByName(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: unexepted error: %w", ErrInternal, err)
	}
	return &user, nil
}

var ErrUserAlreadyExists = errors.New("user already exists")

func (r Dal) AddNewUser(ctx context.Context, createArgs models.CreateUserParams) (*models.User, error) {
	user, err := r.Queries.CreateUser(ctx, createArgs)
	if err == nil {
		return &user, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {

		switch pgErr.ConstraintName {
		case "users_coins_non_negative":
			return nil, fmt.Errorf("%w: invalid user to create: nums coin must be non negative: this constraint must check upper layer", ErrInternal)
		case "users_pkey":
			return nil, ErrUserAlreadyExists
		}
	}
	return nil, fmt.Errorf("%w: unexepted error: %w", ErrInternal, err)

}

var ErrNotEnoughCoins = errors.New("the user does not have enough coins")

func (r Dal) MinusCoins(ctx context.Context, userName string, amount int64) (*models.User, error) {
	user, err := r.Queries.MinusUserCoins(ctx, models.MinusUserCoinsParams{
		Amount: amount,
		Name:   userName,
	})
	if err == nil {
		return &user, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "users_coins_non_negative":
			return nil, ErrNotEnoughCoins
		}
	}
	return nil, fmt.Errorf("%w: unexepted error: %w", ErrInternal, err)
}

func (r Dal) PlusCoins(ctx context.Context, userName string, amount int64) (*models.User, error) {
	user, err := r.Queries.PlusUserCoins(ctx, models.PlusUserCoinsParams{
		Amount: amount,
		Name:   userName,
	})
	if err == nil {
		return &user, nil
	}
	return nil, fmt.Errorf("%w: unexepted error: %w", ErrInternal, err)
}

var ErrInvalidUser = errors.New("invalid user")
var ErrInvalidMerchItem = errors.New("invalid merch item")
var ErrUserMerchPairExist = errors.New("user - merch pair already exists")

func (r Dal) AddMerchToUser(ctx context.Context, userName string, merchName string) (*models.MerchOwnership, error) {
	merchOwn, err := r.Queries.AddMerchItem(ctx, models.AddMerchItemParams{
		UserName:  userName,
		MerchItem: merchName,
	})
	if err == nil {
		return &merchOwn, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "merch_ownership_fk_user_name":
			return nil, ErrInvalidUser
		case "merch_ownership_fk_merch_item":
			return nil, ErrInvalidMerchItem
		case "merch_ownership_pkey":
			return nil, ErrUserMerchPairExist
		}
	}
	return nil, fmt.Errorf("%w: unexepted error: %w", ErrInternal, err)
}

var ErrNonPositiveAmount = errors.New("transfer amount must be positive number")
var ErrInvalidSender = errors.New("invalid sender")
var ErrInvalidRecipient = errors.New("invalid recipient")

func (r Dal) CreateTransfer(ctx context.Context, sender, recipient string, amount int64) (*models.CoinTransfer, error) {
	transfer, err := r.Queries.CreateTransfer(ctx, models.CreateTransferParams{
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
	})

	if err == nil {
		return &transfer, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "coin_transfers_amount_positive_number":
			return nil, ErrNonPositiveAmount
		case "coin_transfers_fk_sender":
			return nil, ErrInvalidSender
		case "coin_transfers_fk_recipient":
			return nil, ErrInvalidRecipient
		}
	}
	return nil, fmt.Errorf("%w: unexepted error: %w", ErrInternal, err)
}

func (r Dal) GetCompositeUserInfo(ctx context.Context, userName string) (*models.UserInfo, error) {
	tmpInfo, err := r.Queries.GetCompositeUserIndo(ctx, userName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: unexepted error: %w", ErrInternal, err)
	}

	return &models.UserInfo{
		UserName:          tmpInfo.Name,
		Coins:             tmpInfo.Coins,
		FullUserInventory: NonNil(tmpInfo.Inventory),
		CoinHistory: models.CoinHistory{
			Received: NonNil(tmpInfo.Recived),
			Sent:     NonNil(tmpInfo.Sent),
		},
	}, nil
}

func NonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
