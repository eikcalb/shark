package inventory

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"eikcalb.dev/shark/src/constants"
	"github.com/gin-gonic/gin"
)

// startServer starts a server for the service.
// @ref https://github.com/gin-gonic/gin?tab=readme-ov-file#running-gin
func (i *Inventory) startServer(ctx context.Context, port uint16) {
	log.Info("Starting web server", "port", port)
	defer func() {
		log.Info("Stopped web server")
	}()

	var rg *gin.RouterGroup
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	appVersion := ctx.Value(constants.CONTEXT_APPLICATION_VERSION_KEY)
	if appVersion == nil {
		// no app version specified.
		rg = r.Group("/inventory")
	} else {
		rg = r.Group(fmt.Sprintf("/%s/inventory", appVersion))
	}

	// Fetch all inventory items.
	rg.GET("/", func(c *gin.Context) {
		data := i.serialize()
		c.JSON(http.StatusOK, gin.H{"response": data})
	})

	rg.PUT("/:id", func(c *gin.Context) {
		// When an update is received for an item, parse the request body.
		id := c.Param("id")
		var json []Pack
		if err := c.BindJSON(&json); err != nil {
			return
		}

		// We have received an id and a pack array, so we will need to update
		// the item.
		// New item, so we just save.
		newPackSet := NewPackSet()
		for _, pack := range json {
			log.Info("Add new pack to inventory", "pack", pack)
			err := newPackSet.Add(pack)
			if err != nil {
				log.Error("Failed to add new pack to inventory", "pack", pack, "error", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
				return
			}
		}
		newPackSet.Sort()
		i.data[id] = *newPackSet

		// Save the update to disk without blocking this request.
		go func() {
			i.persist()
		}()

		data := i.serialize()
		c.JSON(http.StatusOK, gin.H{"response": data})
	})

	rg.GET("/:id/order/:count", func(c *gin.Context) {
		// When an update is received for an item, parse the request body.
		id := c.Param("id")
		rawCount := c.Param("count")
		count, err := strconv.Atoi(rawCount)
		if err != nil {
			log.Error("Failed to parse order count", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		var data gin.H = gin.H{}
		packsOrder := i.ProcessOrder(id, count)
		for pack, count := range packsOrder {
			data[strconv.Itoa(int(pack.Size))] = count
		}
		c.JSON(http.StatusOK, gin.H{"response": data})
	})

	r.Run(fmt.Sprintf(":%d", port))
}
