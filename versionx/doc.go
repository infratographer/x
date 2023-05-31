// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

// Package versionx provides a single location for setting the version on
// infratographer binaries. You can set these values at build time using:
//
//	 		-X go.infratographer.com/x/versionx.appName=NAME
//	 		-X go.infratographer.com/x/versionx.version=VERSION
//			-X go.infratographer.com/x/versionx.gitCommit=GIT_SHA
//			-X go.infratographer.com/x/versionx.buildDate=DATE
//			-X go.infratographer.com/x/versionx.builder=USER
package versionx
