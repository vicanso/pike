// Copyright 2020 tree xie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package application

import (
	"runtime"
	"time"

	"github.com/gobuffalo/packr/v2"
)

var (
	defaultApp *application
	assetBox   = packr.New("app-asset", "../assets")
)

type application struct {
	BuildedAt    time.Time `json:"buildedAt,omitempty"`
	StartedAt    time.Time `json:"startedAt,omitempty"`
	CommitID     string    `json:"commitId,omitempty"`
	GOOS         string    `json:"goos,omitempty"`
	MaxProcs     int       `json:"maxProcs,omitempty"`
	NumGoroutine int       `json:"numGoroutine,omitempty"`
	Version      string    `json:"version,omitempty"`
}

func init() {
	version, _ := assetBox.FindString("version")
	defaultApp = &application{
		MaxProcs:  runtime.GOMAXPROCS(0),
		GOOS:      runtime.GOOS,
		StartedAt: time.Now(),
		Version:   version,
	}
}

func (app *application) SetBuildedAt(buildedAt string) {
	t, err := time.Parse("20060102.150405", buildedAt)
	if err != nil {
		return
	}
	app.BuildedAt = t
}

func (app *application) SetCommitID(id string) {
	app.CommitID = id
}

func Default() *application {
	defaultApp.NumGoroutine = runtime.NumGoroutine()
	return defaultApp
}

func DefaultAsset() *packr.Box {
	return assetBox
}
