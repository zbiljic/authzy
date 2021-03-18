package jsonmap

import (
	"encoding/json"
	"errors"
)

type JSONMap map[string]interface{}

func (j JSONMap) Value() (string, error) {
	data, err := json.Marshal(j)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (j JSONMap) Scan(src interface{}) error {
	var source []byte
	switch v := src.(type) {
	case string:
		source = []byte(v)
	case []byte:
		source = v
	default:
		return errors.New("invalid data type for JSONMap")
	}

	if len(source) == 0 {
		source = []byte("{}")
	}
	return json.Unmarshal(source, &j)
}
