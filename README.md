# precache
--
    import "github.com/BenLubar/precache"


## Usage

#### type Cache

```go
type Cache struct {
}
```

Cache remembers the result of Getter.Get until all references are closed.

#### func  NewCache

```go
func NewCache(deferred int) *Cache
```
NewCache constructs a new Cache. The number of concurrent deferred getters may
be zero or more.

#### func (*Cache) Close

```go
func (c *Cache) Close()
```
Close cancels any getters in progress and stops the deferred queue goroutine.
Entry.Close must still be called on each live entry.

#### func (*Cache) Deferred

```go
func (c *Cache) Deferred(g Getter) *Entry
```
Deferred returns a reference to the cached value of the Getter, and adds the
Getter to a queue if it is not in the cache. The queue will run a specified
number of concurrent getters at a time.

#### func (*Cache) Entry

```go
func (c *Cache) Entry(g Getter) *Entry
```
Entry returns a reference to the cached value of the Getter, and starts the
Getter immediately if it is not in the cache.

#### func (*Cache) SetEvicter

```go
func (c *Cache) SetEvicter(evicter Evicter)
```
SetEvicter assigns an Evicter to the Cache. Using nil causes unused entries to
be evicted immediately. An Evicter should only be used with one Cache.

#### type DelayEvicter

```go
type DelayEvicter struct {
	D time.Duration
}
```

DelayEvicter is an Evicter that evicts cache entries after a delay.

#### func (*DelayEvicter) ShouldEvict

```go
func (d *DelayEvicter) ShouldEvict(g Getter) bool
```
ShouldEvict implements Evicter.ShouldEvict.

#### func (*DelayEvicter) Unused

```go
func (d *DelayEvicter) Unused(g Getter)
```
Unused implements Evicter.Unused.

#### func (*DelayEvicter) Used

```go
func (d *DelayEvicter) Used(g Getter)
```
Used implements Evicter.Used.

#### type Entry

```go
type Entry struct {
}
```

Entry is a reference to an entry of a Cache. A single Entry may only be used
from one goroutine. An Entry must be closed to allow it to be removed from the
cache. Calling a method on a closed Entry will cause a panic.

#### func (*Entry) Clone

```go
func (e *Entry) Clone() *Entry
```
Clone returns a copy of this Entry that is closed separately.

#### func (*Entry) Close

```go
func (e *Entry) Close()
```
Close removes the reference to this entry. When all entries for a Getter have
been closed, the value is removed fromt the cache.

#### func (*Entry) Get

```go
func (e *Entry) Get() (interface{}, error)
```
Get returns the result of the Getter. If the Getter has not yet finished, Get
blocks until the Getter returns.

#### type Evicter

```go
type Evicter interface {
	// ShouldEvict returns true if the unused entry associated with this
	// Getter should be removed from the cache.
	ShouldEvict(Getter) bool

	// Unused notifies the Evicter that a Getter has 0 references.
	Unused(Getter)

	// Used notifies the Evicter that a Getter has at least 1 reference.
	Used(Getter)
}
```

Evicter decides when to remove entries from the cache. Its methods are called
from one goroutine at a time as long as it is only associated with one Cache.

#### type Getter

```go
type Getter interface {
	// Get returns an object, possibly from disk, the network, or another
	// blocking resource. Get will be called from a new goroutine. A Getter
	// may optionally return early if the Context is Done.
	Get(context.Context) (interface{}, error)
}
```

Getter implementations must be comparable.

#### type LRUEvicter

```go
type LRUEvicter struct {
	// The most recently used N unused entries in the cache are kept.
	N int
}
```

LRUEvicter is an Evicter that removes the least recently used entries from the
Cache.

#### func (*LRUEvicter) ShouldEvict

```go
func (l *LRUEvicter) ShouldEvict(g Getter) bool
```
ShouldEvict implements Evicter.ShouldEvict.

#### func (*LRUEvicter) Unused

```go
func (l *LRUEvicter) Unused(g Getter)
```
Unused implements Evicter.Unused.

#### func (*LRUEvicter) Used

```go
func (l *LRUEvicter) Used(g Getter)
```
Used implements Evicter.Used.
