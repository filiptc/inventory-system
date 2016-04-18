[![Build Status](https://travis-ci.org/filiptc/inventory-system.svg?branch=master)](https://travis-ci.org/filiptc/inventory-system)
[![codecov.io](https://codecov.io/github/filiptc/inventory-system/coverage.svg?branch=master)](https://codecov.io/github/filiptc/inventory-system?branch=master)


Table of contents
=================

  * [How to run](#how-to-run)
  * [Assumptions & reasoning](#assumptions--reasoning)
  * [Design & Code structure](#design--code-structure)
  * [Notes](#notes)


How to run
============

1. Install go ([documentation](https://golang.org/doc/install))
2. Get project `go get github.com/filiptc/inventory-system/...`
3. Run tests `go test github.com/filiptc/inventory-system/...`

From a different package, the **usage** would be as follows:

```go
package main

import (
    "fmt"
    "time"

    "github.com/filiptc/inventory-system"
    "github.com/filiptc/inventory-system/models"
)

type notifier struct{}

func (*notifier) Notify(msg inventory.Message) {
    // freedom to handle notifications client-side
    fmt.Println(msg.String())
}

func main() {
    // instanciate Inventory service
    inventoryService := inventory.NewInventoryService(&notifier{})

    // instanciate item
    expiration := time.Date(2017, 1, 1, 0, 0, 0, 0, time.Local)
    item := models.NewItem("label", expiration, "type")

    // add item to inventory
    err := inventoryService.Add(item)
    if err != nil {
        panic(err)
    }

    // [...]

    // get slice of current items
    items := inventoryService.GetInventoryItems()

    // do something with items

    // extract item from inventory
    item, err := inventoryService.ExtractByLabel("label")
    if err != nil {
        panic(err)
    }

    // do something with item
}
```


Assumptions & reasoning
=====
The formulation of the question was unclear to me. I was unsure whether the assignment was to
develop a WebAPI or an app and its public interface (API). I assumed the latter.

My main concern was to allow concurrency without overcomplicating the code. The main decision-point
while designing, was weather to implement the inventory as a map. In favour of this was the
possibility of using the label of each item as unique identifier, as the assignment suggests the
removal of individual items via label, hinting on the label's uniqueness.

However, maps are inherently thread-unsafe. There are some projects that provide thread-safe
solutions ([example](https://github.com/streamrail/concurrent-map)), but I felt these were clearly a
step towards overcomplication and went with a simpler, mutex-driven solution. Being aware that this
bears the risk of creating a bottleneck on heavy concurrent operations on the inventory, I weighed
it out against more complex solutions, and finally opted for this one anyway, favouring
code-maintainability over raw speed.

I also decided to deny the allocation of expired items in the inventory. This makes sense
conceptually and solves the problem of handling notifications of already expired items.

The other aspect I considered was how to address the design of the notifications. I considered an
implementing an event driven system with hooks for handlers to be registered. I finally decided to
go with a simpler approach and expose a `Notifier` interface, implementing the `Notify` method which
is called upon notification events. This allows a certain inversion of control without too much
overcomplication.

The `Notify` method takes a `Message` interface instead of a string. This was done as to give the
Notifier more freedom in the handling and of the notification event. The reason it doesn't take an
empty interface all together was to force it to implement the `Stringer` interface so the message
can be serialized into a plain message.

Finally, the most "annoying" part of the problem was the asynchronicity of certain events:

Firstly, I interpreted the "taking out" as an extraction, implying two actions:

1. fetch the item
2. delete the item from the inventory

This gives two options. Make the entire call blocking and not return the item until it has been
removed, or to return it immediately and delete the item asynchronously, notifying on completion.

Secondly, the expiry message of an item, which gives several options design-wise:

1. Iterate over all items on an arbitrary interval in one goroutine.
2. Have each item register a timer which executes the notification in its goroutine

Weighing options, 1. minimizes resources while complicating code and introducing an inaccuracy into
the system, as the notification will not be triggered on expiration, but shortly after. The
inaccuracy is dependant on the resolution of the interval. Option 2., on the other hand, is accurate
while slightly more resource-expensive (even though go is very efficient in goroutine handling). It
also requires canceling timers if items are removed before expiration, though it doesn't overly
complicate the code too much. I went with option 2.

Testing both asynchronous events implied adding `sleeps` to the test code.


Design & code structure
=====

This project consists of two layers and their tests. The general sevice API in the root directory,
and the models in the homonym folder.

* [inventory.go](inventory.go): Entry-point. Exposes a constructor for `InventoryService` and three
methods as the public interface: `Add` for adding items, `ExtractByLabel` to get and remove them and
`GetInventoryItems` to fetch copies of the inventory's contents.
It's members are instances of the inventory model, the `notificationService` and a mutex
* [notification.go](notification.go): Implements `notificationService`. Has aforementioned
`Notifier` instance, constructor function and two methods to register notifications (synchronous
removal and asynchronous expiry item notifications). Also declares public Notifier interface.
* [message.go](message.go): Declares public Message interface. Implements interface with
constructor for both handled messages.
* `models/`: Expose a number of public components in order to be accessible from the service, yet
remain a separate package (layer).
  * [inventory.go](models/inventory.go): Represents the inventory model as a map of Label:Item, a
  constructor and the methods for adding, and getting and removing items by label
  * [item.go](models/item.go): Represents the item model with a constructor and an "isser" method to
  determine if it has expired


Notes
=====

* More messages can be easily added in the [message.go](message.go) file and its methods in
[notification.go](notification.go) (e.g.: a notification when adding items).
* Test overage is of 100% (according to `go test ./... -cover`)
* An alternative version could be developed with channels for write operations instead of mutexes.
It would be interesting to do a performance comparison between both implementations. It would also
pave the way for implementing external queue systems like Gearman, RabbitMQ or Beanstalkd for
scalability (a refactor would be needed to convert system into producer and consumer of jobs).
* DB driver and "ORM" could easily be added "above" the model layer (e.g. `mgo`).
* In case the initial inventory was non-empty (e.g. loaded from DB), the method `initInventory` in
[inventory.go](inventory.go) goes through all items and registers their notifications.