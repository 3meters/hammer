// Http client hammer for stress-testing web services

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

const contentJson = "application/json"
const version = "0.1.2"

var configFileName string
var helpMe bool
var versionMe bool

type Request struct {
	Method string          `json:"method"`
	Url    string          `json:"url"`
	Body   json.RawMessage `json:"body,omitempty"`
}

type Config struct {
	Host   string
	Signin struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		InstallId string `json:"installId"`
	}
	TestParams  TestParams
	Cred        string // Set after authentication
	Hammers     int
	Seconds     int
	MaxProcs    int
	RequestPath string
	Timeout     int
	Log         int // Log requsests taking more the Log miliseconds, -1 for no logging
	LogMax      int // max log entries
}

// These params separate test runs from each other so that
// Unique keys and locations are moved for each test run
type TestParams struct {
	Seed string
	Lat  string
	Lng  string
}

// Module globals
var config Config
var hammerLog *os.File
var cLogged int

type Result struct {
	Runs      int
	Succede   int
	Fail      int
	ByteCount int64
	Times     Times
	Timeouts  Timeouts
}

type Time struct {
	Tag      string
	Reported int
	Measured int
}

type Times []Time

// Sorter interfaces for Times
func (t Times) Len() int           { return len(t) }
func (t Times) Swap(i, j int)      { t[j], t[i] = t[i], t[j] }
func (t Times) Less(i, j int) bool { return t[i].Measured < t[j].Measured }

type Timeout struct {
	Tag    string
	Method string
	Url    string
}

type Timeouts []Times

// Set command line flags
func init() {
	// Seed rand with current nanoseconds
	rand.Seed(time.Now().UTC().UnixNano())
	flag.StringVar(&configFileName, "config", "config.json", "config file")
	flag.StringVar(&configFileName, "c", "config.json", "config file")
	flag.BoolVar(&helpMe, "help", false, "help")
	flag.BoolVar(&helpMe, "h", false, "help")
	flag.BoolVar(&versionMe, "version", false, "version")
	flag.BoolVar(&versionMe, "v", false, "version")
}

func main() {

	flag.Parse()

	if helpMe {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if versionMe {
		fmt.Println(version)
		os.Exit(0)
	}

	fmt.Println("Reading config file " + configFileName)

	content, err := ioutil.ReadFile(configFileName)
	if err != nil {
		log.Fatal(err)
	}

	config = Config{
		Hammers:     1,
		Seconds:     5,
		MaxProcs:    1,
		RequestPath: "request.log",
		Timeout:     0,
		Log:         -1,
		LogMax:      1000,
	}

	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Fatal("Config file not valid JSON: ", err)
	}

	fmt.Printf("Config:\n%s\n", sprintJson(content))

	// Open and parse the request log that will be fired at the target
	requestFile, err := os.Open(config.RequestPath)

	if err != nil {
		log.Fatalln("Could not read request file "+config.RequestPath, err)
	}
	defer requestFile.Close()

	// Open the request log
	requests, err := parseRequestLog(requestFile)
	if err != nil {
		log.Fatal("Error parsing "+config.RequestPath+": ", err)
	}

	// Create the hammer log file
	hammerLog, err = os.Create("hammer.log")
	if err != nil {
		log.Fatal("Could not create hammer.log")
	}
	defer hammerLog.Close()
	cLogged = 0

	// Configure a transport that accepts self-singed certificates
	// Similar to curl --insecure
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	// Make sure we can reach the host
	if config.Host == "" {
		log.Fatalln("config.Host is required")
	}
	fmt.Print("Pinging " + config.Host + "... ")
	res, err := client.Get(config.Host)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		log.Fatal("Host response status code: ", res.StatusCode)
	}

	// Make sure the host returns something
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	if len(body) < 1 {
		log.Fatal("Host returned no data", res.StatusCode)
	}
	fmt.Println("Ok")

	// Autheticate and add the user credentail query string to config
	fmt.Print("Authenticating admin... ")
	config.Cred, err = authenticate(client, &config)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Ok")

	// Set max procs
	runtime.GOMAXPROCS(config.MaxProcs)

	// Start the results collector service
	ch := make(chan Result, config.Hammers)
	go sum(ch, config.Hammers)

	// Start the hammers with a small stagger
	for i := 0; i < config.Hammers; i++ {
		fmt.Println("Starting hammer ", i)
		go run(client, requests, ch)
		time.Sleep(25 * time.Millisecond)
	}

	// Infinite loop to prevent exit
	for {
		fmt.Print(".")
		time.Sleep(5 * time.Second)
	}
}

