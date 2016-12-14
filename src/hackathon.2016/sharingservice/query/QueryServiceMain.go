package main

import (
	"os"
	"os/signal"
	"syscall"
	"log"
	"golang.org/x/net/context"
	"hackathon.2016/sharingservice/common"
	"cloud.google.com/go/pubsub"
)

//
// handles /query endpoint
//
func main() {

	// background context
	ctx := context.Background()
	// subscribe to pub sub
	sub := common.Subscribe(ctx)

	// set up signal handlers to delete subscription on termination
	log.Println("Installing signal handlers to unsubscribe on exit...")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT) // SIGHUP? SIGSTOP?
	go func () {
		for sig := range sigCh {
			log.Println("Signal received: ", sig)
			common.UnSubscribe(ctx, sub)
			os.Exit(0)
		}
	}()

	model := NewQueryModel()
	// TODO: re-play old events to reconstruct state

	// TODO: keep receiving messages to update state
	common.StartEventReceiver(ctx, sub, func (msg *pubsub.Message) bool {
		return model.Handle(msg)
	});

	// start query server
	StartServerAndBlock(model)
}
