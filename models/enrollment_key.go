package models

import (
	"time"
)

const (
	Undefined KeyType = iota
	TimeExpiration
	Uses
	Unlimited
)

// KeyType - the type of enrollment key
type KeyType int

// String - returns the string representation of a KeyType
func (k KeyType) String() string {
	return [...]string{"Undefined", "TimeExpiration", "Uses", "Unlimited"}[k]
}

// EnrollmentToken - the tokenized version of an enrollmentkey;
// to be used for host registration
type EnrollmentToken struct {
	Server string `json:"server"`
	Value  string `json:"value"`
}

// EnrollmentKeyLength - the length of an enrollment key - 62^16 unique possibilities
const EnrollmentKeyLength = 32

// EnrollmentKey - the key used to register hosts and join them to specific networks
type EnrollmentKey struct {
	Expiration    time.Time `json:"expiration"`
	UsesRemaining int       `json:"uses_remaining"`
	Value         string    `json:"value"`
	Networks      []string  `json:"networks"`
	Unlimited     bool      `json:"unlimited"`
	Tags          []string  `json:"tags"`
	Token         string    `json:"token,omitempty"` // B64 value of EnrollmentToken
	Type          KeyType   `json:"type"`
}

// APIEnrollmentKey - used to create enrollment keys via API
type APIEnrollmentKey struct {
	Expiration    int64    `json:"expiration"`
	UsesRemaining int      `json:"uses_remaining"`
	Networks      []string `json:"networks"`
	Unlimited     bool     `json:"unlimited"`
	Tags          []string `json:"tags"`
	Type          KeyType  `json:"type"`
}

// RegisterResponse - the response to a successful enrollment register
type RegisterResponse struct {
	ServerConf    ServerConfig `json:"server_config"`
	RequestedHost Host         `json:"requested_host"`
}

// EnrollmentKey.IsValid - checks if the key is still valid to use
func (k *EnrollmentKey) IsValid() bool {
	if k == nil {
		return false
	}
	if k.UsesRemaining > 0 {
		return true
	}
	if !k.Expiration.IsZero() && time.Now().Before(k.Expiration) {
		return true
	}
	if k.Type == Undefined {
		return false
	}

	return k.Unlimited
}

// EnrollmentKey.Validate - validate's an EnrollmentKey
// should be used during creation
func (k *EnrollmentKey) Validate() bool {
	return k.Networks != nil &&
		k.Tags != nil &&
		len(k.Value) == EnrollmentKeyLength &&
		k.IsValid()
}
