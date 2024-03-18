package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	model "github.com/Livingpool/model"

	"github.com/biter777/countries"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type HandlerStruct struct {
	DB       *mongo.Database
	Validate *validator.Validate
}

// Validation functions
func isCountryCode(fl validator.FieldLevel) bool {
	country := fl.Field().String()
	if country == "" {
		return true
	}
	countryCode := countries.ByName(country)
	return countryCode != countries.Unknown
}

func validateGender(fl validator.FieldLevel) bool {
	gender := fl.Field().String()
	switch model.Gender(gender) {
	case model.Male, model.Female, "":
		return true
	default:
		return false
	}
}

func validatePlatform(fl validator.FieldLevel) bool {
	platform := fl.Field().String()
	switch model.Platform(platform) {
	case model.Android, model.IOS, model.Web, "":
		return true
	default:
		return false
	}
}

func validateAgeRange(fl validator.FieldLevel) bool {
	ageStart := fl.Parent().FieldByName("AgeStart").Int()
	ageEnd := fl.Parent().FieldByName("AgeEnd").Int()
	return ageStart == 0 || ageEnd == 0 || ageStart <= ageEnd
}

func Handler(db *mongo.Database) *HandlerStruct {
	validate := validator.New()
	validate.RegisterValidation("countrycode", isCountryCode)
	validate.RegisterValidation("gender", validateGender)
	validate.RegisterValidation("platform", validatePlatform)
	validate.RegisterValidation("ageRange", validateAgeRange)

	return &HandlerStruct{
		DB:       db,
		Validate: validate,
	}
}

// CreateIndex for endAt field (ascending order) to optimize the search
func (h *HandlerStruct) CreateIndex(coll string) error {
	collection := h.DB.Collection(coll)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.M{
				"endAt": 1, // 1 for ascending order, -1 for descending order
			},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// CreateAd 產生廣告
func (h *HandlerStruct) CreateAd(c *gin.Context, coll string) {
	var ad model.CreateAdRequest
	if err := c.ShouldBindJSON(&ad); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ShouldBindJSON error": err.Error()})
		return
	}

	// Validate the input
	if err := h.Validate.Struct(ad); err != nil {
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

	// Insert the document with converted startAt and endAt fields
	collection := h.DB.Collection(coll)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = collection.InsertOne(ctx, adInMongoDB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "creation successful"})
}

// SearchAd 列出符合可用和匹配目標條件的廣告
func (h *HandlerStruct) SearchAd(c *gin.Context, coll string) {
	var search model.SearchAdRequest
	if err := c.ShouldBindQuery(&search); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ShouldBindQuery error": err.Error()})
		return
	}

	// Validate the input
	if err := h.Validate.Struct(search); err != nil {
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

	// Use h.Client to interact with the database
	collection := h.DB.Collection(coll)
	cursor, err := collection.Find(context.Background(), filter, opts.SetProjection(projection))
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

// [Testing function] ReturnAllAds 回傳所有廣告
func (h *HandlerStruct) ReturnAllAds(c *gin.Context, coll string) {
	// Use h.Client to interact with the database
	collection := h.DB.Collection(coll)

	// Create an empty filter to match all documents
	filter := bson.M{}

	// Create options to sort by endAt in ascending order & implement pagination
	opts := options.Find().
		SetSort(bson.D{{Key: "endAt", Value: 1}})

	// Find all documents
	cursor, err := collection.Find(context.Background(), filter, opts)
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
