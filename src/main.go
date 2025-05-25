package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

func main(){
	isDynamic := false
	domain := "localhost"

	if len(os.Args) < 2 {
		fmt.Println("Missing the config file path ! ./paleoproxy /path/to/the/config")
		return
	}

	if len(os.Args) == 3 && os.Args[2] == "dynamic" {
		isDynamic = true
	}

    configFilePath := os.Args[1]

	if _, err := os.Stat(configFilePath); errors.Is(err, os.ErrNotExist) {
		fmt.Println("Missing the config file path ! ./paleoproxy /path/to/the/config")
		return
	}

	// Setup log file
	logfileName := "logs/" + fmt.Sprint(time.Now().Unix()) + ".log"
	logfile, err := os.OpenFile(logfileName, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logfile.Close()
	log.SetOutput(logfile)


	fmt.Println("Paleo Proxy is up !")

	if !isDynamic {

	config, err := ReadConfig(configFilePath)

	if err != nil {
		fmt.Println("An error occured when trying to read the config file: ", err)
		return
	}

	services := config.Services

	handlers := map[string]http.HandlerFunc{}
	// We make a map for all our routes creating handlers
	for _, s := range services {
		service := s
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
			log.Println("routing request to: " + serviceUrl)
			if err != nil {
				io.WriteString(w, "[ERROR], an error occured when trying to reach GET" + serviceUrl)
			}
			defer res.Body.Close()

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
	})

	log.Fatal(http.ListenAndServe(":8080", mainHandler))
	} else {
		fmt.Println("hey this is dynamic !")

		cli, err := client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			panic(err)
		}

		eventChannel, errChannel := cli.Events(context.Background(), events.ListOptions{})

		for {
			select {
			case event := <-eventChannel:
				// for key, value := range event.Actor.Attributes {
				// 	fmt.Printf("%s: %s\n", key, value)
				// }			
				_, gotPaleoLabel := event.Actor.Attributes["paleo-subdomain"]

				if event.Action == "start" && gotPaleoLabel {
					fmt.Println("start", event.Action)
				} else if event.Action == "kill" && gotPaleoLabel {
					fmt.Println("kill", event.Action)
				}

			case err := <-errChannel:
				if err != nil {
					panic(err)
				}
			}
		}
	}

}