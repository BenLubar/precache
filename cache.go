package precache

import (
	"runtime"
	"sync"
	"sync/atomic"

	"golang.org/x/net/context"
)

// Cache remembers the result of Getter.Get until all references are closed.
type Cache struct {
	m    map[Getter]*entry
	ch   chan *entry
	ctx  context.Context
	done context.CancelFunc
	lock sync.RWMutex
}

// NewCache constructs a new Cache. The number of concurrent deferred getters
// may be zero or more.
func NewCache(deferred int) *Cache {
	ctx, done := context.WithCancel(context.TODO())
	c := &Cache{
		m:    make(map[Getter]*entry),
		ch:   make(chan *entry),
		ctx:  ctx,
		done: done,
	}
	go c.deferred(deferred)
	return c
}

// Entry returns a reference to the cached value of the Getter, and starts
// the Getter immediately if it is not in the cache.
func (c *Cache) Entry(g Getter) *Entry {
	e := &Entry{c.entry(g, true)}
	runtime.SetFinalizer(e, (*Entry).warnIfNotClosed)
	return e
}

// Deferred returns a reference to the cached value of the Getter, and adds the
// Getter to a queue if it is not in the cache. The queue will run a specified
// number of concurrent getters at a time.
func (c *Cache) Deferred(g Getter) *Entry {
	e := &Entry{c.entry(g, false)}
	runtime.SetFinalizer(e, (*Entry).warnIfNotClosed)
	return e
}

func (c *Cache) entry(g Getter, immediate bool) *entry {
	c.lock.RLock()
	// Easy way out: we already have the entry in cache.
	if e, ok := c.m[g]; ok {
		atomic.AddInt64(&e.ref, 1)
		c.lock.RUnlock()
		return e
	}
	c.lock.RUnlock()

	c.lock.Lock()
	defer c.lock.Unlock()
	// Check again to make sure it wasn't added while we were unlocked.
	if e, ok := c.m[g]; ok {
		atomic.AddInt64(&e.ref, 1)
		return e
	}

	// Make a new entry, put it in the map, and spawn a goroutine.
	return c.newEntry(g, immediate)
}

func (c *Cache) newEntry(g Getter, immediate bool) *entry {
	ctx, done := context.WithCancel(c.ctx)
	e := &entry{c: c, g: g, ref: 1, ctx: ctx, done: done}
	c.m[g] = e

	if immediate {
		go e.once.Do(e.fetch)
	} else {
		select {
		case c.ch <- e:
		case <-c.ctx.Done():
		}
	}

	return e
}

// Close cancels any getters in progress and stops the deferred queue
// goroutine. Entry.Close must still be called on each live entry.
func (c *Cache) Close() {
	c.done()
}

func (c *Cache) deferred(maxConcurrent int) {
	if maxConcurrent <= 0 {
		for {
			select {
			case <-c.ctx.Done():
				return

			case <-c.ch:
				// discard
			}
		}
	}

	semaphore := make(chan struct{}, maxConcurrent)
	for i := 0; i < maxConcurrent; i++ {
		semaphore <- struct{}{}
	}

	var waiting []*entry

	for {
		var ready <-chan struct{}
		if len(waiting) > 0 {
			ready = semaphore
		}

		select {
		case e := <-c.ch:
			waiting = append(waiting, e)

		case <-ready:
			e := waiting[0]
			waiting = waiting[1:]

			c.lock.RLock()
			if c.m[e.g] == e {
				atomic.AddInt64(&e.ref, 1)
				go func(e *Entry) {
					e.Get()
					e.Close()
					semaphore <- struct{}{}
				}(&Entry{e})
			} else {
				semaphore <- struct{}{}
			}
			c.lock.RUnlock()

		case <-c.ctx.Done():
			return
		}
	}
}
