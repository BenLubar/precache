package precache

import (
	"sync/atomic"
	"testing"
)

func TestLRUEvicter(t *testing.T) {
	evicter := &LRUEvicter{N: 1}
	cache := NewCache(0)
	defer cache.Close()

	cache.SetEvicter(evicter)

	g1 := new(countGetter)
	g2 := new(countGetter)
	g3 := new(countGetter)

	e := cache.Entry(g1)
	e.Get()
	e.Close()
	if expected, actual := uint64(1), atomic.LoadUint64((*uint64)(g1)); expected != actual {
		t.Errorf("expected: %v\nactual: %v", expected, actual)
	}
	if expected, actual := uint64(0), atomic.LoadUint64((*uint64)(g2)); expected != actual {
		t.Errorf("expected: %v\nactual: %v", expected, actual)
	}
	if expected, actual := uint64(0), atomic.LoadUint64((*uint64)(g3)); expected != actual {
		t.Errorf("expected: %v\nactual: %v", expected, actual)
	}

	e = cache.Entry(g2)
	e.Get()
	e.Close()
	if expected, actual := uint64(1), atomic.LoadUint64((*uint64)(g1)); expected != actual {
		t.Errorf("expected: %v\nactual: %v", expected, actual)
	}
	if expected, actual := uint64(1), atomic.LoadUint64((*uint64)(g2)); expected != actual {
		t.Errorf("expected: %v\nactual: %v", expected, actual)
	}
	if expected, actual := uint64(0), atomic.LoadUint64((*uint64)(g3)); expected != actual {
		t.Errorf("expected: %v\nactual: %v", expected, actual)
	}

	e = cache.Entry(g3)
	e.Get()
	e.Close()
	if expected, actual := uint64(1), atomic.LoadUint64((*uint64)(g1)); expected != actual {
		t.Errorf("expected: %v\nactual: %v", expected, actual)
	}
	if expected, actual := uint64(1), atomic.LoadUint64((*uint64)(g2)); expected != actual {
		t.Errorf("expected: %v\nactual: %v", expected, actual)
	}
	if expected, actual := uint64(1), atomic.LoadUint64((*uint64)(g3)); expected != actual {
		t.Errorf("expected: %v\nactual: %v", expected, actual)
	}

	e = cache.Entry(g2)
	e.Get()
	e.Close()
	if expected, actual := uint64(1), atomic.LoadUint64((*uint64)(g1)); expected != actual {
		t.Errorf("expected: %v\nactual: %v", expected, actual)
	}
	if expected, actual := uint64(1), atomic.LoadUint64((*uint64)(g2)); expected != actual {
		t.Errorf("expected: %v\nactual: %v", expected, actual)
	}
	if expected, actual := uint64(1), atomic.LoadUint64((*uint64)(g3)); expected != actual {
		t.Errorf("expected: %v\nactual: %v", expected, actual)
	}

	e = cache.Entry(g1)
	e.Get()
	e.Close()
	if expected, actual := uint64(2), atomic.LoadUint64((*uint64)(g1)); expected != actual {
		t.Errorf("expected: %v\nactual: %v", expected, actual)
	}
	if expected, actual := uint64(1), atomic.LoadUint64((*uint64)(g2)); expected != actual {
		t.Errorf("expected: %v\nactual: %v", expected, actual)
	}
	if expected, actual := uint64(1), atomic.LoadUint64((*uint64)(g3)); expected != actual {
		t.Errorf("expected: %v\nactual: %v", expected, actual)
	}
}
