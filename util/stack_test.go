package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStack(t *testing.T) {
	stack := GetStack(0, 5)
	assert.Equal(t, len(stack), 5)
}
