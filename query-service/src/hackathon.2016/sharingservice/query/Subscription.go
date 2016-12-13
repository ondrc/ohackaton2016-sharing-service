package main

import(
	"log"
	"math/rand"
	"cloud.google.com/go/pubsub"
	"golang.org/x/net/context"
	"time"
	"google.golang.org/api/iterator"
	"os"
)

const topicName = "events"
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const retryCount = 10
const eventBatchSize = 10
const interBatchDelayMs = 500

func Subscribe(ctx context.Context) *pubsub.Subscription {
	// create google cloud client
	var err error;
	var ex bool;
	projectId := MustGetEnv("PROJECT_ID")

	// Creates a client
	var client *pubsub.Client;
	client, err = pubsub.NewClient(ctx, projectId)
	if err != nil {
		log.Fatalf("Failed to create client: %v \n", err)
	}

	// Create or get topic
	var topic *pubsub.Topic;
	topic, err = client.CreateTopic(ctx, topicName)
	if err != nil {
		topic = client.Topic(topicName)
	}
	if topic == nil {
		log.Fatalf("Failed to create or obtain topic %v \n", topicName)
	}
	ex, err = topic.Exists(ctx)
	if err != nil {
		log.Fatalf("Error checking topic %v existence: %v \n", topicName, err)
	}
	if !ex {
		log.Fatalf("Topic does not exist on the server %v \n", topicName)
	}

	// create subscription with retry
	var sub *pubsub.Subscription;
	var subName string;
	for i:= 1; i <= retryCount; i++ {
		log.Printf("Attempt %v to create subscription\n", i)
		subName = GenerateSubscriptionName()
		sub, err = client.CreateSubscription(ctx, subName, topic, 20*time.Second, nil)
		if err == nil {
			break
		} else {
			log.Printf("Attemt to create subscription with name = %v failed: %v", subName, err)
		}
	}
	if err != nil {
		log.Fatalf("Failed to create subscription: %v \n", err);
	}

	ex, err = sub.Exists(ctx)
	if err != nil {
		log.Fatalf("Error checking subscription existence: %v \n", err)
	}
	if !ex {
		log.Fatalf("Created subscription does not exit. \n")
	}

	log.Printf("Created subcription %v", sub.ID())
	return sub
}

func UnSubscribe(ctx context.Context, sub *pubsub.Subscription) {
	name := sub.ID()
	err := sub.Delete(ctx);
	if err != nil {
		log.Fatalf("Failed to delete subscription: %v \n", err)
	}
	log.Println("Deleted subscription: ", name)
}

func StartEventReceiver(ctx context.Context, subscription *pubsub.Subscription, model *QueryModel) {
	go func() {
		for { // forever
			time.Sleep(interBatchDelayMs * time.Millisecond)
			ReadEventBatch(ctx, subscription, model)
		}
	}()
}

func ReadEventBatch(ctx context.Context, subscription *pubsub.Subscription, model *QueryModel) {
	it, err := subscription.Pull(ctx)
	if err != nil {
		log.Printf("Error pulling from message stream: %v \n", err)
		return
	}
	defer it.Stop()

	for i := 0; i < eventBatchSize; i++ {
		msg, err := it.Next()
		if err == iterator.Done {
			log.Printf("DEBUG: event iterator.Done \n") // TODO: remove
			return
		}
		if (err != nil) {
			log.Printf("Error reading message from iterator: %v \n", err)
		}
		ack := model.handle(msg)
		msg.Done(ack)
	}
}

func GenerateSubscriptionName() string {
	hostname, err := os.Hostname()
	if (err != nil) {
		log.Printf("Error obtaining hostname %v \n", err)
		hostname = "localhost"
	}
	name := "query-service-" + hostname + "-" + RandomStringBytes(8)
	log.Printf("DEBUG: subcription name = " + name)
	return name;
}

func RandomStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

//func IpString() string {
//	ipstring := ""
//	ifaces, err := net.Interfaces()
//	if (err != nil) {
//		return "-iferror"
//	}
//	for _, i := range ifaces {
//		addrs, err := i.Addrs()
//		if (err != nil) {
//			return "-iperror"
//		}
//		for _, addr := range addrs {
//			switch v := addr.(type) {
//				case *net.IPNet:
//					ipstring += ("-" + v.IP.String())
//				case *net.IPAddr:
//					ipstring += ("-" + v.IP.String())
//			}
//		}
//	}
//	return ipstring
//}
