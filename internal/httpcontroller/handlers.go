package httpcontroller

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	models "github.com/maximmihin/aw25/internal/dal/modelsgen"
	"log/slog"

	"github.com/golang-jwt/jwt/v5"

	"github.com/maximmihin/aw25/internal/dal"
)

type Handlers struct {
	Dal    *dal.Dal
	Logger *slog.Logger

	JWTPrivateKey string
}

const (
	welcomeBonusCoins = 1000
)

var MerchShowCase = map[string]int64{ // TODO hide in dal with cash
	"t-shirt":    80,
	"cup":        20,
	"book":       50,
	"pen":        10,
	"powerbank":  200,
	"hoody":      300,
	"umbrella":   200,
	"socks":      10,
	"wallet":     50,
	"pink-hoody": 500,
}

func newJWT(userName string, JWTPrivateKey string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject: userName,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWTPrivateKey))
}

func resAuth(c *fiber.Ctx, jwt string) error {
	c.Response().Header.Set("Content-Type", "application/json")
	c.Status(200)

	return c.JSON(&AuthResponse{
		Token: &jwt,
	})
}

func resUserInfo(c *fiber.Ctx, info models.UserInfo) error { // TODO use view  models
	c.Response().Header.Set("Content-Type", "application/json")
	c.Status(200)

	return c.JSON(&info)
}

var ErrJWTWasNotSet = errors.New("jwt was not set or set on unexpected fiber local name")
var ErrJWTWrongType = errors.New("jwt wrong type")
var ErrJWTFailedToExtractSubject = errors.New("fail to extract subject from jwt")
var ErrJWTEmptySubject = errors.New("empty subject")

func ExtractUserNameFromJwt(c *fiber.Ctx) (string, error) {
	jwtt := c.Locals("user")
	if jwtt == nil {
		return "", ErrJWTWasNotSet
	}
	strJwt, ok := jwtt.(*jwt.Token)
	if !ok {
		return "", ErrJWTWrongType
	}
	subj, err := strJwt.Claims.GetSubject()
	if err != nil {
		return "", ErrJWTFailedToExtractSubject
	}
	if subj == "" {
		return "", ErrJWTEmptySubject
	}
	return subj, nil
}

func (r Handlers) Auth(c *fiber.Ctx) error { // TODO add bcrypt check

	ctx := c.Context()

	log := r.Logger.With(slog.String("handler_name", "Auth"))

	var req AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(400, "invalid body: "+err.Error())
	}

	if errValidate := req.Validate(log); errValidate != nil {
		return errValidate
	}

	user, err := r.Dal.GetUserByName(ctx, req.Username)
	if err != nil {
		log.Error("fail to GetUserByName: " + err.Error())
		return err
	}

	if user != nil {
		if user.Password != req.Password {
			return fiber.NewError(401, "wrong password")
		}
		jwtt, err := newJWT(user.Name, r.JWTPrivateKey)
		if err != nil {
			log.Error("fail to create new JWT: " + err.Error())
			return err
		}
		return resAuth(c, jwtt)
	}

	user, err = r.Dal.AddNewUser(ctx, models.CreateUserParams{
		Name:     req.Username,
		Password: req.Password,
		Coins:    welcomeBonusCoins,
	})
	if err != nil {
		log.Error("fail to add new user AddNewUser: " + err.Error())
		return err
	}
	jwtt, err := newJWT(user.Name, r.JWTPrivateKey)
	if err != nil {
		log.Error("fail to create new JWT: " + err.Error())
		return err
	}
	return resAuth(c, jwtt)
}

