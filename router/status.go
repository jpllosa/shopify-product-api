package router

import (
	"net/http"
	"shopify-product-api/marshaller"
)

type MessageDTO struct {
	Id      int64  `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

func GetStatus(w http.ResponseWriter, r *http.Request) {
	body := MessageDTO{
		Id:      1,
		Message: "API ready and waiting",
	}

	jsonBytes := marshaller.Marshal(body)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(jsonBytes)
}
