# Overview 
This is the 2024 Backend Intern Assignment for Dcard.\
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
- Load balancing: Nginx
  
### Implementation
- Choosing Go Gin
  - Gin is a lightweight, easy-to-use web framework and I am familiar with it. But starting from go 1.22,
    advanced routing is provided in the standard net/http package, so I may try that for my next project.
- Handling high load for Public API
  - This is the main challenge.
- Cache strategy
  - There are 4 main cache strategies: cache-aside, write-through, write-behind, and read-through. For our
    advertising application, I think there are 2 options:
    - Write-through/write-behind: Pre-populate cache with valid ads when Admin API is called.
      - Pros: 
      - Cons: 
    - Cache-aside: Set [Key:url params, Value:search result] as cache when Public API is called
      - Pros:
      - Cons: 
- Load balancing strategy
  - I used Nginx as a reverse proxy to distribute load between several replicas of the Go server.
- Unit tests

### Benchmarks
I used Vegeta as the load testing tool. The testing scripts are in load_testing/vegeta/create_ad and loadtesting/vegeta/search_ad


### Running the project
- Start docker daemon, run the following command, and the endpoints can be accessed at `localhost:80`
```
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
### Future work
