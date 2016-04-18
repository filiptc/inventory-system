package models

import (
	"fmt"
	"time"
)

const ItemNotFoundError = "No item with label %s exists in inventory"
const DuplicateItemError = "Item with label %s exists in inventory"
const ExpirationInPastError = "Expiration may not be set in the past"

type Inventory map[string]*Item

func NewInventory() Inventory {
	return Inventory{}
}

func (inv Inventory) Add(i *Item) error {
	if _, ok := inv[i.Label]; ok {
		return fmt.Errorf(DuplicateItemError, i.Label)
	}

	if i.HasExpired(time.Now()) {
		return fmt.Errorf(ExpirationInPastError)
	}

	inv[i.Label] = i
	return nil
}

func (inv Inventory) GetByLabel(label string) (*Item, error) {
	item, ok := inv[label]
	if !ok {
		return nil, fmt.Errorf(ItemNotFoundError, label)
	}
	return item, nil
}

func (inv Inventory) RemoveByLabel(label string) error {
	if _, err := inv.GetByLabel(label); err != nil {
		return err
	}

	delete(inv, label)
	return nil
}
