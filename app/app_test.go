package app

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetSetBuildInfo(t *testing.T) {
	assert := assert.New(t)
	id := "123"
	SetBuildInfo("20201206.051341", id)

	assert.Equal(id, commitID)
	assert.Equal("2020-12-06 05:13:41 +0000 UTC", buildedAt.UTC().String())
}

func TestUpdateCPUUsage(t *testing.T) {
	assert := assert.New(t)

	err := UpdateCPUUsage()
	assert.Nil(err)
}

func TestGetInfo(t *testing.T) {
	assert := assert.New(t)

	info := GetInfo()
	assert.Equal(runtime.GOOS, info.GOOS)
}
