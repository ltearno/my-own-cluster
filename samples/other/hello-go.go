package main

import (
	"fmt"
	"io/ioutil"
)

func main() {
	c := make(chan bool, 1)
	go func() {
		println("goodbye")
		c <- true
	}()

	<-c
	println("Hello world! doing some work")

	b, err := ioutil.ReadFile("https://home.lteconsulting.fr")
	if err != nil {
		fmt.Printf("erro ! %v\n", err)
	}

	fmt.Println(string(b))

	// does not work with tinygo, maybe works with go ?
	/*resp, err := http.Get("https://home.lteconsulting.fr")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()*/
}
