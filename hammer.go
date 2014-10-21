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
	"net/http"
	"os"
	"strings"
)

const contentJson = "application/json"

var configFileName string
var helpMe bool

type Request struct {
	Method string          `json:"method"`
	Url    string          `json:"url"`
	Body   json.RawMessage `json:"body,omitempty"`
}

type Requests []Request

type Config struct {
	Host   string
	Signin struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		InstallId string `json:"installId"`
	}
	Seed        string
	Hammers     int
	Seconds     int
	RequestPath string
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

	cred, authErr := authenticate(client, &config)
	if authErr != nil {
		log.Fatal(authErr)
	}
	run(client, &config, requests, cred)

}

// parseRequestLog: parse our modified csv log format
func parseRequestLog(file *os.File) (requests []Request, err error) {

	const max = 10000
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

		// fmt.Println(record[0])
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
	fmt.Printf("%v", indented.String())
	return nil
}

// Authenticate the user specified in config.json
func authenticate(client *http.Client, config *Config) (string, error) {

	// Attempt to sign in
	url := config.Host + "/v1/auth/signin"

	reqBodyBytes, _ := json.Marshal(config.Signin)

	fmt.Println("signin url:", url)
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

// run: fire requests at the target with config credentials
func run(client *http.Client, config *Config, requests Requests, cred string) {
	for _, logReq := range requests {
		delim := "?"
		if strings.Contains(logReq.Url, "?") {
			delim = "&"
		}
		method := strings.ToUpper(logReq.Method)
		url := config.Host + logReq.Url + delim + cred
		req, reqErr := http.NewRequest(method, url, bytes.NewReader(logReq.Body))
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
		fmt.Printf("\n%s %s: %v\n%+s\n", method, logReq.Url, res.StatusCode, body)
	}
}
