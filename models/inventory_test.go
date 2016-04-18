package models

import (
	"fmt"
	"time"

	. "gopkg.in/check.v1"
)

type IventorySuite struct {
	inventory Inventory
}

var _ = Suite(&IventorySuite{NewInventory()})

func (s *IventorySuite) TearDownTest(c *C) {
	s.inventory = NewInventory()
}

func (s *IventorySuite) insertFixtureItems(c *C) []*Item {
	itemLabels := []string{"foo", "bar", "biz"}
	items := []*Item{}

	for i, label := range itemLabels {
		items = append(items, NewItem(label, time.Now().Add(5*time.Minute), "type1"))
		c.Assert(
			s.inventory.Add(items[i]),
			IsNil,
		)
	}
	return items
}

func (s *IventorySuite) TestAddNew(c *C) {
	c.Assert(len(s.inventory), Equals, 0)
	s.insertFixtureItems(c)
	c.Assert(len(s.inventory), Equals, 3)

	err := s.inventory.Add(NewItem("foo", time.Now().Add(5*time.Minute), "type1"))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, fmt.Sprintf(DuplicateItemError, "foo"))
	c.Assert(len(s.inventory), Equals, 3)

	err = s.inventory.Add(NewItem("qux", time.Now().Add(-5*time.Minute), "type1"))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, ExpirationInPastError)
	c.Assert(len(s.inventory), Equals, 3)
}

func (s *IventorySuite) TestGetByLabel(c *C) {
	items := s.insertFixtureItems(c)

	for _, expectedItem := range items {
		empiricalItem, err := s.inventory.GetByLabel(expectedItem.Label)
		c.Assert(err, IsNil)
		c.Assert(empiricalItem, DeepEquals, expectedItem)
	}

	item, err := s.inventory.GetByLabel("qux")
	c.Assert(item, IsNil)
	c.Assert(err.Error(), Equals, fmt.Sprintf(ItemNotFoundError, "qux"))
}

func (s *IventorySuite) TestRemoveByLabel(c *C) {
	items := s.insertFixtureItems(c)
	origLength := len(items)

	for i, expectedItem := range items {
		err := s.inventory.RemoveByLabel(expectedItem.Label)
		c.Assert(err, IsNil)
		c.Assert(len(s.inventory)+1, Equals, origLength-i)
	}

	err := s.inventory.RemoveByLabel("qux")
	c.Assert(err.Error(), Equals, fmt.Sprintf(ItemNotFoundError, "qux"))
}
