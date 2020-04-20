package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (e *engine) HandleThrough(c *gin.Context) {
	c.JSON(http.StatusOK, e.Through.Get())
}
