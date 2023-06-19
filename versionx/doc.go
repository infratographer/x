// Copyright 2023 The Infratographer Authors
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

// Package versionx provides a single location for setting the version on
// infratographer binaries. You can set these values at build time using:
//
//	 		-X go.infratographer.com/x/versionx.appName=NAME
//	 		-X go.infratographer.com/x/versionx.version=VERSION
//			-X go.infratographer.com/x/versionx.gitCommit=GIT_SHA
//			-X go.infratographer.com/x/versionx.buildDate=DATE
//			-X go.infratographer.com/x/versionx.builder=USER
package versionx
