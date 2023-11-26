package sidecar

// Emulates stack of cleanup functions
type Cleaner struct {
	cleans []func()
}

// Push cleanup function to stack
func (c *Cleaner) Push(cf func()) {
	c.cleans = append([]func(){cf}, c.cleans...)
}

// Cleanup in LIFO order
func (c *Cleaner) Clean() {
	if c == nil {
		return
	}

	for _, cf := range c.cleans {
		cf()
	}
}
