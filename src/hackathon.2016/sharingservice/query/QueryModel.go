package main

import (
	"cloud.google.com/go/pubsub"
	"log"
	"hackathon.2016/sharingservice/common"
	"encoding/json"
	"github.com/wangjia184/sortedset"
	"sync"
	"bytes"
	"time"
)

type ItemsByCategory map[string] *sortedset.SortedSet
type ItemsByCityAndCategory map[string] ItemsByCategory

type QueryModel struct {
	CityCategoryItems ItemsByCityAndCategory
	DeferredBookings sortedset.SortedSet
	Lock sync.RWMutex
}

// updates the model with given event and returns whether to ack the event
func (this *QueryModel) Handle(evt *pubsub.Message) bool {

	eventType := evt.Attributes[common.EVENT_TYPE_ATTRIBUTE_NAME]
	timestamp := evt.Attributes[common.TIMESTAMP_ATTRIBUTE_NAME]


	if (eventType == common.REGISTRATION_EVENT_TYPE) {
		item := common.Item{}
		err := json.Unmarshal(evt.Data, &item)
		if err != nil {
			log.Printf("ERROR: failed to unmarshal registration message data. Error: %v \nMessage: %v \n", err, evt)
		} else {
			this.HandleRegistration(&item, timestamp)
		}

	}

	return true
}

func (this *QueryModel) HandleRegistration(item *common.Item, timestamp string) {
	this.Lock.Lock()
	defer this.Lock.Unlock()

	log.Printf("DEBUG: handle registration: category: %v, location %v, from %v, to %v \n", item.What.Category, item.Where.From, item.When.From, item.When.To)
	city := GetCity(item)
	category_items_map, present := this.CityCategoryItems[city]
	if !present {
		category_items_map = make(ItemsByCategory)
		this.CityCategoryItems[city] = category_items_map
	}
	cat := GetCategory(item)
	itemSet, present := category_items_map[cat]
	if !present {
		itemSet = sortedset.New()
		category_items_map[cat] = itemSet
	}

	itemSet.AddOrUpdate(GetItemKey(item, timestamp), sortedset.SCORE(item.When.To), item)

	// TODO: handle deferred bookings

	RemoveOutdatedItems(itemSet, time.Now().Unix())
}

func (this *QueryModel) Query(city string, category string, from, to int64) {
	this.Lock.RLock()
	defer this.Lock.RUnlock()
	// TODO
}

func RemoveOutdatedItems(set *sortedset.SortedSet, olderThan int64) {
	for {
		node := set.PeekMin()
		if node != nil {
			toValue := int64(node.Value.(common.Item).When.To)
			if toValue < olderThan {
				set.PopMin()
			} else {
				break
			}
		} else {
			break
		}
	}
}

func GetCity(item *common.Item) string {
	return item.Where.From
}

func GetCategory(item *common.Item) string {
	return item.What.Category
}

func GetItemKey(item *common.Item, timestamp string) string {
	jsonString := ""
	jsonBytes, err := json.Marshal(item)
	if err != nil {
		log.Fatalf("Can not marshal item %v \n", item)
	}
	n := bytes.IndexByte(jsonBytes, 0)
	if n > 0 {
		jsonString = string(jsonBytes[:n])
	}
	return timestamp + ":" + jsonString
}