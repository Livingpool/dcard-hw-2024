# Base image for building the go project
FROM golang:1.20-alpine AS build

RUN mkdir /src
WORKDIR /src
COPY . /src/

# Builds the current project to a binary file called api
RUN go build -o ./out/api

# Exposes the 8080 port from the container
EXPOSE 8080

# Runs the binary once the container starts
CMD ["./out/api"]