// parseRequestLog: parse our modified csv log format
func parseRequestLog(file *os.File) ([]Request, error) {

	const max = 10000
	requests := []Request{}
	lineCount := 0
	reader := csv.NewReader(file)
	reader.Comma = 0 // ignore commas, one field per line
	reader.FieldsPerRecord = 1
	reader.LazyQuotes = true

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break // done
		} else if err != nil {
			return nil, err
		}

		lineCount++
		if lineCount > max {
			return requests, fmt.Errorf("Request log exceeded max of %v", max)
		}

		recordBytes := []byte(record[0])
		request := Request{}
		err = json.Unmarshal(recordBytes, &request)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	fmt.Printf("%s%v%s\n", "Parsed ", lineCount, " requests Ok")
	return requests, nil
}

// sprintJson: prettyPrint JSON to stdout
func sprintJson(data []byte) string {
	var indented bytes.Buffer
	err := json.Indent(&indented, data, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%v\n", indented.String())
}

// Authenticate the user specified in config.json
func authenticate(client *http.Client, config *Config) (string, error) {

	// Attempt to sign in
	url := config.Host + "/v1/auth/signin"

	reqBodyBytes, _ := json.Marshal(config.Signin)

	res, err := client.Post(url, contentJson, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	type Session struct {
		UserId    string `json:"_owner"`
		SessionId string `json:"key"`
	}

	type Body struct {
		Session Session `json:"session"`
	}

	body := Body{}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if res.StatusCode != 200 {
		return "", fmt.Errorf("Authentication failed with status %v", res.StatusCode)
	}

	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		return "", err
	}

	credentials := "user=" + body.Session.UserId + "&session=" + body.Session.SessionId

	return credentials, nil
}

// Generate a new random request seed and location
func genTestParams() TestParams {
	return TestParams{
		Seed: strconv.FormatInt(rand.Int63n(100000000), 10),
		Lat:  strconv.Itoa((rand.Int() % 179) - 89),
		Lng:  strconv.Itoa((rand.Int() % 359) - 179),
	}
}

// run: fire requests at the target with config credentials
func run(client *http.Client, requests []Request, ch chan Result) {

	result := Result{}

	stop := false

	// Start the clock
	go func() {
		time.Sleep(time.Duration(config.Seconds) * time.Second)
		stop = true
	}()

	runParams := TestParams{}

	cReqs := len(requests)

	for i := 0; stop == false; i++ {

		if i >= cReqs { // start over sending requests from the request log
			result.Runs++
			i = 0
		}

		// All requests within a request log share a random seed in
		// fields that are required to be unique in the db.  For each
		// run of the hammer, replace those params with new, randomly
		// generated ones
		if i == 0 {
			runParams = genTestParams()
		}

		logReq := requests[i]

		delim := "?"
		if strings.Contains(logReq.Url, "?") {
			delim = "&"
		}

		method := strings.ToUpper(logReq.Method)
		url := config.Host + logReq.Url + delim + config.Cred

		// Replace the seed in urls with our newly generated seed
		url = strings.Replace(url, config.TestParams.Seed, runParams.Seed, -1)

		// Same for the body
		reqBody := bytes.Replace(logReq.Body, []byte(config.TestParams.Seed), []byte(runParams.Seed), -1)

		// Move request location to another latitude
		target := []byte("\"lat\":" + config.TestParams.Lat)
		replace := []byte("\"lat\":" + runParams.Lat)
		reqBody = bytes.Replace(reqBody, target, replace, -1)

		// Same with longitude
		target = []byte("\"lng\":" + config.TestParams.Lng)
		replace = []byte("\"lng\":" + runParams.Lng)
		reqBody = bytes.Replace(reqBody, target, replace, -1)

		// Create a request
		req, reqErr := http.NewRequest(method, url, bytes.NewReader(reqBody))
		if reqErr != nil {
			log.Fatal(reqErr)
		}
		req.Header.Set("Content-Type", contentJson)

		// Start the request timer
		before := time.Now().UnixNano()
		res, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()

		// Record the measured time the response took to return
		after := time.Now().UnixNano()

		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}

		if 200 <= res.StatusCode && 400 > res.StatusCode {
			result.Succede++
		} else {
			result.Fail++
		}
		result.ByteCount += int64(len(bodyBytes))

		// Parse the request tag and reported time from the response body
		body := struct {
			Tag  string  `json:"tag"`
			Time float32 `json:"time"`
		}{} // anonymous struct type

		err = json.Unmarshal(bodyBytes, &body)
		if err != nil {
			log.Fatal(err)
		}

		// Add the respone time to the result
		time := Time{
			Tag:      body.Tag,
			Reported: int(body.Time),                  // miliseconds
			Measured: int((after - before) / 1000000), // miliseconds from nanoseconds
		}
		result.Times = append(result.Times, time)

		if config.Log >= 0 && config.Log < time.Measured && cLogged < config.LogMax {
			logEntry := []byte(fmt.Sprintf("Tag: %s, Reported: %d, Measured: %d\n", time.Tag, time.Reported, time.Measured))
			logEntry = append(logEntry, []byte(fmt.Sprintf("%d %s %s\n", res.StatusCode, method, url))...)
			logEntry = append(logEntry, []byte(fmt.Sprintf("%s\n\n", sprintJson(reqBody)))...)
			fmt.Printf("%s", logEntry)
		}
	}

	ch <- result
}

