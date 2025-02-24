-- name: GetUserByName :one
SELECT * FROM users
WHERE name = @name
;

-- name: CreateUser :one
INSERT INTO users (name, password, coins)
VALUES (@name, @password, @coins)
RETURNING *
;

-- name: MinusUserCoins :one
UPDATE users
SET
    coins = coins - @amount
WHERE name = @name
RETURNING *
;

-- name: PlusUserCoins :one
UPDATE users
SET
    coins = coins + @amount
WHERE name = @name
RETURNING *
;

-- name: CreateTransfer :one
INSERT INTO coin_transfers (sender, recipient, amount)
VALUES (@sender, @recipient, @amount)
RETURNING *
;

-- name: AddMerchItem :one
INSERT INTO merch_ownership (user_name, merch_item, quantity)
VALUES (@user_name, @merch_item, 1)
ON CONFLICT (user_name, merch_item) DO UPDATE
    SET quantity = merch_ownership.quantity + 1
RETURNING *
;

-- name: GetCompositeUserIndo :one
WITH
    user_name_inventory AS (
        SELECT
            user_name,
            json_agg(
                    json_build_array(merch_item, quantity)
                ) as merch_item_quantity -- key_value json format
        FROM merch_ownership
        WHERE user_name = @ft_user_name
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
        WHERE recipient = @ft_user_name
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
        WHERE sender = @ft_user_name
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
WHERE name = @ft_user_name
;