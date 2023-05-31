// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package ginx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// livenessCheckHandler ensures that the server is up and responding
func (s *Server) livenessCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "UP",
	})
}

// readinessCheckHandler ensures that the server is up and that we are able to process
// requests. It will check any readinessChecks that have been provided and return
// their status when calculating if the service is ready.
func (s *Server) readinessCheckHandler(c *gin.Context) {
	failed := false
	status := map[string]string{}

	for name, check := range s.readinessChecks {
		if err := check(c.Request.Context()); err != nil {
			s.logger.Sugar().Errorf("readiness check failed", "name", name, "error", err)

			failed = true
			status[name] = err.Error()
		} else {
			status[name] = "OK"
		}
	}

	if failed {
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}

	c.JSON(http.StatusOK, status)
}

// version returns the version build information.
func (s *Server) versionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.version)
}
