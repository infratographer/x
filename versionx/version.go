// Copyright 2022 The Infratographer Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package versionx

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// These variables are substituted with real values at build time
var (
	appName = "versionx/unknown"
	version = "unknown"
	commit  = ""
	date    = ""
	builtBy = "dev"
)

// Details stores all the information collected about a version.
type Details struct {
	AppName string     `json:"app"`
	Version string     `json:"version"`
	Commit  string     `json:"commit,omitempty"`
	BuiltAt *time.Time `json:"built_at,omitempty"`
	Builder string     `json:"builder"`
}

// String returns the version as a formatted string
func (d *Details) String() string {
	return fmt.Sprintf("%s: %s (%s@%s by %s)", d.AppName, d.Version, d.Commit, d.BuiltAt.String(), d.Builder)
}

// BuildDetails will return a Details struct containing all the values that were
// set at build time to provide you the current version information.
func BuildDetails() *Details {
	d := &Details{
		AppName: appName,
		Version: version,
		Commit:  commit,
		Builder: builtBy,
	}

	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		d.BuiltAt = &t
	}

	return d
}

// RegisterCobraCommand will add a version command to the cobra command provided
// that prints out the version. An optional logger may be provided at which point
// the version will be printed as an Info level log with the logger.
func RegisterCobraCommand(cmd *cobra.Command, printFunc func()) {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "returns the application version information",
		Run: func(cmd *cobra.Command, args []string) {
			printFunc()
		},
	}

	cmd.AddCommand(versionCmd)
}

// PrintVersion will print out the details of the current build. If a logger is
// provided they will be printed with the logger, otherwise they will just be
// printed as output.
func PrintVersion(lgr *zap.SugaredLogger) {
	if lgr == nil {
		fmt.Println(BuildDetails().String())
		return
	}

	lgr.Infow("version details",
		"AppName", appName,
		"Version", version,
		"Commit", commit,
		"Builder", builtBy,
		"BuiltAt", date,
	)
}
