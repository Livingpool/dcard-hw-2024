package api

import (
	"context"
	"net/http"
	"time"

	model "go-project/model"

	"github.com/biter777/countries"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type HandlerStruct struct {
	Client   *mongo.Client
	Validate *validator.Validate
}

// Validation functions
func isCountryCode(fl validator.FieldLevel) bool {
	country := fl.Field().String()
	countryCode := countries.ByName(country)
	return countryCode != countries.Unknown
}

func validateGender(fl validator.FieldLevel) bool {
	gender := fl.Field().String()
	switch model.Gender(gender) {
	case model.Male, model.Female:
		return true
	default:
		return false
	}
}

func validatePlatform(fl validator.FieldLevel) bool {
	platform := fl.Field().String()
	switch model.Platform(platform) {
	case model.Android, model.IOS, model.Web:
		return true
	default:
		return false
	}
}

func validateAgeRange(fl validator.FieldLevel) bool {
	ageStart := fl.Parent().FieldByName("AgeStart").Int()
	ageEnd := fl.Parent().FieldByName("AgeEnd").Int()
	return ageStart < ageEnd
}

func Handler(client *mongo.Client) *HandlerStruct {
	validate := validator.New()
	validate.RegisterValidation("countrycode", isCountryCode)
	validate.RegisterValidation("gender", validateGender)
	validate.RegisterValidation("platform", validatePlatform)
	validate.RegisterValidation("ageRange", validateAgeRange)

	return &HandlerStruct{
		Client:   client,
		Validate: validate,
	}
}

// CreateAd 產生廣告
func (h *HandlerStruct) CreateAd(c *gin.Context, db string, coll string) {
	var ad model.CreateAdRequest
	if err := c.ShouldBindJSON(&ad); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the input
	if err := h.Validate.Struct(ad); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert startAt and endAt strings to time.Time
	startAt, err := time.Parse(time.RFC3339, ad.StartAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	endAt, err := time.Parse(time.RFC3339, ad.EndAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create an index for the endAt field to optimize the search
	collection := h.Client.Database(db).Collection(coll)
	_, err = collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.M{
				"endAt": 1, // 1 for ascending order, -1 for descending order
			},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Insert the document with converted startAt and endAt fields
	_, err = collection.InsertOne(context.Background(), bson.M{
		"title":   ad.Title,
		"startAt": startAt,
		"endAt":   endAt,
		"conditions": bson.M{
			"ageStart": ad.Conditions.AgeStart,
			"ageEnd":   ad.Conditions.AgeEnd,
			"gender":   ad.Conditions.Gender,
			"country":  ad.Conditions.Country,
			"platform": ad.Conditions.Platform,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "insertion successful"})
}

// SearchAd 列出符合可用和匹配目標條件的廣告
func (h *HandlerStruct) SearchAd(c *gin.Context, db string, coll string) {
	var search model.SearchAdRequest
	if err := c.ShouldBindQuery(&search); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the input
	if err := h.Validate.Struct(search); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		"conditions.ageStart": bson.M{
			"$lte": search.Conditions.Age,
		},
		"conditions.ageEnd": bson.M{
			"$gte": search.Conditions.Age,
		},
		"conditions.gender":   search.Conditions.Gender,
		"conditions.country":  search.Conditions.Country,
		"conditions.platform": search.Conditions.Platform,
	}

	// Create a projection to return only the title and endAt fields
	projection := bson.D{
		{Key: "title", Value: 1},
		{Key: "endAt", Value: 1},
	}

	// Create options to sort by endAt in ascending order & implement pagination
	opts := options.Find().
		SetSort(bson.D{{Key: "endAt", Value: 1}}).
		SetSkip(int64(search.Offset)).
		SetLimit(int64(search.Limit))

	// Use h.Client to interact with the database
	collection := h.Client.Database(db).Collection(coll)
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

// Testing function
func (h *HandlerStruct) ReturnAllAds(c *gin.Context, db string, coll string) {
	// Use h.Client to interact with the database
	collection := h.Client.Database(db).Collection(coll)

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