// sum: read and sum the results from a result channel
func sum(ch chan Result, expected int) {

	// create an aggregate result
	total := Result{}

	// Compute the result as returned by each chanel
	for i := 0; i < expected; i++ {
		result := <-ch
		total.Runs += result.Runs
		total.Succede += result.Succede
		total.Fail += result.Fail
		total.ByteCount += result.ByteCount
		total.Times = append(total.Times, result.Times...) // hmm, not sure I need the ...
		total.Timeouts = append(total.Timeouts, result.Timeouts...)
	}

	close(ch)

	// Compute some stats
	failRate := float32(total.Fail) / float32(total.Succede+total.Fail)
	sort.Sort(total.Times)
	min := total.Times[0].Measured
	max := total.Times[len(total.Times)-1].Measured
	median := total.Times[len(total.Times)/2].Measured
	sumMeasured := 0
	sumReported := 0
	for i := range total.Times {
		sumReported += total.Times[i].Reported
		sumMeasured += total.Times[i].Measured
	}
	meanMeasured := int(sumMeasured / len(total.Times))
	meanLatency := int((sumMeasured - sumReported) / len(total.Times))

	fmt.Printf("\n\nResults: \n\n")
	fmt.Printf("Seconds: %v\n", config.Seconds)
	fmt.Printf("Runs: %v\n", total.Runs)
	fmt.Printf("Requests: %v\n", total.Succede+total.Fail)
	fmt.Printf("Errors: %v\n", total.Fail)
	fmt.Printf("Timeouts: %v\n", len(total.Timeouts))
	fmt.Printf("Fail Rate: %.2f\n", failRate)
	fmt.Printf("KBytes per second: %v\n", total.ByteCount/int64(config.Seconds)/1000)
	fmt.Printf("Requests per second: %v\n", (total.Succede+total.Fail)/config.Seconds)
	fmt.Printf("Min time: %v\nMax time: %v\nMean time: %v\nMean latency: %v\nMedian time: %v\n\n", min, max, meanMeasured, meanLatency, median)

	os.Exit(0)
}
