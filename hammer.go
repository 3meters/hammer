// Http client hammer for stress-testing web services

package main

import (
	"os"
	"flag"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

var configFileName string
var helpMe bool 

type Config struct {
	Host			string
	UserName	string
	Password	string
	UserId		string
	Session		string
	Hammers		int
	Seconds		int
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
		fmt.Print("Could not read config file " + configFileName, err)
		os.Exit(1)
	}

	var conf Config
	err = json.Unmarshal(content, &conf)
	if err != nil {
		fmt.Print("Error parsing config file: ", err)
		os.Exit(1)
	}

	fmt.Println(conf)
}
