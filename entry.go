package precache

import (
	"log"
	"runtime"
	"sync"
	"sync/atomic"

	"golang.org/x/net/context"
)

// Entry is a reference to an entry of a Cache. A single Entry may only be used
// from one goroutine. An Entry must be closed to allow it to be removed from
// the cache. Calling a method on a closed Entry will cause a panic.
type Entry struct {
	e *entry
}

// Get returns the result of the Getter. If the Getter has not yet finished,
// Get blocks until the Getter returns.
func (e *Entry) Get() (interface{}, error) {
	if e.e == nil {
		panic("precache: call to Get on a closed Entry.")
	}
	return e.e.get()
}

// Clone returns a copy of this Entry that is closed separately.
func (e *Entry) Clone() *Entry {
	if e.e == nil {
		panic("precache: call to Clone on a closed Entry.")
	}
	c := &Entry{e.e}
	// e.e.ref is at least 1 because e has not been closed yet, so we don't
	// need to worry about e.e being removed before we add the reference.
	atomic.AddInt64(&e.e.ref, 1)
	runtime.SetFinalizer(c, (*Entry).warnIfNotClosed)
	return c
}

// Close removes the reference to this entry. When all entries for a Getter
// have been closed, the value is removed fromt the cache.
func (e *Entry) Close() {
	ee := e.e
	if ee == nil {
		panic("precache: call to Close on a closed Entry.")
	}
	e.e = nil

	switch remain := atomic.AddInt64(&ee.ref, -1); {
	case remain > 0:
		return

	case remain == 0:
		ee.c.lock.Lock()
		defer ee.c.lock.Unlock()

		// Double check to make sure a new reference wasn't added while
		// we were getting the lock.
		if atomic.LoadInt64(&ee.ref) != 0 {
			return
		}

		// Tell the Getter to stop if it's still running.
		ee.done()

		// Remove the entry from the map.
		delete(ee.c.m, ee.g)

	default:
		panic("precache: internal error: reference count underflow")
	}
}

func (e *Entry) warnIfNotClosed() {
	if e.e != nil {
		log.Printf("precache: Entry(%p) leaked! (Getter: %v)", e, e.e.g)
		runtime.Breakpoint()
		e.Close()
	}
}

type entry struct {
	// c, g, ctx, done, and once should be considered immutable.
	// ref is the reference count. it may only be changed atomically.
	//
	// v and err are illegal to access before e.once.Do(e.fetch) returns,
	// and are considered immutable after.

	c    *Cache
	g    Getter
	ctx  context.Context
	done context.CancelFunc
	ref  int64
	once sync.Once
	v    interface{}
	err  error
}

func (e *entry) get() (interface{}, error) {
	e.once.Do(e.fetch)
	return e.v, e.err
}

func (e *entry) fetch() {
	e.v, e.err = e.g.Get(e.ctx)
}
