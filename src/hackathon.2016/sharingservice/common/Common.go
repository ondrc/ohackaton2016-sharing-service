package common

import "time"

const EVENT_TYPE_ATTRIBUTE_NAME = "eventType"
const TIMESTAMP_ATTRIBUTE_NAME = "timestamp"
const HASH_ATTRIBUTE_NAME = "hash"
const EMAIL_ATTRIBUTE_NAME = "email"

const REGISTRATION_EVENT_TYPE = "registration"
const BOOKING_EVENT_TYPE = "booking"

const PROJECT_ID = "august-ascent-152314"
const TOPIC_NAME = "events"

const REGISTRATION_LISTEN_ADDRESS = ":8081"
const REGISTRATION_URI = "/add_item"

const QUERY_SERVICE_LISTEN_ADDRESS = "8080"
const QUERY_SERVICE_URI = "/query"

const BOOKING_LISTEN_ADDRESS = ":8083"
const BOOKING_URI = "/book_item"

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type ItemRegistration struct {
	What struct {
			Category string `json:"Category"`
			Description string `json:"Description"`
		}
	Where struct {
			Location string `json:"Location"`
		}
	When struct {
			From int64 `json:"From"`
			To int64 `json:"To"`
		}
	Who struct {
			Email string `json:"Email"`
	}
}

type BookingInfo struct {
	Timestamp int64
	Hash string
	Email string
}

func RandomStringBytes(n int) string {
	b := make([]byte, n)
	ts := time.Now().UnixNano()
	for i := range b {
		index := ts % int64(len(letterBytes))
		ts = ts / int64(len(letterBytes))
		b[n-i-1] = letterBytes[index]
	}
	return string(b)
}

