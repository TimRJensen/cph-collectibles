package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	//https://localhost/fs/assets/01KMH4PYR7S0XTS70WQA1G0M1H/0001.png
	mux.HandleFunc("/fs/assets/{id}/{file}", handler)
	http.ListenAndServe(":8081", WithCors(mux))
}
