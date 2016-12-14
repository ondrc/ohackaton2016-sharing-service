package main

import (
	"cloud.google.com/go/pubsub"
	"log"
	"hackathon.2016/sharingservice/common"
	"encoding/json"
	"github.com/wangjia184/sortedset"
	"sync"
	"time"
	"strings"
)

type ItemAvailability struct {
	Item *common.ItemRegistration
	Available bool
	Hash string
	Timestamp string
}

type ItemsByCategory map[string] *sortedset.SortedSet // sorted set: key ~ see GetItemKey, score ~ When.To, value ~ *ItemAvailability
type ItemsByCityAndCategory map[string] ItemsByCategory

type QueryModel struct {
	CityCategoryItems ItemsByCityAndCategory
	ItemsByKey map[string]*ItemAvailability
	DeferredBookings map[string]bool
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
			this.handleRegistration(&item, timestamp, hash)
		}

	} else if (eventType == common.BOOKING_EVENT_TYPE) {
		this.handleBooking(timestamp, hash)
	}

	return true
}

// Handles registration event
// - Adds available item into the query model structure
// - Handles deferred bookings (registration & booking events observed out of order)
// - Removes outdated items in the slot
func (this *QueryModel) handleRegistration(item *common.ItemRegistration, timestamp, hash string) {
	this.Lock.Lock()
	defer this.Lock.Unlock()

	log.Printf("DEBUG: handle registration: (%v, %v) category: %v, location %v, from %v, to %v \n",
		timestamp, hash, item.What.Category, item.Where.From, item.When.From, item.When.To)
	city := getCity(item)
	category_items_map, present := this.CityCategoryItems[city]
	if !present {
		category_items_map = make(ItemsByCategory)
		this.CityCategoryItems[city] = category_items_map
	}
	cat := getCategory(item)
	itemSet, present := category_items_map[cat]
	if !present {
		itemSet = sortedset.New()
		category_items_map[cat] = itemSet
	}

	key := getItemKey(timestamp, hash)
	node := itemSet.GetByKey(key)
	if node == nil {
		itemAvail := ItemAvailability{ Item: item, Available: true, Hash: hash, Timestamp: timestamp }

		if this.DeferredBookings[key] {
			itemAvail.Available = false
		}
		delete(this.DeferredBookings, key)

		itemSet.AddOrUpdate(key, sortedset.SCORE(item.When.To), &itemAvail)
		this.ItemsByKey[key] = &itemAvail
	}

	this.removeOutdatedItemsInSet(itemSet, getNow())
}

func (this *QueryModel) handleBooking(timestamp, hash string) {
	this.Lock.Lock()
	defer this.Lock.Unlock()

	log.Printf("DEBUG: handle booking event: (%v, %v)")

	key := getItemKey(timestamp, hash)
	item, present := this.ItemsByKey[key]
	if !present {
		log.Printf("DEBUG: booking deferred - no such item yet (%v, %v)")
		this.DeferredBookings[key] = true
	} else {
		item.Available = false
		this.removeOutdatedItems(getCity(item.Item), getCategory(item.Item), getNow())
	}



}

func (this *QueryModel) Query(city string, category string, from, to int64, exactTime, includeBooked bool, skip, take int) []*ItemAvailability {
	this.Lock.RLock()
	defer this.Lock.RUnlock()

	theCity := strings.TrimSpace(strings.ToLower(city))
	theCategory := strings.TrimSpace(strings.ToLower(category))
	array := make([]*ItemAvailability, take)

	var itemSet *sortedset.SortedSet = nil
	itemsByCat := this.CityCategoryItems[theCity]
	if itemsByCat != nil {
		itemSet = itemsByCat[theCategory]
	}
	if itemSet == nil {
		itemSet = sortedset.New()
	}

	now := getNow()
	if to < now {
		log.Printf("DEBUG: query - taken = 0 as querying the past\n")
		return array[:0];
	}

	nodes := itemSet.GetByRankRange(1, -1, false)
	taken := 0
	for _, n := range nodes {
		if n != nil {
			itemAvail := n.Value.(*ItemAvailability)
			exactTimeMatch := itemAvail.Item.When.From == from && itemAvail.Item.When.To == to
			inExactTimeMatch := itemAvail.Item.When.From <= from && itemAvail.Item.When.To >= to
			timeMatch :=  (exactTime && exactTimeMatch) || (!exactTime && inExactTimeMatch)
			include := includeBooked || itemAvail.Available

			if timeMatch && include {
				if (skip <= 0) {
					if taken < take {
						array[taken] = itemAvail
						taken = taken + 1
					}
				} else {
					skip = skip - 1
				}
			}
		}
	}

	log.Printf("DEBUG: query - taken = %v \n", taken)
	return array[:taken]
}

func (this *QueryModel) cleanupOldDeferredBookings() {
	now := getNow()
	for k := range this.DeferredBookings {
		expired := isDeferedBookingKeyExpired(k, now)
		if expired {
			delete(this.DeferredBookings, k)
		}
	}
}

func (this *QueryModel) removeOutdatedItems(city, category string, olderThan int64) {
	catMap, present := this.CityCategoryItems[city]
	if present {
		set, present := catMap[category]
		if present {
			this.removeOutdatedItemsInSet(set, olderThan)
		}
	}
}


func (this *QueryModel) removeOutdatedItemsInSet(set *sortedset.SortedSet, olderThan int64) {
	for {
		node := set.PeekMin()
		if node != nil {
			toValue := node.Value.(common.ItemRegistration).When.To
			if toValue < olderThan {
				set.PopMin()
				delete(this.ItemsByKey, node.Key())
			} else {
				break
			}
		} else {
			break
		}
	}
}

func getCity(item *common.ItemRegistration) string {
	return strings.TrimSpace(strings.ToLower(item.Where.From))
}

func getCategory(item *common.ItemRegistration) string {
	return strings.TrimSpace(strings.ToLower(item.What.Category))
}

func getItemKey(timestamp, hash string) string {
	return timestamp + ":" + hash
}

func getNow() int64 {
	return time.Now().Unix()
}

func isDeferedBookingKeyExpired(key string, now int64) bool {
	// TODO: decode timestamp and say if it is expired
	//       i.e. way older than now
	return false
}
