package apitest

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TestDatabase struct {
	DbInstance     *mongo.Database
	DbAddress      string
	MongoContainer testcontainers.Container
	RedisContainer testcontainers.Container
	RedisClient    *redis.Client
}

// set up a test database
func SetupTestDatabase() *TestDatabase {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	container, dbInstance, dbAddr, err := createMongoContainer(ctx)
	if err != nil {
		log.Fatal("failed to setup test", err)
	}

	redisContainer, redisClient, _, err := createRedisContainer(ctx)
	if err != nil {
		log.Fatal("failed to setup test", err)
	}

	return &TestDatabase{
		MongoContainer: container,
		DbInstance:     dbInstance,
		DbAddress:      dbAddr,
		RedisContainer: redisContainer,
		RedisClient:    redisClient,
	}
}

// tear down the test database after testing
func (tdb *TestDatabase) TearDown() {
	_ = tdb.MongoContainer.Terminate(context.Background())
	_ = tdb.RedisContainer.Terminate(context.Background())
}

// connect to a mongo client and return a database instance
func newMongoDatabase(uri string) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("mongodb ping error :%v", err)
	}

	db := client.Database("testdb")

	return db, nil
}

// create a mongo container
func createMongoContainer(ctx context.Context) (testcontainers.Container, *mongo.Database, string, error) {
	var env = map[string]string{
		"MONGO_INITDB_ROOT_USERNAME": "admin",
		"MONGO_INITDB_ROOT_PASSWORD": "pass",
		"MONGO_INITDB_DATABASE":      "testdb",
	}
	var port = "27017/tcp"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mongo",
			ExposedPorts: []string{port},
			Env:          env,
		},
		Started: true,
	}
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return container, nil, "", fmt.Errorf("failed to start mongo container: %v", err)
	}

	p, err := container.MappedPort(ctx, "27017")
	if err != nil {
		return container, nil, "", fmt.Errorf("failed to get mongo container external port: %v", err)
	}

	log.Println("mongo container ready and running at port: ", p.Port())

	uri := fmt.Sprintf("mongodb://admin:pass@localhost:%s/", p.Port())
	db, err := newMongoDatabase(uri)
	if err != nil {
		return container, db, uri, fmt.Errorf("failed to establish mongo database connection: %v", err)
	}

	return container, db, uri, nil
}

// connect to a redis client
func newRedisClient(uri string) (*redis.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := redis.NewClient(&redis.Options{
		Addr:     uri,
		Password: "",
		DB:       0,
	})

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	return client, nil
}

// create a redis container
func createRedisContainer(ctx context.Context) (testcontainers.Container, *redis.Client, string, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis",
			ExposedPorts: []string{"6379/tcp"},
		},
		Started: true,
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to start redis container: %v", err)
	}

	p, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to get redis container external port: %v", err)
	}

	log.Println("redis container ready and running at port: ", p.Port())

	uri := fmt.Sprintf("localhost:%s", p.Port())
	redisClient, err := newRedisClient(uri)
	if err != nil {
		return redisContainer, nil, uri, fmt.Errorf("failed to establish redis client connection: %v", err)
	}
	return redisContainer, redisClient, uri, nil
}
