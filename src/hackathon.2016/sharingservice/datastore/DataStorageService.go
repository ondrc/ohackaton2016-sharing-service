package main

import (
	"fmt"
	"golang.org/x/net/context"
	"cloud.google.com/go/pubsub"
	"log"
	"hackathon.2016/sharingservice/common"
	"cloud.google.com/go/datastore"
)

func main() {
	ctx := context.Background()

	sub := common.SubscribeFixed(ctx, common.DATA_SUBSCRIPTION_NAME)

	fmt.Println("Creating new client")
	client, err := datastore.NewClient(ctx, common.PROJECT_ID)
	fmt.Println("Client creation returned")
	if err != nil {
		fmt.Printf("Failed to create datastore client: %v\n", err)
	}

	fmt.Println("Starting receiver")
	common.StartEventReceiver(ctx, sub, func(msg *pubsub.Message) bool {
		fmt.Printf("Action %v\n", msg)
		return storeEvent(ctx, client, msg)
	})
}

func storeEvent(ctx context.Context, client *datastore.Client, msg *pubsub.Message) bool {
	attrs := msg.Attributes
	data := common.StorageData{
		Attributes:attrs,
		Data:string(msg.Data),
	}
	key := datastore.NameKey(common.DATA_KIND, attrs[common.TIMESTAMP_ATTRIBUTE_NAME], nil)

	_, err := client.Put(ctx, key, data)
	if err != nil {
		log.Printf("Failed to save task: %v", err)
		return false;
	} else {
		fmt.Printf("Saved %v to the datastore", key.String())
		return true;
	}
}