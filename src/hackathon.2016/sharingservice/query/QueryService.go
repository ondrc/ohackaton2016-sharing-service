package main

import (
	"net/http"
	"encoding/json"
	"log"
	"hackathon.2016/sharingservice/common"
	"strconv"
	"fmt"
)

const PARAM_CATEGORY = "category"
const PARAM_LOCATION = "location"
const PARAM_TIME_FROM = "timeFrom"
const PARAM_TIME_TO = "timeTo"

type Response struct {
	Items []*ItemAvailability
}

type ErrorResponse struct {
	Error string
	Hint string
}

func createHandlerFunction(model *QueryModel) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		w.Header().Add("Content-Type", "application/json")

		category := q.Get(PARAM_CATEGORY)
		location := q.Get(PARAM_LOCATION)
		timeFromString := q.Get(PARAM_TIME_FROM)
		timeToString := q.Get(PARAM_TIME_TO)

		log.Printf("DEBUG: endpoint hit { %v: %v, %v: %v, %v: %v, %v: %v }\n",
			PARAM_CATEGORY, category,
			PARAM_LOCATION, location,
			PARAM_TIME_FROM, timeFromString,
			PARAM_TIME_TO, timeToString)

		if category == "" || location == "" || timeFromString == "" || timeToString == "" {
			w.WriteHeader(400)
			resp := ErrorResponse{
				Error: fmt.Sprintf("No parameter values of at least one of parameters: %v, %v, %v, %v", PARAM_CATEGORY, PARAM_LOCATION, PARAM_TIME_FROM, PARAM_TIME_TO),
				Hint: fmt.Sprintf("Mandatory parameters: " +
				 	"%v (string, e.g. Car)," +
					"%v (string, e.g. London)," +
					"%v , %v (timestamps - milliseconds since 1.1.1970)",
					PARAM_CATEGORY, PARAM_LOCATION, PARAM_TIME_FROM, PARAM_TIME_TO),
			}
			v, e := json.Marshal(resp)
			if e != nil {
				log.Printf("Error marshalling response!\n")
			} else {
				w.Write(v)
			}
			return
		}

		timeFrom, err1 := strconv.ParseInt(timeFromString, 10, 64)
		timeTo, err2 := strconv.ParseInt(timeToString, 10, 64)

		if err1 != nil || err2 != nil {
			w.WriteHeader(400)
			resp := ErrorResponse{
				Error: fmt.Sprintf("Invalid format of at least one of parameters: %v, %v", PARAM_TIME_FROM, PARAM_TIME_TO),
				Hint: fmt.Sprintf("The parameters %v and %v are expected to be timestamps in UNIX format (miliseconds since 1.1.1970).", PARAM_TIME_FROM, PARAM_TIME_TO),
			}
			v, e := json.Marshal(resp)
			if e != nil {
				log.Printf("Error marshalling response!\n")
			} else {
				w.Write(v)
			}
			return
		}

		items:= model.Query(location, category, timeFrom, timeTo, false, false, 0, 200)

		resp := Response{ Items: items }
		v, e := json.Marshal(resp)
		if e != nil {
			log.Printf("Error marshalling response!\n")
		}
		w.Write(v)
		log.Printf("DEBUG: endpoint hit -> response sent\n")
	}
}

func StartServerAndBlock(model *QueryModel ) {
	http.HandleFunc(common.QUERY_SERVICE_URI, createHandlerFunction(model))
	res := http.ListenAndServe(":" + GetEnvOr("PORT", common.QUERY_SERVICE_LISTEN_ADDRESS), nil)
	log.Fatal(res)
}
