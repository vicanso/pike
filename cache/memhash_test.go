package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemHash(t *testing.T) {
	assert := assert.New(t)

	data := "GET aslant.site /users/v1/me"
	v := MemHash([]byte(data))
	assert.Equal(v, MemHash([]byte(data)))
	assert.Equal(v, MemHashString(data))

	assert.NotEqual(v, MemHash([]byte("abc")))
}
