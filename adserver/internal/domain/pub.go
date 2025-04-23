package domain

import (
	"time"

	"github.com/google/uuid"
)

type Pub struct {
	ID          uuid.UUID `bson:"_id,omitempty" json:"id"`
	Title       string    `bson:"title" json:"title"`
	Description *string    `bson:"description,omitempty" json:"description,omitempty"`
	URL         string    `bson:"url" json:"url"`
	ExpiresAt   time.Time `bson:"expires_at" json:"expires_at"`
	Impressions int64     `bson:"impressions" json:"impressions"`
}
