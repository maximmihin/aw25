CREATE TABLE users
(
    name               varchar(255) PRIMARY KEY,
    password varchar(255) NOT NULL, -- todo store hash
    coins BIGINT NOT NULL,

    CONSTRAINT users_coins_non_negative CHECK (coins >= 0)
);

CREATE TABLE merch
(
    slug varchar(255) PRIMARY KEY,
    price bigint NOT NULL ,

    CONSTRAINT merch_price_non_negative CHECK (price >= 0)
);

CREATE TABLE coin_transfers
(
    sender varchar(255) NOT NULL,
    recipient varchar(255) NOT NULL,
    amount bigint NOT NULL,

    CONSTRAINT coin_transfers_fk_sender
        FOREIGN KEY (sender)
            REFERENCES users,

    CONSTRAINT coin_transfers_fk_recipient
        FOREIGN KEY (recipient)
            REFERENCES users,

    CONSTRAINT coin_transfers_amount_positive_number CHECK (amount > 0),
    CONSTRAINT coin_transfers_sender_is_recipient CHECK (sender != recipient)
);

CREATE TABLE merch_ownership
(
    user_name varchar(255) NOT NULL,
    merch_item varchar(255) NOT NULL,
    quantity bigint NOT NULL,

    CONSTRAINT merch_ownership_fk_user_name
        FOREIGN KEY (user_name)
            REFERENCES users,

    CONSTRAINT merch_ownership_fk_merch_item
        FOREIGN KEY (merch_item)
            REFERENCES merch,

    PRIMARY KEY (user_name, merch_item)
);

-- todo remove to initialize container
INSERT INTO merch (slug, price)
VALUES ('t-shirt', 80),
       ('cup', 20),
       ('book', 50),
       ('pen', 10),
       ('powerbank', 200),
       ('hoody', 300),
       ('umbrella', 200),
       ('socks', 10),
       ('wallet', 50),
       ('pink-hoody', 500)
;
