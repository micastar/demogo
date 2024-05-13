package routes

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	URL            string `json:"url"`
	CustomShort    string `json:"short"`
	Expiry         int    `json:"expiry"`
	XRateRemaing   int    `json:"rate_limit"`
	XRateLimitRest int    `json:"rate_limit_reset"`
}

func ShortenURL(c *gin.Context) {

	ctx := context.Background()

	body := new(request)

	url := c.Request.FormValue("url")

	// if err := c.BindJSON(&body); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "cannot parse JSON"})
	// 	return
	// }

	// set data
	r := database.CreateClient(helpers.RedisDatabaseMain)
	defer r.Close()

	// rate limiting
	Incr := database.CreateClient(helpers.RedisDatabaseIncr)
	defer Incr.Close()

	_, err := Incr.Get(ctx, c.RemoteIP()).Result()

	switch {
	case err == redis.Nil:
		Incr.Set(ctx, c.RemoteIP(), os.Getenv("API_QUOTA"), helpers.ApiQuotaTTL*60*time.Second).Err()
	case err == nil:
		// Get remaining count
		val, _ := Incr.Get(ctx, c.RemoteIP()).Result()

		valInt, _ := strconv.Atoi(val)

		if valInt <= 0 {
			limit, _ := Incr.TTL(ctx, c.RemoteIP()).Result()

			c.JSON(
				http.StatusServiceUnavailable,
				gin.H{
					"error":           "Rate limit exceed",
					"rate_limit_rest": limit / time.Nanosecond / time.Minute,
				},
			)
			return
		}
	}

	//check the input is actual URL
	if !govalidator.IsURL(url) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		return
	}

	//check for the domain error
	if !helpers.RemoveDomainError(url) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"eror": "Current URL equal to the local domain"})
		return
	}

	// enforce http, https, SSL
	body.URL = helpers.EnforceHTTP(url)

	var id string

	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	if body.Expiry == 0 {
		body.Expiry = 8
	}

	val, _ := r.Get(ctx, id).Result()
	if val != "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "URL custom short is already in use"})
		return
	}

	err = r.Set(ctx, id, body.URL, body.Expiry*60*time.Second).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to server"})
		return
	}

	resp := response{
		URL:    body.URL,
		Expiry: int(body.Expiry),
	}

	Incr.DecrBy(ctx, c.RemoteIP(), helpers.RateLimitDecrement)

	// Get Remaining Count
	val, _ = Incr.Get(ctx, c.RemoteIP()).Result()
	resp.XRateRemaing, _ = strconv.Atoi(val)

	ttl, _ := Incr.TTL(ctx, c.RemoteIP()).Result()
	resp.XRateLimitRest = int(ttl.Minutes())

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	c.JSON(http.StatusOK, resp)

}
