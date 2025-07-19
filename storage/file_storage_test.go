package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_locker(t *testing.T) {
	s := NewFileStorage()

	s.lock("test")
	s.unlock("test")

	s.lock("test")
	s.unlock("test")

	s.lock("test")
	s.lock("test2")
	s.unlock("test")
	s.unlock("test2")

	start := time.Now()
	go func() {
		s.lock("test")
		time.Sleep(time.Second)
		s.unlock("test")
	}()

	time.Sleep(time.Millisecond)

	s.lock("test")
	s.unlock("test")

	assert.GreaterOrEqual(t, time.Since(start), time.Second)
}
