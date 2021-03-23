# Hashing API Code Challenge

## Getting Started

### Dependencies

Ensure that Go is installed on the machine running the binary if compiling the app

### First Time Setup

To compile the application, one can run 
`go build main.go`

This will create a binary called `main`

This binary can be executed from the root directory with the following command:
`./main`

This will start the server which listens on port :8080 for requests

# API Documentation

## Basics

This API contains four routes 

To send data to a POST route via `curl`, the following command can be used:
`curl -d "password=test" localhost:8080/hash`

## Routes

### POST `/hash`

This route will take a `password` input, hash the password using SHA256 and base64, and store the record
in an in-memory data store until the server is shutdown. The records will then be saved to disk

Returns the ID of the hash created by the server

**Request**

`curl -d "password=test" localhost:8080/hash`

**Response**
```json
{
    "Id": 7
}
```

### POST `/hash/:id`

This route will return the hash of the provided `id` if one is stored in local memory/disk.

**Request**

`curl -X POST localhost:8080/hash/7`

**Response**
```json
{
    "Hash": "v+nNU8e5m3vWWZkyXBIhEbNdRVUr+p5hpMEja/VUPnY="
}
```

### GET `/stats`

This route will return the total number of records created as well as the average amount of time to create all records, in microseconds

**Request**

`curl localhost:8080/stats`

**Response**
```json
{
    "Total": 10,
    "Average": 6
}
```

### GET/POST `/shutdown`

This route will gracefully shutdown the server and save all the records created during the session,
as well as the total amount of time it took to create all the records

**Request**

`curl localhost:8080/shutdown`
