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
	"math"
	"math/rand"
	"net/http"
	"os"
	"runtime"
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
}

type Result struct {
	Succede   int
	Fail      int
	ByteCount int64
}

// Set command line flags
func init() {
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

	config := Config{
		Hammers:     1,
		Seconds:     10,
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

	maxProcs := runtime.NumCPU() - 1
	fmt.Println("MaxProcs: ", maxProcs)
	runtime.GOMAXPROCS(maxProcs)

	// Start the hammers with a 0.1 second stagger
	sums := []<-chan Result{}
	for i := 0; i < config.Hammers; i++ {
		fmt.Println("Starting hammer ", i)
		go func() {
			sums[i] = run(client, &config, requests)
		}()
		// time.Sleep(100 * time.Millisecond)
	}
	for result := range sums {
		go sum(result)
	}
}

func sum(in ...<-chan Result) {
	total := Result{}
	for result := range in {
		fmt.Printf("Result: %#v\n", result)
		total.Succede += result.Succede
		total.Fail += result.Fail
		total.ByteCount += result.ByteCount
	}
	fmt.Printf("Grand Total: %#v\n", total)
}

// run: fire requests at the target with config credentials
func run(client *http.Client, config *Config, requests []Request) <-chan Result {

	out := make(chan Result, 1)
	result := Result{}

	stop := false

	// Start the clock
	go func() {
		fmt.Println("Start...")
		time.Sleep(time.Duration(config.Seconds) * time.Second)
		fmt.Println("Stop")
		stop = true
	}()

	newSeed := ""
	cReqs := len(requests)

	for i := 0; stop == false; i++ {

		if i >= cReqs { // start over
			i = 0
		}

		if i == 0 {
			newSeed = genNewSeed(config.Seed)
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
		res, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()

		body, _ := ioutil.ReadAll(res.Body)

		if 200 <= res.StatusCode && 400 > res.StatusCode {
			result.Succede++
		} else {
			result.Fail++
		}
		result.ByteCount += int64(len(body))
		// fmt.Printf("\n%s %s: %v\n%+s\n", method, logReq.Url, res.StatusCode, body)
	}

	out <- result
	close(out)
	return out
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

// Generate a new random numeric string the same length as the one passed in
func genNewSeed(seed string) string {
	seedRangeFloat := math.Pow10(len(seed))
	seedRangeInt := int64(seedRangeFloat)
	newSeedInt := rand.Int63n(seedRangeInt)
	newSeed := strconv.FormatInt(newSeedInt, 10)
	return newSeed
}
