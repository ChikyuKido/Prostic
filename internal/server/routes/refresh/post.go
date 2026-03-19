package refresh

import (
	"net/http"

	"github.com/gin-gonic/gin"

	cacheservice "prostic/internal/service/cache"
)

func refreshAll(c *gin.Context) {
	result, err := cacheservice.RefreshAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh cache"})
		return
	}

	c.JSON(http.StatusOK, result)
}
