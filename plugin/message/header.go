package message

// Header is appended to the top of every message
type Header struct {
	ID      string `json:"id"`      // Application level identifier for the message, intended to be a UUID
	Version string `json:"version"` // version of messages
	Type    string `json:"type"`    // message type
}
