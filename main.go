package main

import (
	"github.com/Livingpool/constants"
	"github.com/Livingpool/database"
	"github.com/Livingpool/router"
)

func main() {
	// Set up database
	db := database.ConstructDB()
	defer db.Disconnect()

	// Set up router
	r := router.Initialize(db.GetCollection(constants.DATABASE_NAME, constants.COLLECTION_NAME), db.RedisClient)
	r.Run(":5000")
}
