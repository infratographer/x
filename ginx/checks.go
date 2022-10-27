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
