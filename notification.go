package inventory

import (
	"sync"
	"time"

	"github.com/filiptc/inventory-system/models"
)

type Notifier interface {
	Notify(Message)
}

type notificationService struct {
	notifier             Notifier
	notificationMutex    sync.Mutex
	pendingNotifications map[string]*time.Timer
}

func newNotificationsService(notifier Notifier) *notificationService {
	return &notificationService{notifier, sync.Mutex{}, make(map[string]*time.Timer)}
}

func (n *notificationService) notifyExtraction(item models.Item) {
	defer n.notificationMutex.Unlock()
	n.notificationMutex.Lock()
	n.notifier.Notify(newRemovedItemMessage(item))
	if timer, ok := n.pendingNotifications[item.Label]; ok {
		timer.Stop()
		delete(n.pendingNotifications, item.Label)
	}
}

func (n *notificationService) queueExpiryNotification(item models.Item) {
	defer n.notificationMutex.Unlock()
	n.notificationMutex.Lock()
	n.pendingNotifications[item.Label] = time.AfterFunc(
		item.Expiration.Sub(time.Now()),
		func() {
			defer n.notificationMutex.Unlock()
			n.notificationMutex.Lock()

			n.notifier.Notify(newExpiredItemMessage(item))
			delete(n.pendingNotifications, item.Label)
		})
}
