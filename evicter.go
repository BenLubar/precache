package precache

import "time"

// Evicter decides when to remove entries from the cache. Its methods are
// called from one goroutine at a time as long as it is only associated with
// one Cache.
type Evicter interface {
	// ShouldEvict returns true if the unused entry associated with this
	// Getter should be removed from the cache.
	ShouldEvict(Getter) bool

	// Unused notifies the Evicter that a Getter has 0 references.
	Unused(Getter)

	// Used notifies the Evicter that a Getter has at least 1 reference.
	Used(Getter)
}

// LRUEvicter is an Evicter that removes the least recently used entries from
// the Cache.
type LRUEvicter struct {
	// The most recently used N unused entries in the cache are kept.
	N int

	g []Getter
}

// ShouldEvict implements Evicter.ShouldEvict.
func (l *LRUEvicter) ShouldEvict(g Getter) bool {
	for i, j := len(l.g)-1, l.N; i >= 0 && j > 0; i, j = i-1, j-1 {
		if g == l.g[i] {
			return false
		}
	}
	l.remove(g)
	return true
}

// Unused implements Evicter.Unused.
func (l *LRUEvicter) Unused(g Getter) {
	l.g = append(l.g, g)
}

// Used implements Evicter.Used.
func (l *LRUEvicter) Used(g Getter) {
	l.remove(g)
}

func (l *LRUEvicter) remove(g Getter) {
	for i, v := range l.g {
		if v == g {
			l.g = append(l.g[:i], l.g[i+1:]...)
			return
		}
	}
}

// DelayEvicter is an Evicter that evicts cache entries after a delay.
type DelayEvicter struct {
	D time.Duration
	g map[Getter]time.Time
}

// ShouldEvict implements Evicter.ShouldEvict.
func (d *DelayEvicter) ShouldEvict(g Getter) bool {
	if d.g[g].After(time.Now()) {
		delete(d.g, g)
		return true
	}
	return false
}

// Unused implements Evicter.Unused.
func (d *DelayEvicter) Unused(g Getter) {
	if d.g == nil {
		d.g = make(map[Getter]time.Time)
	}

	d.g[g] = time.Now().Add(d.D)
}

// Used implements Evicter.Used.
func (d *DelayEvicter) Used(g Getter) {
	delete(d.g, g)
}
