// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: queries.sql

package modelsgen

import (
	"context"
)

const addMerchItem = `-- name: AddMerchItem :one
INSERT INTO merch_ownership (user_name, merch_item, quantity)
VALUES ($1, $2, 1)
ON CONFLICT (user_name, merch_item) DO UPDATE
    SET quantity = merch_ownership.quantity + 1
RETURNING user_name, merch_item, quantity
`

type AddMerchItemParams struct {
	UserName  string
	MerchItem string
}

func (q *Queries) AddMerchItem(ctx context.Context, arg AddMerchItemParams) (MerchOwnership, error) {
	row := q.db.QueryRow(ctx, addMerchItem, arg.UserName, arg.MerchItem)
	var i MerchOwnership
	err := row.Scan(&i.UserName, &i.MerchItem, &i.Quantity)
	return i, err
}

const createTransfer = `-- name: CreateTransfer :one
INSERT INTO coin_transfers (sender, recipient, amount)
VALUES ($1, $2, $3)
RETURNING sender, recipient, amount
`

type CreateTransferParams struct {
	Sender    string
	Recipient string
	Amount    int64
}

func (q *Queries) CreateTransfer(ctx context.Context, arg CreateTransferParams) (CoinTransfer, error) {
	row := q.db.QueryRow(ctx, createTransfer, arg.Sender, arg.Recipient, arg.Amount)
	var i CoinTransfer
	err := row.Scan(&i.Sender, &i.Recipient, &i.Amount)
	return i, err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (name, password, coins)
VALUES ($1, $2, $3)
RETURNING name, password, coins
`

type CreateUserParams struct {
	Name     string
	Password string
	Coins    int64
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser, arg.Name, arg.Password, arg.Coins)
	var i User
	err := row.Scan(&i.Name, &i.Password, &i.Coins)
	return i, err
}

const getCompositeUserIndo = `-- name: GetCompositeUserIndo :one
WITH
    user_name_inventory AS (
        SELECT
            user_name,
            json_agg(
                    json_build_array(merch_item, quantity)
                ) as merch_item_quantity -- key_value json format
        FROM merch_ownership
        WHERE user_name = $1
        GROUP BY user_name
    ),

    -- received
    user_name_recipient AS (
        SELECT
            recipient as user_name,
            json_agg(
                    json_build_array(sender, amount)
                ) as sender_amount -- key_value json format
        FROM coin_transfers
        WHERE recipient = $1
        GROUP BY recipient
    ),
    -- sent
    user_name_sender AS (
        SELECT
            sender as user_name,
            json_agg(
                    json_build_array(recipient, amount)
                ) as recipient_amount -- key_value json format
        FROM coin_transfers
        WHERE sender = $1
        GROUP BY sender
    )

SELECT
    name,
    coins,
    merch_item_quantity as inventory,
    recipient_amount as sent,
    sender_amount as recived
FROM users
         LEFT JOIN user_name_inventory
                   ON users.name = user_name_inventory.user_name
         LEFT JOIN user_name_sender
                   ON users.name = user_name_sender.user_name
         LEFT JOIN user_name_recipient
                   ON users.name = user_name_recipient.user_name
WHERE name = $1
`

type GetCompositeUserIndoRow struct {
	Name      string
	Coins     int64
	Inventory []UserInventoryItem
	Sent      []Sent
	Recived   []Receive
}

func (q *Queries) GetCompositeUserIndo(ctx context.Context, ftUserName string) (GetCompositeUserIndoRow, error) {
	row := q.db.QueryRow(ctx, getCompositeUserIndo, ftUserName)
	var i GetCompositeUserIndoRow
	err := row.Scan(
		&i.Name,
		&i.Coins,
		&i.Inventory,
		&i.Sent,
		&i.Recived,
	)
	return i, err
}

const getUserByName = `-- name: GetUserByName :one
SELECT name, password, coins FROM users
WHERE name = $1
`

func (q *Queries) GetUserByName(ctx context.Context, name string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByName, name)
	var i User
	err := row.Scan(&i.Name, &i.Password, &i.Coins)
	return i, err
}

const minusUserCoins = `-- name: MinusUserCoins :one
UPDATE users
SET
    coins = coins - $1
WHERE name = $2
RETURNING name, password, coins
`

type MinusUserCoinsParams struct {
	Amount int64
	Name   string
}

func (q *Queries) MinusUserCoins(ctx context.Context, arg MinusUserCoinsParams) (User, error) {
	row := q.db.QueryRow(ctx, minusUserCoins, arg.Amount, arg.Name)
	var i User
	err := row.Scan(&i.Name, &i.Password, &i.Coins)
	return i, err
}

const plusUserCoins = `-- name: PlusUserCoins :one
UPDATE users
SET
    coins = coins + $1
WHERE name = $2
RETURNING name, password, coins
`

type PlusUserCoinsParams struct {
	Amount int64
	Name   string
}

func (q *Queries) PlusUserCoins(ctx context.Context, arg PlusUserCoinsParams) (User, error) {
	row := q.db.QueryRow(ctx, plusUserCoins, arg.Amount, arg.Name)
	var i User
	err := row.Scan(&i.Name, &i.Password, &i.Coins)
	return i, err
}
