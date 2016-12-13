package main

import (
	"cloud.google.com/go/pubsub"
	"log"
)

type QueryModel struct {
	// some fields
}

func (this *QueryModel) handle(event *pubsub.Message) bool { // return whether to ack the message
	log.Printf("Updating model with event %v \n", event)
	return true // we shall very rarely return false (because if repeated, it may lead to endless loops)
}
