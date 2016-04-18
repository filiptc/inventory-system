package models

import "time"

type Item struct {
	Label      string
	Expiration time.Time
	Type       string
}

func NewItem(label string, expiration time.Time, typ string) *Item {
	return &Item{label, expiration, typ}
}

func (i *Item) HasExpired(t time.Time) bool {
	return t.After(i.Expiration)
}
