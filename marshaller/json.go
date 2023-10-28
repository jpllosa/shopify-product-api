package marshaller

import "encoding/json"

// Marshall returns a byte representation of v
func Marshal(v any) []byte {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return jsonBytes
}

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by T
func Unmarshal[T any](data []byte) T {
	var result T

	err := json.Unmarshal(data, &result)
	if err != nil {
		panic(err)
	}

	return result
}
