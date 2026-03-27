package v1

import (
	"encoding/json"
	"log"
	"net/http"
)

func Success[T any](w http.ResponseWriter, code int, data ...T) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(&struct {
		Code  int  `json:"code"`
		Error bool `json:"error"`
		Data  []T  `json:"data"`
	}{
		Code:  code,
		Error: false,
		Data:  data,
	})
}

func Error(w http.ResponseWriter, code int, err error) {
	log.Panicln(err)
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(&struct {
		Code  int    `json:"code"`
		Error bool   `json:"error"`
		Msg   string `json:"msg"`
	}{
		Code:  code,
		Error: true,
		Msg:   err.Error(),
	})
}
