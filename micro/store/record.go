package store

import (
	"time"
)

// Record is a data interface
// in eagle/storage we have implemented version
// but the code looks not pure, so we try to
// refactor those code in the future.
type Record interface {
	GetID() string // TODO: maybe we can add version at here with version id.
	GetData() []byte
	GetValue() interface{}
	GetExpiry() time.Duration // life of this record, implement auto deleted
	GetCreateAt() time.Time
	GetUpdateAt() time.Time
	GetDeleteAt() time.Time
	// SetID(string)
	// SetData([]byte)
	// SetValue(interface{})
	// SetExpiry(time.Duration)
	// SetDeleteAt(time.Time)
	// SetUpdateAt(time.Time)
	// SetCreateAt(time.Time)
}

// SimpleRecord is a simple data item
type SimpleRecord struct {
	ID       string        `json:"id"`
	Data     []byte        `json:"-"` // removed?
	Value    interface{}   `json:"value"`
	Expiry   time.Duration `json:"expory"`
	CreateAt time.Time     `json:"create_at"`
	UpdateAt time.Time     `json:"update_at"`
	DeleteAt time.Time     `json:"delete_at"`
}

// GetID returns id
func (sr *SimpleRecord) GetID() string {
	return sr.ID
}

// GetData returns marshalled data
func (sr *SimpleRecord) GetData() []byte {
	return sr.Data
}

// GetValue returns original value
func (sr *SimpleRecord) GetValue() interface{} {
	return sr.Value
}

// GetExpiry returns ...
func (sr *SimpleRecord) GetExpiry() time.Duration {
	return sr.Expiry
}

// GetCreateAt returns ...
func (sr *SimpleRecord) GetCreateAt() time.Time {
	return sr.CreateAt

}

// GetUpdateAt returns ...
func (sr *SimpleRecord) GetUpdateAt() time.Time {
	return sr.UpdateAt
}

// GetDeleteAt returns ...
func (sr *SimpleRecord) GetDeleteAt() time.Time {
	return sr.DeleteAt
}

// NewSimpleRecord create a simple record
func NewSimpleRecord(id string, v interface{}) Record {
	return &SimpleRecord{
		ID:       id,
		Value:    v,
		CreateAt: time.Now(),
	}
}

// IsExpired returns if the record expired
func IsExpired(r Record) bool {
	return r.GetCreateAt().Add(r.GetExpiry()).Before(time.Now())
}

// IsDeleted returns is the record deleted
func IsDeleted(r Record) bool {
	return !r.GetDeleteAt().IsZero()
}
