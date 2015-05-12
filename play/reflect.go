package main

import (
	"fmt"
	"os"
	"reflect"
)

func swap(x, y interface{}) (interface{}, interface{}, error) {
	if reflect.TypeOf(x) != reflect.TypeOf(y) {
		return nil, nil, fmt.Errorf("%v and %v must be of the same type", x, y)
	}
	return y, x, nil
}

func main() {
	a, b, err := swap("hello", 1)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(a, b)
}
