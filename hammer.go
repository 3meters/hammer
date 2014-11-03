// Http client hammer for stress-testing web services

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"errors"
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

var configFileName string
var helpMe bool

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
	Cred        string // Set after authentication
	Seed        string
	Hammers     int
	Seconds     int
	RequestPath string
	Log         bool // output requests and responses to stdout
}

// Module global
var config Config

type Result struct {
	Runs      int
	Succede   int
	Fail      int
	ByteCount int64
	Times     []int
}

// Set command line flags
func init() {
	// Seed rand with current nanoseconds
	rand.Seed(time.Now().UTC().UnixNano())
	flag.StringVar(&configFileName, "config", "config.json", "config file")
	flag.StringVar(&configFileName, "c", "config.json", "config file")
	flag.BoolVar(&helpMe, "help", false, "help")
	flag.BoolVar(&helpMe, "h", false, "help")
}

func main() {

	flag.Parse()

	if helpMe {
		flag.PrintDefaults()
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
		RequestPath: "request.log",
	}

	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Fatal("Config file not valid JSON: ", err)
	}

	fmt.Println("Config:")
	_ = printJson(content)

	// Open and parse the request log that will be fired at the target
	requestFile, err := os.Open(config.RequestPath)
	if err != nil {
		log.Fatalln("Could not read request file "+config.RequestPath, err)
	}
	defer requestFile.Close()

	requests, err := parseRequestLog(requestFile)
	if err != nil {
		log.Fatal("Error parsing "+config.RequestPath+": ", err)
	}

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
	res, err := client.Get(config.Host)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	// Make sure the host returns JSON and pretty-print it to the console
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	_ = printJson(body)

	// Autheticate and add the user credentail query string to config
	config.Cred, err = authenticate(client, &config)
	if err != nil {
		log.Fatal(err)
	}

	// Make sure we stay limited to one core so as not to
	// starve server when doing single-box testing
	runtime.GOMAXPROCS(1)

	// Start the collector service
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

// sum: read and sum the results from the channel
func sum(ch chan Result, expected int) {
	total := Result{}
	for i := 0; i < expected; i++ {
		result := <-ch
		total.Runs += result.Runs
		total.Succede += result.Succede
		total.Fail += result.Fail
		total.ByteCount += result.ByteCount
		total.Times = append(total.Times, result.Times...)
	}

	close(ch)
	failRate := float32(total.Fail) / float32(total.Succede+total.Fail)
	sort.Ints(total.Times)
	min := total.Times[0] / 100000
	max := total.Times[len(total.Times)-1] / 100000
	median := total.Times[len(total.Times)/2] / 100000
	var sum int64
	for i := range total.Times {
		sum += int64(total.Times[i])
	}
	mean := int(sum / int64(len(total.Times)) / 100000)

	fmt.Printf("\n\nResults: \n\n")
	fmt.Printf("Runs: %v\n", total.Runs)
	fmt.Printf("Requests: %v\n", total.Succede+total.Fail)
	fmt.Printf("Errors: %v\n", total.Fail)
	fmt.Printf("Fail Rate: %.2f\n", failRate)
	fmt.Printf("KBytes per second: %v\n", total.ByteCount/int64(config.Seconds)/1000)
	fmt.Printf("Requests per second: %v\n", (total.Succede+total.Fail)/config.Seconds)
	fmt.Printf("Min time: %v\nMax time: %v\nMean time: %v\nMedian time: %v\n\n", min, max, mean, median)

	os.Exit(0)
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

	newSeed := ""
	cReqs := len(requests)

	for i := 0; stop == false; i++ {

		if i >= cReqs { // start over
			result.Runs++
			i = 0
		}

		if i == 0 {
			newSeed = genNewSeed()
		}

		logReq := requests[i]

		delim := "?"
		if strings.Contains(logReq.Url, "?") {
			delim = "&"
		}

		method := strings.ToUpper(logReq.Method)
		url := config.Host + logReq.Url + delim + config.Cred

		// Replace the seed in urls with our newly generated seed
		url = strings.Replace(url, config.Seed, newSeed, -1)

		// Same for the body
		reqBody := bytes.Replace(logReq.Body, []byte(config.Seed), []byte(newSeed), -1)

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

		// Time the response
		after := time.Now().UnixNano()
		ns := int(after - before) // nanoseconds
		result.Times = append(result.Times, ns)

		body, _ := ioutil.ReadAll(res.Body)

		if 200 <= res.StatusCode && 400 > res.StatusCode {
			result.Succede++
		} else {
			result.Fail++
		}
		result.ByteCount += int64(len(body))

		if config.Log == true {
			fmt.Printf("\n%s %s\n", method, url)
			if len(reqBody) > 0 {
				fmt.Printf("%s\n", reqBody)
			}
			fmt.Printf("%v\n%s\n", res.StatusCode, body)
		}
	}

	ch <- result
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
			return requests, errors.New("Request log exceeded max of " + string(max))
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

// printJson: prettyPrint JSON to stdout
func printJson(data []byte) error {
	var indented bytes.Buffer
	err := json.Indent(&indented, data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", indented.String())
	return nil
}

// Authenticate the user specified in config.json
func authenticate(client *http.Client, config *Config) (string, error) {

	// Attempt to sign in
	url := config.Host + "/v1/auth/signin"

	reqBodyBytes, _ := json.Marshal(config.Signin)

	fmt.Println("Signin url:", url)
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
		return "", errors.New("Authentication failed with status " + string(res.StatusCode))
	}

	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		return "", err
	}

	fmt.Printf("%s\n%#v\n", "Signin response: ", body)
	credentials := "user=" + body.Session.UserId +
		"&session=" + body.Session.SessionId

	return credentials, nil
}

// Generate a new random numeric string
func genNewSeed() string {
	seedStr := strconv.FormatInt(rand.Int63(), 10)
	if len(seedStr) > 7 {
		// grab the last 8 digits
		seedStr = seedStr[len(seedStr)-8 : len(seedStr)-1]
	}
	return seedStr
}
