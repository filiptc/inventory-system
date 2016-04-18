package inventory

import (
	"fmt"

	"time"

	"github.com/filiptc/inventory-system/models"
	. "gopkg.in/check.v1"
)

type NotificationSuite struct {
	service *notificationService
}

var _ = Suite(&NotificationSuite{})

type notifierStub struct {
	c     *C
	calls int
}

func (n *notifierStub) Notify(msg Message) {
	n.calls = n.calls + 1
	n.c.Assert(msg.(*itemMessage).Item, NotNil)
	n.c.Assert(len(msg.String()) > 0, Equals, true)

	// only visible if tests are run with "-v"
	fmt.Println(msg.String())

}

func (s *NotificationSuite) SetUpTest(c *C) {
	s.service = newNotificationsService(&notifierStub{c, 0})
}

func (s *NotificationSuite) TestNotifyExtraction(c *C) {
	item := models.NewItem("foo", time.Now().Add(5*time.Minute), "type1")
	s.service.notifyExtraction(*item)
	c.Assert(s.service.notifier.(*notifierStub).calls, Equals, 1)
}

func (s *NotificationSuite) TestQueueExpiryNotification(c *C) {
	item := models.NewItem("foo", time.Now().Add(1*time.Second), "type1")
	s.service.queueExpiryNotification(*item)
	time.Sleep(2 * time.Second)
	c.Assert(s.service.notifier.(*notifierStub).calls, Equals, 1)
}
