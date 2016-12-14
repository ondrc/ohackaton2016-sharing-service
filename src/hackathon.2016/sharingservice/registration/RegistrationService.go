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

func main() {
	http.HandleFunc(common.REGISTRATION_URI, addItem)
	log.Fatal(http.ListenAndServe(common.REGISTRATION_LISTEN_ADDRESS, nil))
}

func addItem(w http.ResponseWriter, r *http.Request) {
	var item common.ItemRegistration
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		fmt.Errorf("Failed to parse request: %v", err)
	}
	w.Write([]byte(postItem(item)))
}

func postItem(item common.ItemRegistration) string {
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, common.PROJECT_ID)
	if err != nil {
		fmt.Errorf("Failed to create client: %v", err)
	}

	topic, err := client.CreateTopic(ctx, common.TOPIC_NAME)
	if err != nil {
		fmt.Errorf("Failed to create topic: %v", err)
	}

	msg, err := json.Marshal(item)
	if err != nil {
		fmt.Errorf("Failed to marshal JSON: %v", err)
	}

	res, err := topic.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
		Attributes: map[string]string {
			common.EVENT_TYPE_ATTRIBUTE_NAME:common.REGISTRATION_EVENT_TYPE,
			common.TIMESTAMP_ATTRIBUTE_NAME:string(time.Now().Unix()),
			common.HASH_ATTRIBUTE_NAME:string(time.Now().Unix()),
		},
	})

	if err != nil {
		return err.Error()
	} else {
		return strings.Join(res,",")
	}
}