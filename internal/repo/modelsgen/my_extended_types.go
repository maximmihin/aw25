package modelsgen

import (
	"bytes"
	"encoding/json"
)

type UserInventoryItem struct {
	Type     string `json:"type"`
	Quantity int64  `json:"quantity"`
}

func (u *UserInventoryItem) UnmarshalJSON(b []byte) error {

	dec := json.NewDecoder(bytes.NewReader(b))

	if err := SkipFirstArrayToken(dec); err != nil {
		return err
	}

	var merchType string
	var merchQuantity int64
	var err error

	if merchType, err = GetStringToken(dec); err != nil {
		return err
	}
	if merchQuantity, err = GetInt64Token(dec); err != nil {
		return err
	}
	if err = CheckLastArrayToken(dec); err != nil {
		return err
	}

	u.Type = merchType
	u.Quantity = merchQuantity
	return nil
}

type Receive struct {
	FromUser string `json:"fromUser"`
	Amount   int64  `json:"amount"`
}

func (r *Receive) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))

	if err := SkipFirstArrayToken(dec); err != nil {
		return err
	}

	var fromUser string
	var amount int64
	var err error

	if fromUser, err = GetStringToken(dec); err != nil {
		return err
	}
	if amount, err = GetInt64Token(dec); err != nil {
		return err
	}
	if err = CheckLastArrayToken(dec); err != nil {
		return err
	}

	r.FromUser = fromUser
	r.Amount = amount
	return nil
}

type Sent struct {
	ToUser string `json:"toUser"`
	Amount int64  `json:"amount"`
}

func (s *Sent) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))

	if err := SkipFirstArrayToken(dec); err != nil {
		return err
	}

	var toUser string
	var amount int64
	var err error

	if toUser, err = GetStringToken(dec); err != nil {
		return err
	}
	if amount, err = GetInt64Token(dec); err != nil {
		return err
	}
	if err = CheckLastArrayToken(dec); err != nil {
		return err
	}

	s.ToUser = toUser
	s.Amount = amount
	return nil
}

type CoinHistory struct {
	Received []Receive `json:"received"`
	Sent     []Sent    `json:"sent"`
}

type UserInfo struct {
	UserName          string              `json:"-"`
	Coins             int64               `json:"coins"`
	FullUserInventory []UserInventoryItem `json:"inventory"`
	CoinHistory       CoinHistory         `json:"coinHistory"`
}
