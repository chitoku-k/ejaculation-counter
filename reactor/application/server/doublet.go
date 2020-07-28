package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (e *engine) HandleDoublet(c *gin.Context) {
	c.JSON(http.StatusOK, e.Doublet.Get())
}
