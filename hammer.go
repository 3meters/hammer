// Http client hammer for stress-testing web services

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

var configFileName string
var helpMe bool

type Config struct {
	Host     string
	UserName string
	Password string
	UserId   string
	Session  string
	Cred     string
	Hammers  int
	Seconds  int
	paths    []string
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
		fmt.Print("Could not read config file "+configFileName, err)
		os.Exit(1)
	}

	var conf Config
	err = json.Unmarshal(content, &conf)
	if err != nil {
		fmt.Println("Error parsing config file:", err)
		os.Exit(1)
	}

	fmt.Println("Config: ", conf)

	// Configure a transport that accepts self-singed certs
	// Same as curl --insecure
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},  // accept self-signed certificates
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	// Run get on the host url to 
	res, err := client.Get(conf.Host)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	var indented bytes.Buffer
	json.Indent(&indented, body, "", "  ")
	fmt.Printf("%s", indented)
	fmt.Println()

}
