package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	store        = map[int]string{}
	sessionTime  int64
	totalTime    int64
	savedRecords int
)

const (
	dbStoreName   	 = "hashStore.db"
	timeStoreName 	 = "time.db"
	port          	 = ":8080"
	baseErrorMessage = "Password invalid"
)

type Stats struct {
	Total   int
	Average int
}

type Hash struct {
	Id   int
	Hash string
}

/**
This route will accept a password field in a POST request,
Generate a unique ID, run a SHA256/base64 conversion in a
separate gofunc, and calculate the total time of operation
*/
func hash(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		fmt.Fprintf(w, "only POST method is supported for this route.")
		return
	}

	// This is a workaround to allow dynamic routing using ONLY the base
	// Go library. In a regular environment, one can use gorillamux or
	// httprouter to dynamically route addresses
	pathVars := strings.Split(req.URL.Path, "/")
	if len(pathVars) == 3 && pathVars[1] == "hash" {
		id := getIdFromHashUrl(pathVars)
		if id <= 0 {
			fmt.Fprintf(w, "Cannot GET id from URL, id not valid")
			return
		}
		hash := ""
		// Get the hash from the current in-memory data store if it exists,
		// otherwise, grab it from the database
		if val, ok := store[id]; ok {
			hash = val
		} else {
			hash = findHashInDatabase(id)
		}

		if hash == "" {
			fmt.Fprintf(w, "Cannot find hash for ID: %d", id)
			return
		}

		h := Hash{id, hash}
		outputStructToJsonResponse(w, h)
		return
	}

	req.ParseForm()
	pass := req.Form.Get("password")

	validPass, msg := validatePassword(pass)
	if !validPass {
		fmt.Fprint(w, msg)
		return
	}

	// current ID is generated as follows:
	// the highest value of all the IDs in the current session + 1
	// if there are records in the "database" then we add that to the
	// total value as well. This logic would most likely be handled
	// in an upsert if using an actual database
	newId := savedRecords + len(store) + 1
	store[newId] = ""
	record := Hash{newId, ""}
	if !outputStructToJsonResponse(w, record) {
		return
	}

	go func(id int) {
		// Wait 5 seconds before converting
		time.Sleep(5 * time.Second)
		start := time.Now()

		// Add the SHA to the store after 5 seconds has passed
		store[id] = convertPass(pass)

		// Calculate the amount of time it takes to do one operation
		// and add that value to the total amount of time
		end := time.Now()
		oppDiff := end.Sub(start)
		sessionTime += int64(oppDiff / time.Microsecond)
	}(newId)
}

/**
Converts a provided struct to JSON format and writes it to the response writer
*/
func outputStructToJsonResponse(w http.ResponseWriter, s interface{}) bool {
	js, err := json.Marshal(s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	return true
}

/**
Function that contains business logic to make sure a provided password is
valid and secure
*/
func validatePassword(pass string) (bool, string) {
	if pass == "" {
		return false, baseErrorMessage
	}

	if len(pass) > 10 {
		return false, "Password must be 10 or less characters in length"
	}

	// very very basic SQL injection prevention
	if strings.Contains(pass, "'") {
		return false, baseErrorMessage
	}
	return true, ""
}

/**
Combs through the "database" to find the hash for the provided id
Time Complexity : O(n) where n is the number of records in the db file
*/
func findHashInDatabase(id int) string {
	hash := ""
	idString := strconv.Itoa(id)
	f, err := os.Open(dbStoreName)
	if err != nil {
		log.Println(err)
		return hash
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		vals := strings.Split(line, " ")
		if vals[0] == idString {
			return vals[1]
		}
	}
	return hash
}

/**
Retrieves the ID from the final position in the query URL
*/
func getIdFromHashUrl(pathVars []string) int {
	id, err := strconv.Atoi(pathVars[len(pathVars)-1])
	if err != nil {
		log.Println(err)
		return -1
	}
	return id
}

/**
Generate the SHA256 for the password and convert it into base64
*/
func convertPass(pass string) string {
	sha := sha256.Sum256([]byte(pass))
	return base64.StdEncoding.EncodeToString(sha[:])
}

/**
This route returns the total number of records create for all session
and the average amount of time (in microseconds) a record takes to create
*/
func stats(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		fmt.Fprintf(w, "only GET method is supported for this route.")
		return
	}

	total := len(store) + savedRecords
	stats := Stats{
		total,
		(int(totalTime) + int(sessionTime)) / total,
	}
	outputStructToJsonResponse(w, stats)
}

func main() {
	log.Println("Server starting")
	mux := http.NewServeMux()
	srv := http.Server{Addr: port, Handler: mux}

	// Load up the number of records in the "database"
	// The would probably be a lot more intuitive if I were using
	// an actual database model instead of a file
	f, err := os.Open(dbStoreName)
	if err != nil {
		log.Println(err)
	}
	savedRecords = getLineCount(f)
	// closing this here since we need to use the "Database"
	// in other portions of code
	f.Close()
	loadTotalTime()

	// This function shuts down the server and saves the records
	// created during the session to 'disk'
	shutDown := func(w http.ResponseWriter, req *http.Request) {
		// Wait 10 seconds to wait for any ongoing processes
		time.Sleep(10 * time.Second)
		fmt.Fprint(w, "Shutting down server")
		saveHashStore()
		saveTotalTime()
		srv.Shutdown(context.Background())
	}

	mux.HandleFunc("/", hash)
	mux.HandleFunc("/hash", hash)
	mux.HandleFunc("/shutdown", shutDown)
	mux.HandleFunc("/stats", stats)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	log.Println("bye")
}

/**
This function serves as a helper function to retreive the number of lines
in a provided file.
*/
func getLineCount(file *os.File) int {
	fileScanner := bufio.NewScanner(file)
	lineCount := 0
	for fileScanner.Scan() {
		lineCount++
	}
	return lineCount
}

/**
Loads the total time taken to run ALL transactions
into memory from storage
*/
func loadTotalTime() {
	f2, err := os.Open(timeStoreName)
	if err != nil {
		log.Println(err)
		return
	}
	defer f2.Close()
	scanner := bufio.NewScanner(f2)

	allTimeStr := scanner.Text()
	for scanner.Scan() {
		allTimeStr = scanner.Text()
	}
	allTime, err := strconv.Atoi(allTimeStr)
	if err != nil {
		log.Println(err)
		return
	}
	totalTime = int64(allTime)
}

/**
Save the time as a larger "total time" used to calculate
average for ALL records and not just the ones
created during the session
*/
func saveTotalTime() {
	f2, err := os.OpenFile(timeStoreName,
		os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Println(err)
		return
	}
	defer f2.Close()
	total := strconv.Itoa(int(totalTime) + int(sessionTime))

	err = ioutil.WriteFile(timeStoreName, []byte(total), 0777)
	if err != nil {
		log.Fatalln(err)
		return
	}
}

/**
Save the map to a file to simulate a database transaction
*/
func saveHashStore() {
	f, err := os.OpenFile(dbStoreName,
		os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	for k, v := range store {
		line := strconv.Itoa(k) + " " + v + "\n"
		if _, err := f.WriteString(line); err != nil {
			log.Println(err)
		}
	}
}
