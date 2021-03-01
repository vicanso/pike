// MIT License

// Copyright (c) 2020 Tree Xie

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package app

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/vicanso/pike/log"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type Info struct {
	GOARCH       string `json:"goarch,omitempty"`
	GOOS         string `json:"goos,omitempty"`
	GoVersion    string `json:"goVersion,omitempty"`
	Version      string `json:"version,omitempty"`
	BuildedAt    string `json:"buildedAt,omitempty"`
	CommitID     string `json:"commitID,omitempty"`
	Uptime       string `json:"uptime,omitempty"`
	GoMaxProcs   int    `json:"goMaxProcs,omitempty"`
	CPUUsage     int32  `json:"cpuUsage,omitempty"`
	RoutineCount int    `json:"routineCount,omitempty"`
	ThreadCount  int32  `json:"threadCount,omitempty"`
	RSS          uint64 `json:"rss,omitempty"`
	RSSHumanize  string `json:"rssHumanize,omitempty"`
	Swap         uint64 `json:"swap,omitempty"`
	SwapHumanize string `json:"swapHumanize,omitempty"`
}

var buildedAt time.Time
var commitID string
var version string
var startedAt = time.Now()
var currentProcess *process.Process
var cpuUsage = atomic.NewInt32(-1)

func init() {
	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		log.Default().Error("new process fail",
			zap.Error(err),
		)
	}
	currentProcess = p
	_ = UpdateCPUUsage()
}

// SetBuildInfo set build info
func SetBuildInfo(build, id, ver, buildBy string) {
	buildedAt, _ = time.Parse(time.RFC3339, strings.Replace(build, " ", "T", 1))
	commitID = id
	if len(id) > 7 {
		commitID = id[0:7]
	}
	version = ver
}

const MB = 1024 * 1024

func bytesToMB(value uint64) string {
	v := value / MB
	return strconv.Itoa(int(v)) + " MB"
}

func GetVersion() string {
	return version
}

// GetInfo get application info
func GetInfo() *Info {
	uptime := ""
	d := time.Since(startedAt)
	if d > 24*time.Hour {
		uptime = strconv.Itoa(int(d/(24*time.Hour))) + "d"
	} else if d > time.Hour {
		uptime = strconv.Itoa(int(d.Hours())) + "h"
	} else {
		uptime = (time.Second * time.Duration(d.Seconds())).String()
	}

	info := &Info{
		GOARCH:       runtime.GOARCH,
		GOOS:         runtime.GOOS,
		GoVersion:    runtime.Version(),
		Version:      GetVersion(),
		BuildedAt:    buildedAt.Format(time.RFC3339),
		CommitID:     commitID,
		Uptime:       uptime,
		GoMaxProcs:   runtime.GOMAXPROCS(0),
		CPUUsage:     cpuUsage.Load(),
		RoutineCount: runtime.NumGoroutine(),
	}
	if currentProcess != nil {
		info.ThreadCount, _ = currentProcess.NumThreads()
		memInfo, _ := currentProcess.MemoryInfo()
		if memInfo != nil {
			info.RSS = memInfo.RSS
			info.RSSHumanize = bytesToMB(memInfo.RSS)
			info.Swap = memInfo.Swap
			info.SwapHumanize = bytesToMB(memInfo.Swap)
		}
	}
	return info
}

// UpdateCPUUsage update cpu usage
func UpdateCPUUsage() error {
	if currentProcess == nil {
		return nil
	}
	usage, err := currentProcess.Percent(0)
	if err != nil {
		return err
	}
	cpuUsage.Store(int32(usage))
	return nil
}
