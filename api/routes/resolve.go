package routes

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/micastar/shorten-url/database"
	"github.com/micastar/shorten-url/helpers"
	"github.com/redis/go-redis/v9"
)

func ResolveURL(r *gin.Context) {
	ctx := context.Background()

	url := r.Param("url")

	rc := database.CreateClient(helpers.RedisDatabaseMain)
	defer rc.Close()

	value, err := rc.Get(ctx, url).Result()

	switch {
	case err == redis.Nil:
		r.JSON(http.StatusNotFound, gin.H{"error": "short not found in the database"})
		return
	case err != nil:
		r.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	Incr := database.CreateClient(helpers.RedisDatabaseIncr)
	defer Incr.Close()

	Incr.Incr(ctx, "counter")

	r.Redirect(http.StatusMovedPermanently, value)
}
