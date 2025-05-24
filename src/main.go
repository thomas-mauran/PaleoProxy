package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func main(){
	fmt.Println("Hello World")

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, world!\n")
		fmt.Println("New request", req)
	}

	http.HandleFunc("/hello", helloHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}