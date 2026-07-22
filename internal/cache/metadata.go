package cache

import (
	"time"
)

type Metadata struct {
	RevalidationTime time.Time `json:"revalidation_time"`
	FileName         string    `json:"file_name"`
}

func (m *Metadata) GetTime() time.Time {
	return m.RevalidationTime
}
