package apitest

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/Livingpool/api"
	"github.com/Livingpool/model"
	"github.com/Livingpool/router"
	"github.com/gin-gonic/gin"

	"github.com/steinfletcher/apitest"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	testDBInstance *mongo.Database
	r              *gin.Engine
)

type CreateAdTestCase struct {
	Name       string                `json:"name"`
	Request    model.CreateAdRequest `json:"request"`
	StatusCode int                   `json:"statusCode"`
}

type SearchAdTestCase struct {
	Name       string                `json:"name"`
	Request    model.SearchAdRequest `json:"request"`
	StatusCode int                   `json:"statusCode"`
	Response   struct {
		Items []model.SearchAdResponse `json:"items"`
	} `json:"response"`
}

// Initial setup
func TestMain(m *testing.M) {
	log.Println("setup is running")
	testDB := SetupTestDatabase()
	if testDB == nil {
		log.Fatal("Failed to set up test database")
	}
	testDBInstance = testDB.DbInstance

	// Set up the router
	h := api.Handler(testDBInstance)
	r = router.Initialize(h, "advertisements")

	exitCode := m.Run()
	log.Println("teardown is running")
	testDB.TearDown()
	os.Exit(exitCode)
}

// E2E test for CreateAd
func TestCreateAd(t *testing.T) {
	// Read the test cases from the JSON file
	file, _ := os.Open("testcases/create_ad.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	var testCases []CreateAdTestCase
	if err := decoder.Decode(&testCases); err != nil {
		t.Fatal(err)
	}

	// Run a subtest for each test case
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			requestBodyStr, err := json.Marshal(tc.Request)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			apitest.New().
				// Debug().
				Handler(r).
				Post("/api/v1/ad").
				Header("content-type", "application/json").
				Body(string(requestBodyStr)).
				Expect(t).
				Status(tc.StatusCode).
				End()
		})
	}
}

// E2E test for SearchAd
func TestSearchAd(t *testing.T) {
	// Read the test cases from the JSON file
	file, _ := os.Open("testcases/search_ad.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	var testCases []SearchAdTestCase
	if err := decoder.Decode(&testCases); err != nil {
		t.Fatal(err)
	}

	// Populate the database with test data
	populateDBFromJSON(testDBInstance, "testcases/populate_db.json")

	// Run a subtest for each test case
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Unmarshal the JSON object to a map
			requestBodyStr, err := json.Marshal(tc.Request)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			// Convert the map to a query string
			requestBodyMap := make(map[string]interface{})
			if err := json.Unmarshal(requestBodyStr, &requestBodyMap); err != nil {
				t.Fatalf("Failed to unmarshal request body: %v", err)
			}

			// Convert the map to a QueryParams map
			queryParams := make(map[string]string)
			for key, value := range requestBodyMap {
				queryParams[key] = fmt.Sprintf("%v", value)
			}

			// Convert response to string
			responseStr, err := json.Marshal(tc.Response)
			if err != nil {
				t.Fatalf("Failed to marshal response: %v", err)
			}

			apitest.New().
				// Debug().
				Handler(r).
				Get("/api/v1/ad").
				QueryParams(queryParams).
				Header("content-type", "application/json").
				Expect(t).
				Status(tc.StatusCode).
				Body(string(responseStr)).
				End()
		})
	}
}
