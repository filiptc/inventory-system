package models

import (
	"time"

	. "gopkg.in/check.v1"
)

type ItemSuite struct{}

var _ = Suite(&ItemSuite{})

func (s *ItemSuite) TestHasExpired(c *C) {
	i1 := NewItem("foo", time.Now().Add(-5*time.Minute), "bar")
	i2 := NewItem("foo", time.Now().Add(5*time.Minute), "bar")
	c.Assert(i1.HasExpired(time.Now()), Equals, true)
	c.Assert(i2.HasExpired(time.Now()), Equals, false)
}
