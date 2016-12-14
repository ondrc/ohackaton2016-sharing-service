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

	client, err := datastore.NewClient(ctx, common.PROJECT_ID)
	if err != nil {
		fmt.Errorf("Failed to create datastore client: %v", err)
	}

	common.StartEventReceiver(ctx, sub, func(msg *pubsub.Message) bool {
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