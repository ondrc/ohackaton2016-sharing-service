package main

import (
	"net/http"
	"encoding/json"
	"log"
)

type Response struct {
	Value string `json:"value"`
}

func createHandlerFunction(model *QueryModel) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		category := q.Get("category")
		location := q.Get("location")
		time := q.Get("time")
		duration := q.Get("duration")

		log.Printf("endpoint hit { category: %v, location: %v, time: %v, duration: %v }\n",
			category, location, time, duration)
		value := Response{Value: "hello world" }

		w.WriteHeader(200);
		json.NewEncoder(w).Encode(value);

		log.Println("endpoint hit -> response sent")
	}
}

func startServerAndBlock(model *QueryModel ) {
	http.HandleFunc("/query", createHandlerFunction(model))
	res := http.ListenAndServe(":" + GetEnvOr("PORT", "8080"), nil)
	log.Fatal(res)
}
