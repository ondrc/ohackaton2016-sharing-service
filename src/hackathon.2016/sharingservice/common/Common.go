package common

const EVENT_TYPE_ATTRIBUTE_NAME = "eventType"
const TIMESTAMP_ATTRIBUTE_NAME = "timestamp"
const REGISTRATION_EVENT_TYPE = "registration"

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
			From int32 `json:"From"`
			To int32 `json:"To"`
		}
}

