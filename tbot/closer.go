package tbot

import "sync"

type Closer struct {
	handlers map[string]func() // мапа что б не добавлялось дубликатов
	mx       sync.RWMutex
}

func NewCloser() *Closer {
	return &Closer{handlers: map[string]func(){}}
}

func (c *Closer) Append(key string, f func()) {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.handlers[key] = f
}

func (c *Closer) Close() {
	c.mx.RLock()
	defer c.mx.RUnlock()

	for _, f := range c.handlers {
		f()
	}
}
