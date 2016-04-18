package inventory

import (
	"fmt"
	"time"

	"sync"

	"github.com/filiptc/inventory-system/models"
	. "gopkg.in/check.v1"
)

type IventorySuite struct {
	service *inventoryService
}

var _ = Suite(&IventorySuite{})

func (s *IventorySuite) SetUpTest(c *C) {
	s.service = NewInventoryService(&notifierStub{c, 0})
}

func (s *IventorySuite) insertFixtureItems(c *C) []*models.Item {
	itemLabels := []string{"foo", "bar", "biz"}
	items := []*models.Item{}

	for i, label := range itemLabels {
		items = append(items, models.NewItem(label, time.Now().Add(5*time.Minute), "type1"))
		c.Assert(
			s.service.Add(items[i]),
			IsNil,
		)
	}
	return items
}

func (s *IventorySuite) TestInitInventory(c *C) {
	s.service.Add(models.NewItem("foo", time.Now().Add(100*time.Millisecond), "type1"))
	s.service.Add(models.NewItem("bar", time.Now().Add(100*time.Millisecond), "type1"))

	time.Sleep(400 * time.Millisecond)
	// Checks that passed expiry, the notification is not being triggered
	c.Assert(s.service.notifications.notifier.(*notifierStub).calls, Equals, 2)
}

func (s *IventorySuite) TestAddNew(c *C) {
	c.Assert(len(s.service.inventory), Equals, 0)
	s.insertFixtureItems(c)
	c.Assert(len(s.service.inventory), Equals, 3)

	err := s.service.Add(models.NewItem("foo", time.Now().Add(5*time.Minute), "type1"))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, fmt.Sprintf(models.DuplicateItemError, "foo"))
	c.Assert(len(s.service.inventory), Equals, 3)

	err = s.service.Add(models.NewItem("qux", time.Now().Add(-5*time.Minute), "type1"))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, models.ExpirationInPastError)
	c.Assert(len(s.service.inventory), Equals, 3)
}

func (s *IventorySuite) TestExtractByLabel(c *C) {
	items := s.insertFixtureItems(c)

	for _, expectedItem := range items {
		empiricalItem, err := s.service.ExtractByLabel(expectedItem.Label)
		c.Assert(err, IsNil)
		c.Assert(empiricalItem, DeepEquals, expectedItem)
	}

	// removals are async, we need to wait for them
	time.Sleep(500 * time.Millisecond)
	c.Assert(len(s.service.inventory), Equals, 0)

	item, err := s.service.ExtractByLabel("qux")
	c.Assert(item, IsNil)
	c.Assert(err.Error(), Equals, fmt.Sprintf(models.ItemNotFoundError, "qux"))
}

func (s *IventorySuite) TestServiceWithConcurrency(c *C) {
	var wg = sync.WaitGroup{}

	itemLabels := []string{"foo", "bar", "biz"}
	items := []*models.Item{}
	for _, label := range itemLabels {
		items = append(items, models.NewItem(label, time.Now().Add(5*time.Minute), "type1"))
	}

	// emulate concurrency on same inventory instance to show thread safety
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.Assert(s.service.Add(items[0]), IsNil)
		empiricalItem, err := s.service.ExtractByLabel(items[0].Label)
		c.Assert(err, IsNil)
		c.Assert(empiricalItem, Equals, items[0])

		// removals are async, we need to wait for them
		time.Sleep(500 * time.Millisecond)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		c.Assert(s.service.Add(items[1]), IsNil)
		c.Assert(s.service.Add(items[2]), IsNil)
		empiricalItem, err := s.service.ExtractByLabel(items[1].Label)
		c.Assert(err, IsNil)
		c.Assert(empiricalItem, Equals, items[1])

		// removals are async, we need to wait for them
		time.Sleep(500 * time.Millisecond)
	}()

	wg.Wait()
	c.Assert(s.service.notifications.notifier.(*notifierStub).calls, Equals, 2)
}

func (s *IventorySuite) TestExpiryNotification(c *C) {
	item := models.NewItem("foo", time.Now().Add(1*time.Second), "type1")
	c.Assert(s.service.Add(item), IsNil)
	c.Assert(s.service.notifications.notifier.(*notifierStub).calls, Equals, 0)

	time.Sleep(1500 * time.Millisecond)
	c.Assert(s.service.notifications.notifier.(*notifierStub).calls, Equals, 1)
}

func (s *IventorySuite) TestItemExtractedBeforeExpiryNotification(c *C) {
	c.Assert(len(s.service.inventory), Equals, 0)
	item := models.NewItem("foo", time.Now().Add(1*time.Second), "type1")
	c.Assert(s.service.Add(item), IsNil)

	empiricalItem, err := s.service.ExtractByLabel(item.Label)

	// removals are async, we need to wait for them
	time.Sleep(500 * time.Millisecond)
	c.Assert(s.service.notifications.notifier.(*notifierStub).calls, Equals, 1)
	c.Assert(err, IsNil)
	c.Assert(empiricalItem, Equals, item)
	c.Assert(len(s.service.inventory), Equals, 0)

	time.Sleep(2 * time.Second)
	// Checks that passed expiry, the notification is not being triggered
	c.Assert(s.service.notifications.notifier.(*notifierStub).calls, Equals, 1)
}

func (s *IventorySuite) TestGetInventoryItems(c *C) {
	mappedItems := make(map[string]models.Item)
	for _, i := range s.insertFixtureItems(c) {
		mappedItems[i.Label] = *i
	}
	retrievedItems := s.service.GetInventoryItems()
	c.Assert(len(retrievedItems), Equals, len(mappedItems))

	for _, i := range retrievedItems {
		c.Assert(i, Equals, mappedItems[i.Label])
	}
}

