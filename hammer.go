// Http client hammer for stress-testing web services

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

var configFileName string
var helpMe bool

type Request struct {
	Method string
	Path   string
	Body   string
}

type Config struct {
	Host      string
	Email     string
	Password  string
	InstallId string
	UserId    string
	Session   string
	Cred      string
	Hammers   int
	Seconds   int
	Requests  []Request
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
		fail("Could not read config file "+configFileName, err)
	}

	config := Config{
		Hammers: 1,
		Seconds: 10,
	}

	err = json.Unmarshal(content, &config)
	if err != nil {
		fail("Config file not valid JSON", err)
	}

	fmt.Println("Config: ", config)

	// Configure a transport that accepts self-singed certificates
	// Similar to curl --insecure
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	// Make sure we can reach the host
	if config.Host == "" {
		fmt.Println("Host is required")
		os.Exit(1)
	}
	res, err := client.Get(config.Host)
	if err != nil {
		fail("Could not connect to server", err)
	}
	defer res.Body.Close()

	// Make sure the host returns JSON and pretty-print it to the console
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read response", err)
	}
	_ = printJson(body)

	err = authenticate(client, &config)
	if err != nil {
		fail("Authentication failed", err)
	}
	run(client, &config)

}

// printJson: prettyPrint JSON to stdout
func printJson(data []byte) error {
	var indented bytes.Buffer
	err := json.Indent(&indented, data, "", "  ")
	if err != nil {
		fail("Invalid JSON", err)
	}
	fmt.Printf("%s", indented)
	fmt.Println()
	return nil
}

// Authenticate the user specified in config.json
func authenticate(client *http.Client, config *Config) error {
	if config.Cred != "" {
		return nil
	}
	if config.Session != "" && config.UserId != "" {
		config.Cred = "user=" + config.UserId + "&session=" + config.Session
		return nil
	}
	if config.Email == "" || config.Password == "" || config.InstallId == "" {
		return errors.New("No means to authenticate")
	}

	// Attempt to sign in
	url := config.Host + "/auth/signin?user[email]=" + config.Email +
		"&user[password]=" + config.Password + "&installId=" + config.InstallId
	fmt.Println("url", url)
	res, err := client.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return errors.New("Failed")
	}
	return nil
}

func run(client *http.Client, config *Config) {
	fmt.Println("Requests:")
	for _, req := range config.Requests {
		fmt.Println(req)
	}
}

func fail(msg string, err error) {
	if err != nil {
		msg += ": " + err.Error()
	}
	fmt.Println("Error:", msg)
	os.Exit(1)
}
