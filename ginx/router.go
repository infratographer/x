package ginx

import "github.com/gin-gonic/gin"

// Router is an interface that can be implemented to add routes to the server.
// The implementation is to be registered with the server using the AddHandler
// method. It is then added to the known routes list on initialization.
type Router interface {
	Routes(*gin.RouterGroup)
}
