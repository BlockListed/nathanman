package model

import (
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Entry struct {
	ID   string `gorm:"primaryKey;not null;unique"`
	Name string
}

type Config struct {
	Bot      string
	Lasttime time.Time
}

func New(userId, name string) Entry {
	id, _ := gonanoid.New()
	return Entry{
		ID:   id,
		Name: name,
	}
}
