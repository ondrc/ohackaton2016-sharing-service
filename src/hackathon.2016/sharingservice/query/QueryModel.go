package main

import (
	"cloud.google.com/go/pubsub"
	"log"
	"hackathon.2016/sharingservice/common"
	"encoding/json"
	"github.com/wangjia184/sortedset"
	"sync"
	"time"
)

type ItemAvailability struct {
	Item *common.ItemRegistration
	Available bool
	Hash string
	Timestamp string
}

type ItemsByCategory map[string] *sortedset.SortedSet // sorted set: key ~ see GetItemKey, score ~ When.To, value ~ ItemAvailability
type ItemsByCityAndCategory map[string] ItemsByCategory

type QueryModel struct {
	CityCategoryItems ItemsByCityAndCategory
	DeferredBookings sortedset.SortedSet // key ~ see GetItemKey, score ~ timestamp, value ~ _
	Lock sync.RWMutex
}



// updates the model with given event and returns whether to ack the event
func (this *QueryModel) Handle(evt *pubsub.Message) bool {

	eventType := evt.Attributes[common.EVENT_TYPE_ATTRIBUTE_NAME]
	timestamp := evt.Attributes[common.TIMESTAMP_ATTRIBUTE_NAME]
	hash := evt.Attributes[common.HASH_ATTRIBUTE_NAME]


	if (eventType == common.REGISTRATION_EVENT_TYPE) {
		item := common.ItemRegistration{}
		err := json.Unmarshal(evt.Data, &item)
		if err != nil {
			log.Printf("ERROR: failed to unmarshal registration message data. Error: %v \nMessage: %v \n", err, evt)
		} else {
			this.HandleRegistration(&item, hash, timestamp)
		}

	} else if (eventType == common.BOOKING_EVENT_TYPE) {
		// TODO
	}

	return true
}

// Handles registration event
// - Adds available item into the query model structure
// - Handles deferred bookings (registration & booking events observed out of order)
// - Removes outdated items in the slot
func (this *QueryModel) HandleRegistration(item *common.ItemRegistration, hash, timestamp string) {
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

	key := GetItemKey(hash, timestamp)
	node := itemSet.GetByKey(key)
	if node == nil {
		itemAvail := ItemAvailability{ Item: item, Available: true}

		deferredBooking := this.DeferredBookings.Remove(key)
		if deferredBooking != nil {
			itemAvail.Available = false
		}

		itemSet.AddOrUpdate(key, sortedset.SCORE(item.When.To), itemAvail)
	}

	RemoveOutdatedItems(itemSet, GetNow())
}

func (this *QueryModel) Query(city string, category string, from, to int64, take int) []ItemAvailability {
	this.Lock.RLock()
	defer this.Lock.RUnlock()

	array := make([]ItemAvailability, take)
	//now := GetNow()
	var itemSet *sortedset.SortedSet = nil
	itemsByCat := this.CityCategoryItems[city]
	if itemsByCat != nil {
		itemSet = itemsByCat[category]
	}
	if itemSet == nil {
		itemSet = sortedset.New()
	}

	//for itemSet.



	// TODO

	return array
}

func RemoveOutdatedItems(set *sortedset.SortedSet, olderThan int64) {
	for {
		node := set.PeekMin()
		if node != nil {
			toValue := int64(node.Value.(common.ItemRegistration).When.To)
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

func GetCity(item *common.ItemRegistration) string {
	return item.Where.From
}

func GetCategory(item *common.ItemRegistration) string {
	return item.What.Category
}

func GetItemKey(id, timestamp string) string {
	return timestamp + ":" + id
}

func GetNow() int64 {
	return time.Now().Unix()
}
