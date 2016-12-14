package common

import "time"

const EVENT_TYPE_ATTRIBUTE_NAME = "eventType"
const TIMESTAMP_ATTRIBUTE_NAME = "timestamp"
const HASH_ATTRIBUTE_NAME = "hash"

const REGISTRATION_EVENT_TYPE = "registration"
const BOOKING_EVENT_TYPE = "booking"

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type ItemRegistration struct {
	What struct {
			Category string `json:"Category"`
			Description string `json:"Description"`
		}
	Where struct {
			From string `json:"From"`
			To string `json:"To"`
		}
	When struct {
			From int64 `json:"From"`
			To int64 `json:"To"`
		}
	Who struct {
			Email string `json:"Email"`
	}
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

