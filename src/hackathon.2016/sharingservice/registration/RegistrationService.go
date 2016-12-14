package main

import (
	"fmt"
	"golang.org/x/net/context"
	"cloud.google.com/go/pubsub"
	"net/http"
	"log"
	"encoding/json"
	"time"
	"strings"
	"hackathon.2016/sharingservice/common"
)

//
// handles /add_item and /book_item endpoints
//
func main() {
	http.HandleFunc(common.REGISTRATION_URI, addItem)
	http.HandleFunc(common.BOOKING_URI, bookItem)
	log.Fatal(http.ListenAndServe(common.REGISTRATION_LISTEN_ADDRESS, nil))
}

func addItem(w http.ResponseWriter, r *http.Request) {
	var item common.ItemRegistration
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Printf("ERROR: Failed to parse request: %v", err)
	}
	w.Write([]byte(postItem(item)))
}

func postItem(item common.ItemRegistration) string {
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, common.PROJECT_ID)
	if err != nil {
		log.Printf("ERROR: Failed to create client: %v", err)
	}

	topic, err := client.CreateTopic(ctx, common.TOPIC_NAME)
	if err != nil {
		log.Printf("ERROR: Failed to create topic: %v", err)
	}

	msg, err := json.Marshal(item)
	if err != nil {
		log.Printf("ERROR: Failed to marshal JSON: %v", err)
	}

	randB := common.RandomStringBytes(16)
	res, err := topic.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
		Attributes: map[string]string {
			common.EVENT_TYPE_ATTRIBUTE_NAME:common.REGISTRATION_EVENT_TYPE,
			common.TIMESTAMP_ATTRIBUTE_NAME:fmt.Sprintf("%d", time.Now().UnixNano()),
			common.HASH_ATTRIBUTE_NAME:randB,
		},
	})

	if err != nil {
		return err.Error()
	} else {
		return strings.Join(res,",")
	}
}

func bookItem(w http.ResponseWriter, r *http.Request) {
	bookingInfo := common.BookingInfo{}
	err := json.NewDecoder(r.Body).Decode(&bookingInfo)
	log.Printf("booking info: TS=%d, %v, %v\n", bookingInfo.Timestamp, bookingInfo.Hash, bookingInfo.Email)
	if err != nil {
		log.Printf("ERROR: Failed to parse request: %v", err)
	}
	w.Write([]byte(postInfo(bookingInfo)))
}

func postInfo(bookingInfo common.BookingInfo) string {
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, common.PROJECT_ID)
	if err != nil {
		log.Printf("ERROR: Failed to create client: %v", err)
	}

	topic, err := client.CreateTopic(ctx, common.TOPIC_NAME)
	if err != nil {
		log.Printf("ERROR: Failed to create topic: %v", err)
	}

	data, err := json.Marshal(bookingInfo)
	if err != nil {
		log.Printf("ERROR: Failed to marshal booking info: %v", err)
	}

	res, err := topic.Publish(ctx, &pubsub.Message{
		Data:data,
		Attributes: map[string]string{
			common.EVENT_TYPE_ATTRIBUTE_NAME: common.BOOKING_EVENT_TYPE,
			common.TIMESTAMP_ATTRIBUTE_NAME:  fmt.Sprintf("%d", time.Now().UnixNano()),
		},
	})

	if err != nil {
		return err.Error()
	} else {
		return strings.Join(res, ",")
	}
}