func (r Handlers) BuyMerch(c *fiber.Ctx) error {

	log := r.Logger.With(slog.String("handler_name", "BuyMerch"))

	merchItem := c.Params("item")
	if merchItem == "" {
		return fiber.NewError(400, dal.ErrInvalidMerchItem.Error())
	}
	merchCost, ok := MerchShowCase[merchItem]
	if !ok {
		return fiber.NewError(400, "invalid merch item") // TODO return enum
	}

	userName, err := ExtractUserNameFromJwt(c)
	if err != nil {
		log.Error("error via extract name from jwt token: " + err.Error())
		return err
	}

	ctx := c.Context()

	tx, err := r.Dal.PgxPool.Begin(ctx)
	if err != nil {
		log.Error("error via start db transaction: " + err.Error())
		return err
	}
	defer func() {
		if err = tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Error("error via rollback tx: " + err.Error())
		}
	}()

	txRepo := r.Dal.WithTx(tx)

	_, err = txRepo.MinusCoins(ctx, userName, merchCost)
	if err != nil {
		if errors.Is(err, dal.ErrNotEnoughCoins) {
			return fiber.NewError(400, err.Error())
		}
		return fmt.Errorf("fail txRepo.MinusCoins: %w", err)
	}

	_, err = txRepo.AddMerchToUser(ctx, userName, merchItem)
	if err != nil {
		if errors.Is(err, dal.ErrInvalidUser) {
			return fiber.NewError(400, "user was deleted")
		}
		if errors.Is(err, dal.ErrInvalidMerchItem) {
			log.Error("fail via AddMerchToUser: " + err.Error()) // this must be checked in code before
			return fiber.NewError(400, err.Error())              // TODO return enum
		}
		if errors.Is(err, dal.ErrUserMerchPairExist) {
			log.Error("fail via AddMerchToUser: " +
				"this method should never have allowed this limitation to be constraint " +
				err.Error())
		}
		return fmt.Errorf("fail txRepo.AddMerchToUser: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		log.Error("fail tx.Commit: " + err.Error())
		return err
	}

	return nil
}

func (r Handlers) CoinTransfer(c *fiber.Ctx) error {
	log := r.Logger.With(slog.String("handler_name", "CoinTransfer"))

	var req SendCoinRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(400, "invalid request body")
	}

	userName, err := ExtractUserNameFromJwt(c)
	if err != nil {
		log.Error("error via extract name from jwt token: " + err.Error())
		return err
	}

	ctx := c.Context()

	tx, err := r.Dal.PgxPool.Begin(ctx)
	if err != nil {
		log.Error("fail to start tx: " + err.Error())
		return err
	}
	defer func() {
		if err = tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Error("failed to close tx: " + err.Error())
		}
	}()

	txRepo := r.Dal.WithTx(tx)

	_, err = txRepo.CreateTransfer(ctx, userName, req.ToUser, int64(req.Amount))
	if err != nil {
		if errors.Is(err, dal.ErrInternal) {
			log.Error("fail to create transfer: " + err.Error())
			return err
		}
		return fiber.NewError(400, err.Error())
	}

	_, err = txRepo.MinusCoins(ctx, userName, int64(req.Amount))
	if err != nil {
		if errors.Is(err, dal.ErrNotEnoughCoins) {
			return fiber.NewError(400, err.Error())
		}
		return fmt.Errorf("fail txRepo.MinusCoins: %w", err)
	}

	if _, err = txRepo.PlusCoins(ctx, req.ToUser, int64(req.Amount)); err != nil {
		log.Error("fail to add coins: " + err.Error())
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Error("fail to commit tx: " + err.Error())
	}
	return nil
}

func (r Handlers) Info(c *fiber.Ctx) error {
	log := r.Logger.With(slog.String("handler_name", "Info"))

	userName, err := ExtractUserNameFromJwt(c)
	if err != nil {
		log.Error("error via extract name from jwt token: " + err.Error())
		return err
	}

	ctx := c.Context()

	userInfo, err := r.Dal.GetCompositeUserInfo(ctx, userName)
	if err != nil {
		return err
	}
	if userInfo == nil {
		return fmt.Errorf("jwt token valid, but user not found in db")
	}

	return resUserInfo(c, *userInfo)
}
