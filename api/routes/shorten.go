package routes

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/micastar/shorten-url/database"
	"github.com/micastar/shorten-url/helpers"
	"github.com/redis/go-redis/v9"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL            string        `json:"url"`
	CustomShort    string        `json:"short"`
	Expiry         time.Duration `json:"expiry"`
	XRateRemaing   int           `json:"rate_limit"`
	XRateLimitRest time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *gin.Context) {
	body := new(request)

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot parse JSON"})
	}

	// implement rate limiting

	rc := database.CreateClient(1)
	defer rc.Close()
	val, err := rc.Get(database.Ctx, c.RemoteIP()).Result()
	if err == redis.Nil {
		_ = rc.Set(database.Ctx, c.RemoteIP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else {
		val, _ = rc.Get(database.Ctx, c.RemoteIP()).Result()
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			limit, _ := rc.TTL(database.Ctx, c.RemoteIP()).Result()
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Rate limit exceed", "rate_limit_rest": limit / time.Nanosecond / time.Minute})
		}
	}

	// check if the input if an actual URL

	if !govalidator.IsURL(body.URL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
	}

	// check for domain error

	if !helpers.RemoveDomainError(body.URL) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "you can't hack the system"})
	}

	// enforce https, SSL

	body.URL = helpers.EnforceHTTP(body.URL)

	rc.Decr(database.Ctx, c.RemoteIP())
}
