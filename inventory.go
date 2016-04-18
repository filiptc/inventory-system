package inventory

import (
	"sync"

	"github.com/filiptc/inventory-system/models"
)

type inventoryService struct {
	mutex         sync.Mutex
	inventory     models.Inventory
	notifications *notificationService
}

func NewInventoryService(notifier Notifier) *inventoryService {
	is := &inventoryService{
		mutex:         sync.Mutex{},
		inventory:     models.NewInventory(),
		notifications: newNotificationsService(notifier),
	}

	// in case of existing initial data (e.g. DB)
	is.initInventory()

	return is
}

func (i *inventoryService) initInventory() {
	for _, item := range i.GetInventoryItems() {
		i.notifications.queueExpiryNotification(item)
	}
}

func (i *inventoryService) Add(item *models.Item) error {
	defer i.mutex.Unlock()
	i.mutex.Lock()
	if err := i.inventory.Add(item); err != nil {
		return err
	}

	i.notifications.queueExpiryNotification(*item)
	return nil
}

func (i *inventoryService) ExtractByLabel(label string) (*models.Item, error) {
	defer i.mutex.Unlock()
	i.mutex.Lock()
	item, err := i.inventory.GetByLabel(label)
	if err != nil {
		return nil, err
	}

	// asynchronous removal
	go func() {
		defer i.mutex.Unlock()
		i.mutex.Lock()
		i.inventory.RemoveByLabel(label)
		i.notifications.notifyExtraction(*item)
	}()

	return item, nil
}

func (i *inventoryService) GetInventoryItems() []models.Item {
	items := []models.Item{}
	for _, item := range i.inventory {
		items = append(items, *item)
	}
	return items
}