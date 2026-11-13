package model

import "time"

type Book struct {
	ID            int        `json:"id"`
	Title         string     `json:"title"`
	ISBN          string     `json:"isbn"`
	Author        string     `json:"author"`
	Category      string     `json:"category"`
	PublishedYear *int       `json:"published_year"`
	Description   *string    `json:"description"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
