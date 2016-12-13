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
)

const PROJECT_ID = "august-ascent-152314"
const TOPIC_NAME = "events"
const LISTEN_ADDRESS = ":8081"

type Event struct {
	Registration Item `json:"Registration"`
}

type Item struct {
	What struct {
		Category string `json:"Category"`
		Description string `json:"Description"`
     	}
	Where struct {
		From string `json:"From"`
		To string `json:"To"`
      	}
	When struct {
		From int32 `json:"From"`
		To int32 `json:"To"`
     	}
}

func main() {
	http.HandleFunc("/add_item", addItem)
	log.Fatal(http.ListenAndServe(LISTEN_ADDRESS, nil))
}

func addItem(w http.ResponseWriter, r *http.Request) {
	var item Item
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		fmt.Errorf("Failed to parse request: %v", err)
	}
	w.Write([]byte(postItem(item)))
}

func postItem(item Item) string {
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, PROJECT_ID)
	if err != nil {
		fmt.Errorf("Failed to create client: %v", err)
	}

	topic, err := client.CreateTopic(ctx, TOPIC_NAME)
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
			"eventType":"registration",
			"timestamp":string(time.Now().Unix()),
		},
	})

	if err != nil {
		return err.Error()
	} else {
		return strings.Join(res,",")
	}
}