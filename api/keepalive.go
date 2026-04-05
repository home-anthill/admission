package api

import (
	"admission/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetKeepAlive returns a simple health check response.
func GetKeepAlive(c *gin.Context) {
	response := models.KeepAlive{}
	response.Message = "ok"
	c.JSON(http.StatusOK, &response)
}
