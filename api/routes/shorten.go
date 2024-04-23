package routes

import (
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
	_, err := rc.Get(database.Ctx, c.RemoteIP()).Result()
	if err == redis.Nil {
		_ = rc.Set(database.Ctx, c.RemoteIP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else {
		val, _ := rc.Get(database.Ctx, c.RemoteIP()).Result()
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

	// enforce http, https, SSL

	body.URL = helpers.EnforceHTTP(body.URL)

	var id string

	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	r := database.CreateClient(0)
	defer r.Close()

	val, _ := r.Get(database.Ctx, id).Result()
	if val != "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "URL custom short is already in use"})
	}

	if body.Expiry == 0 {
		body.Expiry = 24
	}

	err = r.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to server"})
	}

	resp := response{
		URL:            body.URL,
		CustomShort:    "",
		Expiry:         body.Expiry,
		XRateRemaing:   10,
		XRateLimitRest: 30,
	}

	rc.Decr(database.Ctx, c.RemoteIP())

	val, _ = rc.Get(database.Ctx, c.RemoteIP()).Result()
	resp.XRateRemaing, _ = strconv.Atoi(val)

	ttl, _ := rc.TTL(database.Ctx, c.RemoteIP()).Result()
	resp.XRateLimitRest = ttl / time.Nanosecond / time.Minute

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	c.JSON(http.StatusOK, resp)
}
