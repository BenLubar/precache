package precache

import (
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/context"
)

type durationGetter time.Duration

func (g durationGetter) Get(ctx context.Context) (interface{}, error) {
	d := time.Duration(g)

	select {
	case <-time.After(d):
		return d, nil
	case <-ctx.Done():
		return nil, nil
	}
}

func TestPanics(t *testing.T) {
	expectPanic := func(msg string, f func()) {
		defer func() {
			if r := recover(); r == nil {
				t.Error(msg)
			} else {
				t.Logf("message: %v\npanic: %v", msg, r)
			}
		}()

		f()
	}

	c := NewCache(1)
	defer c.Close()

	e1 := c.Entry(durationGetter(time.Millisecond))
	e2 := e1.Clone()
	e1.Close()

	expectPanic("get after close", func() {
		e1.Get()
	})

	expectPanic("clone after close", func() {
		e1.Clone()
	})

	expectPanic("close after close", func() {
		e1.Close()
	})

	// the remaining function calls should not panic.
	e2.Get()
	e2.Clone().Close()
	e2.Close()
}

type countGetter uint64

func (g *countGetter) Get(ctx context.Context) (interface{}, error) {
	atomic.AddUint64((*uint64)(g), 1)

	return nil, nil
}

func TestGetter(t *testing.T) {
	c := NewCache(1)
	defer c.Close()

	g := new(countGetter)
	e1 := c.Entry(g)
	e2 := c.Entry(g)
	e3 := c.Entry(g)
	e4 := c.Entry(g)
	e5 := c.Entry(g)
	e1.Get()
	e2.Get()
	e3.Get()
	e4.Get()
	e5.Get()
	e1.Close()
	e2.Close()
	e3.Close()
	e4.Close()
	e5.Close()

	if expected, actual := uint64(1), atomic.LoadUint64((*uint64)(g)); expected != actual {
		t.Errorf("expected: %v\nactual: %v", expected, actual)
	}

	e6 := c.Entry(g)
	e6.Get()

	if expected, actual := uint64(2), atomic.LoadUint64((*uint64)(g)); expected != actual {
		t.Errorf("expected: %v\nactual: %v", expected, actual)
	}

	e7 := e6.Clone()
	e6.Close()
	e7.Get()

	if expected, actual := uint64(2), atomic.LoadUint64((*uint64)(g)); expected != actual {
		t.Errorf("expected: %v\nactual: %v", expected, actual)
	}

	e7.Close()
}
