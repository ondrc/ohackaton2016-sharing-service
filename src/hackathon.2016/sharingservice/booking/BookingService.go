package main

import (
	"cloud.google.com/go/pubsub"
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"hackathon.2016/sharingservice/common"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	http.HandleFunc(common.BOOKING_URI, bookItem)
	log.Fatal(http.ListenAndServe(common.BOOKING_LISTEN_ADDRESS, nil))
}

func bookItem(w http.ResponseWriter, r *http.Request) {
	var bookingInfo common.BookingInfo
	err := json.NewDecoder(r.Body).Decode(&bookingInfo)
	if err != nil {
		fmt.Errorf("Failed to parse request: %v", err)
	}
	w.Write([]byte(postInfo(bookingInfo)))
}

func postInfo(bookingInfo common.BookingInfo) string {
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, common.PROJECT_ID)
	if err != nil {
		fmt.Errorf("Failed to create client: %v", err)
	}

	topic, err := client.CreateTopic(ctx, common.TOPIC_NAME)
	if err != nil {
		fmt.Errorf("Failed to create topic: %v", err)
	}

	data, err := json.Marshal(bookingInfo)
	if err != nil {
		fmt.Errorf("Failed to marshal booking info: %v", err)
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
