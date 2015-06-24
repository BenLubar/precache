package precache

import "golang.org/x/net/context"

// Getter implementations must be immutable and comparable.
type Getter interface {
	// Get returns an object, possibly from disk, the network, or another
	// blocking resource. Get will be called from a new goroutine. A Getter
	// may optionally return early if the Context is Done.
	Get(context.Context) (interface{}, error)
}
