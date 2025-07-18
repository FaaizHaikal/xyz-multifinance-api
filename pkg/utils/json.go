package utils

import "encoding/json"

func MarshalJSON(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func UnmarshalJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}
