package app

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetSetBuildInfo(t *testing.T) {
	assert := assert.New(t)
	id := "123"
	SetBuildInfo("2021-02-27T01:41:32.416Z", id, "version", "")

	assert.Equal(id, commitID)
	assert.Equal("2021-02-27 01:41:32.416 +0000 UTC", buildedAt.UTC().String())
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
