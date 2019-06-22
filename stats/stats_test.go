package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStats(t *testing.T) {
	s := New()
	t.Run("concurrency", func(t *testing.T) {
		assert := assert.New(t)
		assert.Equal(uint32(1), s.IncreaseConcurrency())
		assert.Equal(uint32(1), s.GetConcurrency())
		assert.Equal(uint32(0), s.DecreaseConcurrency())
		assert.Equal(uint32(0), s.GetConcurrency())
	})

	t.Run("request count", func(t *testing.T) {
		assert := assert.New(t)
		assert.Equal(uint64(1), s.IncreaseRequestCount())
	})

	t.Run("recover count", func(t *testing.T) {
		assert := assert.New(t)
		assert.Equal(uint32(1), s.IncreaseRecoverCount())
	})

	t.Run("add request stats", func(t *testing.T) {
		assert := assert.New(t)
		s.AddRequestStats(100, 10)
		assert.Equal(uint64(1), s.status1Count)
		assert.Equal(uint64(1), s.spdy0Count)

		s.AddRequestStats(200, 40)
		assert.Equal(uint64(1), s.status2Count)
		assert.Equal(uint64(1), s.spdy1Count)

		s.AddRequestStats(300, 400)
		assert.Equal(uint64(1), s.status3Count)
		assert.Equal(uint64(1), s.spdy2Count)

		s.AddRequestStats(400, 1500)
		assert.Equal(uint64(1), s.status4Count)
		assert.Equal(uint64(1), s.spdy3Count)

		s.AddRequestStats(500, 5000)
		assert.Equal(uint64(1), s.status5Count)
		assert.Equal(uint64(1), s.spdy4Count)
	})

	t.Run("get info", func(t *testing.T) {
		assert := assert.New(t)
		info := s.GetInfo()
		assert.NotNil(info.Status)
	})
}
