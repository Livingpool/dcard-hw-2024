package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Livingpool/constants"
	"github.com/Livingpool/model"
	"github.com/redis/go-redis/v9"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ApiGroup struct {
	Collection  *mongo.Collection
	RedisClient *redis.Client
}

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// SetAdInRedis 將廣告cache在 Redis 中
func (r *ApiGroup) SetAdInRedis(ctx context.Context, searchAdRequest model.SearchAdRequest, adInMongoDB []model.SearchAdResponse) error {
	offset := ""
	limit := ""
	age := ""
	if searchAdRequest.Offset != 0 {
		offset = strconv.Itoa(searchAdRequest.Offset)
	}
	if searchAdRequest.Limit != 0 {
		limit = strconv.Itoa(searchAdRequest.Limit)
	}
	if searchAdRequest.Age != 0 {
		age = strconv.Itoa(searchAdRequest.Age)
	}

	cacheKey := fmt.Sprintf(constants.CacheKeyFormat, offset, limit, age, searchAdRequest.Gender, searchAdRequest.Country, searchAdRequest.Platform)

	if len(adInMongoDB) == 0 {
		_, err := r.RedisClient.Set(ctx, cacheKey, "", 10*time.Second).Result()
		if err != nil {
			return err
		}
		return nil
	}

	now := time.Now()
	if adInMongoDB[0].EndAt.Before(now) {
		return nil
	}

	adString, err := json.Marshal(adInMongoDB)
	if err != nil {
		return err
	}

	_, err = r.RedisClient.Set(ctx, cacheKey, string(adString), now.Sub(adInMongoDB[0].EndAt)).Result()
	if err != nil {
		return err
	}

	return nil
}

// CreateAd 產生廣告
func (r *ApiGroup) CreateAd(c *gin.Context) {
	var ad model.CreateAdRequest
	if err := c.ShouldBindJSON(&ad); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ShouldBindJSON error": err.Error()})
		return
	}

	// Validate the input
	if err := model.Validator().Struct(ad); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Validation error": err.Error()})
		return
	}

	// Convert startAt and endAt strings to time.Time
	startAt, err := time.Parse(time.RFC3339, ad.StartAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"time.Parse error": err.Error()})
		return
	}
	endAt, err := time.Parse(time.RFC3339, ad.EndAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"time.Parse error": err.Error()})
		return
	}

	// Use converted fields
	var adInMongoDB model.CreateAdInMongoDB
	adInMongoDB.Title = ad.Title
	adInMongoDB.StartAt = startAt
	adInMongoDB.EndAt = endAt
	// Use default values if not provided
	if ad.Conditions.AgeStart != 0 {
		adInMongoDB.Conditions.AgeStart = ad.Conditions.AgeStart
	} else {
		adInMongoDB.Conditions.AgeStart = 1
	}
	if ad.Conditions.AgeEnd != 0 {
		adInMongoDB.Conditions.AgeEnd = ad.Conditions.AgeEnd
	} else {
		adInMongoDB.Conditions.AgeEnd = 100
	}
	adInMongoDB.Conditions.Gender = ad.Conditions.Gender
	adInMongoDB.Conditions.Country = ad.Conditions.Country
	adInMongoDB.Conditions.Platform = ad.Conditions.Platform

	// Insert the document in mongodb with converted startAt and endAt fields
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = r.Collection.InsertOne(ctx, adInMongoDB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "creation successful"})
}

// SearchAd 列出符合可用和匹配目標條件的廣告
func (r *ApiGroup) SearchAd(c *gin.Context) {
	var search model.SearchAdRequest
	if err := c.ShouldBindQuery(&search); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ShouldBindQuery error": err.Error()})
		return
	}

	// Validate the input
	if err := model.Validator().Struct(search); err != nil {
		fmt.Printf("SearchAdRequest: %+v\n", search)
		c.JSON(http.StatusBadRequest, gin.H{"Validation error": err.Error()})
		return
	}

	now := time.Now()

	// Create a filter from the search parameters
	filter := bson.M{
		"startAt": bson.M{
			"$lte": now,
		},
		"endAt": bson.M{
			"$gte": now,
		},
	}

	if search.Age != 0 {
		filter["$and"] = bson.A{
			bson.M{
				"$or": bson.A{
					bson.M{"conditions.ageStart": bson.M{"$exists": false}},
					bson.M{"conditions.ageStart": bson.M{"$lte": search.Age}},
				},
			},
			bson.M{
				"$or": bson.A{
					bson.M{"conditions.ageEnd": bson.M{"$exists": false}},
					bson.M{"conditions.ageEnd": bson.M{"$gte": search.Age}},
				},
			},
		}
	}

	if search.Gender != "" {
		filter["conditions.gender"] = bson.M{
			"$in": []interface{}{search.Gender, nil},
		}
	}

	if search.Country != "" {
		filter["conditions.country"] = bson.M{
			"$in": []interface{}{search.Country, nil},
		}
	}

	if search.Platform != "" {
		filter["conditions.platform"] = bson.M{
			"$in": []interface{}{search.Platform, nil},
		}
	}

	// Create a projection to return only the title and endAt fields
	projection := bson.D{
		{Key: "title", Value: 1},
		{Key: "endAt", Value: 1},
	}

	// Create options to sort by endAt in ascending order & implement pagination
	opts := options.Find().SetSort(bson.D{{Key: "endAt", Value: 1}})
	if search.Offset != 0 {
		opts.SetSkip(int64(search.Offset))
	}
	if search.Limit != 0 {
		opts.SetLimit(int64(search.Limit))
	} else {
		opts.SetLimit(5) // default limit
	}

	// Find matching documents
	cursor, err := r.Collection.Find(context.Background(), filter, opts.SetProjection(projection))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Mongo Collection.Find error": err.Error()})
		return
	}

	// Iterate through the cursor and decode each document
	var ads []model.SearchAdResponse
	if err = cursor.All(c, &ads); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Mongo cursor.All error": err.Error()})
		return
	}

	// Cache the ads in redis
	if err = r.SetAdInRedis(c, search, ads); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Caching error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": ads})
}

// [Testing function] ReturnAllAds 回傳所有廣告
func (r *ApiGroup) ReturnAllAds(c *gin.Context) {
	// Create an empty filter to match all documents
	filter := bson.M{}

	// Create options to sort by endAt in ascending order & implement pagination
	opts := options.Find().
		SetSort(bson.D{{Key: "endAt", Value: 1}})

	// Find all documents
	cursor, err := r.Collection.Find(context.Background(), filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Iterate through the cursor and decode each document
	var ads []model.SearchAdResponse
	if err = cursor.All(context.Background(), &ads); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": ads})
}
