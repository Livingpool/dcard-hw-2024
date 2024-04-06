# Overview 
[![Docker Build, Unit Test, and Deploy to GKE](https://github.com/Livingpool/dcard-hw-2024/actions/workflows/build-test-deploy.yaml/badge.svg)](https://github.com/Livingpool/dcard-hw-2024/actions/workflows/build-test-deploy.yaml) \
This is the 2024 Backend Intern Assignment for Dcard🔥.\
🔗[Assignment Document Link](https://drive.google.com/file/d/1dnDiBDen7FrzOAJdKZMDJg479IC77_zT/view)

### Requirements 
- [x] Admin API for creating advertisements
  - < 3000 requests each day
  - < 1000 active ads in the database (StartAt < Now < EndAt)
- [x] Public API for searching for matching advertisements
  - Must handle 10,000 requests per second
- [x] No authentication required
- [x] Adequate unit tests
      
### Tech Stack 
- Golang Gin-Gonic framework
- Database: MongoDB
- Cache: Redis
  
### Implementation
- #### Choosing Go Gin
  - Gin is a lightweight, easy-to-use web framework and I am familiar with it. But starting from go 1.22,
    advanced routing is provided in the standard net/http package, so I may try that for my next project.
- #### Handling high load for Public API
  - This is the main challenge. Handling high load ususally involves caching and/or load balancing.
- #### Cache strategy
  - There are 4 main cache strategies: cache-aside, write-through, write-behind, and read-through. For our
    advertising application, I think there are 2 options as follows. I chose Cache-aside for implementation.
    - Write-through/write-behind: Pre-populate cache with valid ads when Admin API is called.
      - Pros: Can potentially support query filtering.
      - Cons: Much harder to implement as this involves using RedisJSON and this might not be a good practice.
    - Cache-aside: Set [Key:url params, Value:search result] as cache when Public API is called
      - Pros: Easier to implement.
      - Cons: Less resilient to different queries.
- #### Deployment on GKE
  - I used a 3-node distributed system (machine: e2-medium), each containing a Go server, a MongoDB, and a Redis instance. 
  - 1 node (machine: e2-standard-2) for an NFS server and client to facilitate ReadWriteMany operation for the 3 MongoDB instances. 
  - Ingress with a static IP to serve as a load balancer. 
  - Note: due to resource limitations, this server might not be able to handle high load.
- #### Unit tests
  - I used [Testcontainers](https://testcontainers.com) to set up MongoDB and Redis.
    The testcases are in api/api_test/testcases.
    Test functions TestCreateAd and TestSearchAd use the created db clients to test if the APIs give correct responses w.r.t. the testcases.
  - Implemented CI/CD workflow and semantic versioning using GitHub Actions.

### Benchmarks
I used Vegeta as the load testing tool. The testing scripts are in load_testing/vegeta/create_ad and loadtesting/vegeta/search_ad.
As the server's ability to handle load largely depends on the capabilities and resources of the underlying machine, benchmarking results may vary.
My machine: Mac Air M1. OS: macOS Sonoma 14.4.1. RAM: 8GB. You may need to manually set a higher value of allowed file descriptors on macOS/Linux.
![Screenshot 2024-04-05 at 10 01 10 AM](https://github.com/Livingpool/dcard-hw-2024/assets/52132459/bc4e89c1-2ee1-4471-89b5-37478da2d187)

### Running the project locally
- Start docker daemon, run the following command, and the endpoints can be accessed at `localhost:8080`
```bash
docker-compose up --build
```
- Run unit tests
```bash
cd api/api_test
go test
```
- Run load tests (create_ad is for populating the database)
```bash
cd load_testing/vegeta/create_ad
bash run.sh
```
```bash
cd load_testing/vegeta/search_ad
bash run.sh
```
- You can use this command to view the number of TIME_WAIT states on port 8080
```bash
netstat -na | grep 8080 | grep "TIME_WAIT" | wc -l
```
