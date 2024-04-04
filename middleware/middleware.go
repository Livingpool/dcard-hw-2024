package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/Livingpool/constants"
	"github.com/Livingpool/model"
)

func CacheByRequestQuery(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the URL and parse it
		url, err := url.Parse(c.Request.RequestURI)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Url parsing error at middleware": err.Error()})
			c.Abort()
			return
		}

		// Parse the query parameters and format them into a cache key
		query := url.Query()
		cacheKey := fmt.Sprintf(constants.CacheKeyFormat, query.Get("offset"), query.Get("limit"), query.Get("age"),
			query.Get("gender"), query.Get("country"), query.Get("platform"))

		content, err := redisClient.Get(c, cacheKey).Result()

		// Not in cache or expired
		if err != nil {
			log.Println("Cache miss")
			c.Next()
			return
		}

		// Unmarshal the content from cache
		var ads []model.SearchAdResponse
		if content == "" {
			c.JSON(http.StatusOK, gin.H{"items": ads})
			c.Abort()
			return
		}
		if err := json.Unmarshal([]byte(content), &ads); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Unmarshal error": err.Error()})
			c.Abort()
			return
		}

		// In cache
		log.Println("Cache hit")
		c.JSON(http.StatusOK, ads)
		c.Abort()
	}
}
