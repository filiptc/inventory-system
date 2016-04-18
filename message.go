package inventory

import (
	"fmt"

	"github.com/filiptc/inventory-system/models"
)

type Message interface {
	String() string
}

type itemMessage struct {
	FormatMessage func(models.Item) string
	Item          models.Item
}

func (im *itemMessage) String() string {
	return im.FormatMessage(im.Item)
}

func newRemovedItemMessage(item models.Item) *itemMessage {
	return &itemMessage{
		func(i models.Item) string {
			return fmt.Sprintf(
				"Item removed with label %s, expiry %v and type %s",
				i.Label,
				i.Expiration,
				i.Type)
		},
		item,
	}
}

func newExpiredItemMessage(item models.Item) *itemMessage {
	return &itemMessage{
		func(i models.Item) string {
			return fmt.Sprintf(
				"Item expired with label %s and type %s",
				i.Label,
				i.Type,
			)
		},
		item,
	}
}
