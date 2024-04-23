package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/micastar/shorten-url/database"
	"github.com/redis/go-redis/v9"
)

func ResolveURL(r *gin.Context) {
	url := r.Param("url")

	rc := database.CreateClient(0)
	defer rc.Close()

	value, err := rc.Get(database.Ctx, url).Result()
	if err == redis.Nil {
		r.JSON(http.StatusNotFound, gin.H{"error": "short not found in the database"})
	} else if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{"error": "cannot connect to DB"})
	}

	rInt := database.CreateClient(1)
	defer rInt.Close()

	_ = rInt.Incr(database.Ctx, "counter")

	r.Redirect(301, value)
}
