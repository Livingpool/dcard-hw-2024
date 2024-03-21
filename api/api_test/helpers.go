package apitest

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/Livingpool/constants"
	"github.com/Livingpool/model"
	"go.mongodb.org/mongo-driver/mongo"
)

func populateDBFromJSON(testDBInstance *mongo.Database, filename string) error {
	// Open the JSON file
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode the file into a slice of CreateAdInMongoDB structs
	var ads []model.CreateAdInMongoDB
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&ads); err != nil {
		return err
	}

	// Convert the slice to a slice of interface{} for the InsertMany function
	docs := make([]interface{}, len(ads))
	for i, v := range ads {
		docs[i] = v
	}

	// Insert the documents into the database
	collection := testDBInstance.Collection(constants.COLLECTION_NAME)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = collection.InsertMany(ctx, docs)
	if err != nil {
		return err
	}

	return nil
}
