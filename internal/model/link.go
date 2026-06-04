package model

import "time"

type Link struct {
	ID          string    `json:"id"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	Visits      int64     `json:"visits"`
}
