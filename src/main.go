package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
)

func main(){
	domain := "localhost"

	fmt.Println("len", len(os.Args))

	if len(os.Args) < 2 {
		fmt.Println("Missing the config file path ! ./paleoproxy /path/to/the/config")
		return
	}

    configFilePath := os.Args[1]

	if _, err := os.Stat(configFilePath); errors.Is(err, os.ErrNotExist) {
		fmt.Println("Missing the config file path ! ./paleoproxy /path/to/the/config")
		return
	}

	fmt.Println("Paleo Proxy is up !")

	config, err := ReadConfig(configFilePath)
    fmt.Printf("%#v %#v", config, err)

	services := config.Services

	handlers := map[string]http.HandlerFunc{}
	// We make a map for all our routes creating handlers
	for _, service := range services {
		if !service.Enabled {
			continue
		}
		port := service.Port

		// Subdomain of the current service
		subdomain := service.Subdomain + "." + domain + ":" + fmt.Sprint(port)

		// Function to handle the passing of the request to the service
		handler := func (w http.ResponseWriter, req *http.Request){
			// Endpoints 
			endpoints := service.Endpoints

			// Pick a random endpoint
			randomEndpoint := endpoints[rand.Intn(len(endpoints))]

			// Ip and port of the service we chose
			ip := randomEndpoint.Ip

			// The url of the service to contact
			serviceUrl := "http://" + ip + ":" + fmt.Sprint(port)
			
			res , err := http.Get(serviceUrl)
			fmt.Println("SENDING THE REQUEST TO" + serviceUrl)
			if err != nil {
				io.WriteString(w, "[ERROR], an error occured when trying to reach GET" + serviceUrl)
			}

			body, err := io.ReadAll(res.Body)

			if err != nil {
				io.WriteString(w, "[ERROR], failed to read response body from "+ serviceUrl)
				return
			}
			w.Write(body)
		}
		handlers[subdomain] = handler
	}


	// Each time we get a request we check if the domain matches to route the traffic to it
	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		host := req.Host

		if handler, found := handlers[host]; found {
			handler(w, req)
			return
		}

		fmt.Println("New request received: \n", req) // Write the logs in a file
	})

	log.Fatal(http.ListenAndServe(":8080", mainHandler))
}