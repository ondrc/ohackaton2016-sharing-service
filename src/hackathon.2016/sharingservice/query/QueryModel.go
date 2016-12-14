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
	"fmt"
)

type ItemAvailability struct {
	Item *common.ItemRegistration
	Available bool
	Hash string
	Timestamp string
}




type QueryModel struct {
	CityCategoryItems map[string]map[string]*sortedset.SortedSet
	ItemsByKey map[string]*ItemAvailability
	DeferredBookings map[string]bool
	Lock sync.RWMutex
}

func NewQueryModel() *QueryModel {
	m := QueryModel{
		CityCategoryItems: make(map[string]map[string]*sortedset.SortedSet),
		ItemsByKey: make(map[string]*ItemAvailability),
		DeferredBookings: make(map[string]bool),
		Lock: sync.RWMutex{},
	}
	return &m;
}

// updates the model with given event and returns whether to ack the event
func (this *QueryModel) Handle(evt *pubsub.Message) bool {

	eventType := evt.Attributes[common.EVENT_TYPE_ATTRIBUTE_NAME]
	timestamp := evt.Attributes[common.TIMESTAMP_ATTRIBUTE_NAME]


	if (eventType == common.REGISTRATION_EVENT_TYPE) {
		hash := evt.Attributes[common.HASH_ATTRIBUTE_NAME]
		item := common.ItemRegistration{}
		err := json.Unmarshal(evt.Data, &item)
		if err != nil {
			log.Printf("ERROR: failed to unmarshal registration message data. Error: %v \nMessage: %v \n", err, evt)
		} else {
			this.handleRegistration(&item, timestamp, hash)
		}

	} else if (eventType == common.BOOKING_EVENT_TYPE) {
		item := common.BookingInfo{}
		err := json.Unmarshal(evt.Data, &item)
		if err != nil {
			log.Printf("ERROR: failed to unmarshal booking message data. Error: %v \nMessage: %v \n", err, evt)
		} else {
			this.handleBooking(&item, timestamp)
		}
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

	log.Printf("DEBUG: handle registration: (%v, %v) category: %v, location %v, from %v, to %v, by %v \n",
		timestamp, hash, item.What.Category, item.Where.Location, item.When.From, item.When.To, item.Who.Email)
	this.debugPrint()

	city := getCity(item)
	cat := getCategory(item)
	log.Printf("DEBUG: city = %v\n", city);

	category_items_map, present := this.CityCategoryItems[city]
	if !present {
		log.Printf("DEBUG: city %v not present - inserting\n", city);
		category_items_map = make(map[string] *sortedset.SortedSet)
		this.CityCategoryItems[city] = category_items_map
		log.Printf("DEBUG: city %v not present - inserted\n", city);
	}
	log.Printf("DEBUG: cat = %v\n", cat);
	itemSet, present := category_items_map[cat]
	if !present {
		log.Printf("DEBUG: cat %v not present - inserting\n", cat);
		itemSet = sortedset.New()
		category_items_map[cat] = itemSet
		log.Printf("DEBUG: cat %v not present - inserted\n", cat);
	}

	key := getItemKey(timestamp, hash)
	node := itemSet.GetByKey(key)
	if node == nil {
		log.Printf("DEBUG: key %v not present - inserting\n", key);
		itemAvail := ItemAvailability{ Item: item, Available: true, Hash: hash, Timestamp: timestamp }

		if this.DeferredBookings[key] {
			itemAvail.Available = false
		}
		delete(this.DeferredBookings, key)

		itemSet.AddOrUpdate(key, sortedset.SCORE(item.When.To), &itemAvail)
		this.ItemsByKey[key] = &itemAvail
	} else {
		log.Printf("DEBUG: key %v present - ignoring\n", key);
	}

	log.Printf("DEBUG: before outdated cleanup:")
	this.debugPrint()

	this.removeOutdatedItemsInSet(itemSet, getNow())

	log.Printf("DEBUG: after outdated cleanup:")
	this.debugPrint()
}

func (this *QueryModel) handleBooking(item *common.BookingInfo, timestamp string) {
	this.Lock.Lock()
	defer this.Lock.Unlock()

	iTs := fmt.Sprintf("%d", item.Timestamp)
	log.Printf("DEBUG: handle booking event: (%v, %v)", iTs, item.Hash)

	key := getItemKey(iTs, item.Hash)
	rItem, present := this.ItemsByKey[key]
	if !present {
		log.Printf("DEBUG: booking deferred - no such item yet (%v, %v)")
		this.DeferredBookings[key] = true
	} else {
		rItem.Available = false
		this.removeOutdatedItems(getCity(rItem.Item), getCategory(rItem.Item), getNow())
	}



}

func (this *QueryModel) Query(city string, category string, from, to int64, exactTime, includeBooked bool, skip, take int) []*ItemAvailability {
	this.Lock.RLock()
	defer this.Lock.RUnlock()

	theCity := strings.TrimSpace(strings.ToLower(city))
	theCategory := strings.TrimSpace(strings.ToLower(category))

	array := make([]*ItemAvailability, take)

	itemSet := this.CityCategoryItems[theCity][theCategory]
	if itemSet == nil {
		log.Printf("DEBUG: nil set for q(%v, %v)\n", city, category)
		itemSet = sortedset.New()
	} else {
		log.Printf("DEBUG: set not nil for q(%v, %v)\n", city, category)
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

			log.Printf("DEBUG: visiting node (exactTimeMatch=%v, inexactTimeMath=%v, timeMatch=%v, include=%v)\n",
				exactTimeMatch, inExactTimeMatch, timeMatch, include)

			if timeMatch && include {
				if (skip <= 0) {
					if taken < take {
						log.Printf("DEBUG: taking\n")
						array[taken] = itemAvail

						taken = taken + 1
						if taken == take {
							break;
							log.Printf("DEBUG: taken all - breaking\n")
						}
					}
				} else {
					log.Printf("DEBUG: skipping");
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
	log.Printf("Cleaning outdated entries\n")
	for {
		node := set.PeekMin()
		if node != nil {
			log.Printf("Cleaning outdated entries - non-empty\n")
			toValue := node.Value.(*ItemAvailability).Item.When.To
			log.Printf("Cleaning outdated entries - to = %v", toValue)
			if toValue < olderThan {
				log.Printf("Cleaning outdated entries - deleting\n")
				set.PopMin()
				delete(this.ItemsByKey, node.Key())
				log.Printf("Cleaning outdated entries - deleted\n")
			} else {
				log.Printf("Cleaning outdated entries - returning\n")
				return
			}
		} else {
			log.Printf("Cleaning outdated entries - returning\n")
			return
		}
	}
}

func getCity(item *common.ItemRegistration) string {
	return strings.TrimSpace(strings.ToLower(item.Where.Location))
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

func (this *QueryModel) debugPrint() {
	log.Printf("DEBUG: ====\n")
	defer log.Printf("DEBUG: ====\n")
	log.Printf("DEBUG: printing QueryModel structure\n")
	log.Printf("DEBUG: cities:\n")
	for k, v := range this.CityCategoryItems {
		log.Printf("DEBUG: --- %v\n", k)
		for k2, v2 := range v {
			log.Printf("DEBUG: ------ %v -> %v\n", k2, v2)
		}
	}
}
