package modelsgen

import (
	"encoding/json"
	"fmt"
)

func SkipFirstArrayToken(dec *json.Decoder) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}
	if td, ok := t.(json.Delim); !ok || td != json.Delim('[') {
		return fmt.Errorf("exetped [ as first json token")
	}
	return nil
}

func CheckLastArrayToken(dec *json.Decoder) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}
	if td, ok := t.(json.Delim); !ok || td != json.Delim(']') {
		return fmt.Errorf("exetped ] as last json token")
	}
	return nil
}

func GetStringToken(dec *json.Decoder) (string, error) {
	t, err := dec.Token()
	if err != nil {
		return "", err
	}

	var tp string
	tp, ok := t.(string)
	if !ok {
		return "", fmt.Errorf("invalid type of inventory (must be string)")
	}
	return tp, nil
}

func GetNumericToken(dec *json.Decoder) (float64, error) {
	t, err := dec.Token()
	if err != nil {
		return 0, err
	}

	fQuantity, ok := t.(float64)
	if !ok {
		return 0, fmt.Errorf("invalid quantity of inventory (must be int64)")
	}
	return fQuantity, nil
}

func GetInt64Token(dec *json.Decoder) (int64, error) {
	tmp, err := GetNumericToken(dec)
	if err != nil {
		return 0, err
	}
	return int64(tmp), nil
}
