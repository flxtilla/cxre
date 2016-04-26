package state

import (
	"sync"
	"time"

	"github.com/flxtilla/cxre/xrr"
)

type context struct {
	parent   *context
	mu       sync.Mutex
	children map[canceler]bool
	done     chan struct{}
	err      error
	value    *state
}

type canceler interface {
	cancel(removeFromParent bool, err error)
	Done() <-chan struct{}
}

var Canceled = xrr.NewXrror("flotilla.State canceled")

func propagateCancel(p *context, child canceler) {
	if p.Done() == nil {
		return
	}
	p.mu.Lock()
	if p.err != nil {
		child.cancel(false, p.err)
	} else {
		if p.children == nil {
			p.children = make(map[canceler]bool)
		}
		p.children[child] = true
	}
	p.mu.Unlock()
}

func (c *context) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c *context) Done() <-chan struct{} {
	return c.done
}

func (c *context) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err
}

func (c *context) Value(key interface{}) interface{} {
	return c.value
}

func (c *context) cancel(removeFromParent bool, err error) {
	if err == nil {
		panic("State.context: internal error: missing cancel error")
	}
	c.mu.Lock()
	c.err = err
	close(c.done)
	for child := range c.children {
		child.cancel(false, err)
	}
	c.children = nil
	c.mu.Unlock()

	if removeFromParent {
		if c.children != nil {
			delete(c.children, c)
		}
	}
}
