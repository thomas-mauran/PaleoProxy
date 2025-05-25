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
	"sync"
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

	if !isDynamic {
		// Not dynamic, using the config file
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

			handlers[subdomain] = CreateHandler(service)
		}

		fmt.Println("Paleo Proxy (config mode) is up !")

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
		// Handling the dynamic mode !
		var (
			handlers = make(map[string]http.HandlerFunc)
			handlersMu sync.RWMutex
		)
		// Dynamic config, listening to docker events
		cli, err := client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			panic(err)
		}

		eventChannel, errChannel := cli.Events(context.Background(), events.ListOptions{})

		// We start the infinite event listener loop with a gorouting 
		go DynamicListen(cli, eventChannel, errChannel, handlers, &handlersMu, domain)

		fmt.Println("Paleo Proxy (dynamic mode) is up !")

		mainHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			host := req.Host

			handlersMu.RLock()
			handler, found := handlers[host]
			handlersMu.RUnlock()

			if found {
				handler(w, req)
			} else {
				http.NotFound(w, req)
			}
		})
		log.Fatal(http.ListenAndServe(":8080", mainHandler))
	}
}

// Method to loop infinitely on the docker events and add handlers to the map for events with the paleo-subdomain label
func DynamicListen(cli *client.Client, eventChannel <- chan events.Message, errChannel <-chan error, handlers map[string]http.HandlerFunc, handlersMu *sync.RWMutex, domain string) {
	for {
		select {
		case event := <-eventChannel:
			subdomain, gotPaleoLabel := event.Actor.Attributes["paleo-subdomain"]
			if !gotPaleoLabel {
				continue
			}
			fmt.Println("new event")
			if event.Action == "start" && gotPaleoLabel {
				containerID := event.ID
				containerJSON, err := cli.ContainerInspect(context.Background(), containerID)
				if err != nil {
					log.Printf("Failed to inspect container %s: %v", containerID, err)
					continue
				}

				// Get the first network's IP address ! This might be a bad idea in the end idk how it works with multiple networks
				// Right now we don't handle multiple services, we shall refactor 
				var ipAddress string
				for _, netSettings := range containerJSON.NetworkSettings.Networks {
					ipAddress = netSettings.IPAddress
					break
				}
				// Build service from container info (assuming paleo-subdomain, port labels exist)
				service := Service{
					Subdomain: subdomain,
					Port: 8080, // For now we hardocode this to the 8080 port 
					Endpoints: []Endpoint{
						{Ip: ipAddress},
					},
					Enabled: true,
				}

				handler := CreateHandler(service)

				handlersMu.Lock()
				handlers[subdomain + "." + domain + ":8080"] = handler
				handlersMu.Unlock()

				log.Printf("Added handler for %s\n", subdomain)
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

// This method creates the http handle methods
func CreateHandler(service Service) http.HandlerFunc {
	// Function to handle the passing of the request to the service
	return func (w http.ResponseWriter, req *http.Request){
		// Endpoints 
		endpoints := service.Endpoints

		// Pick a random endpoint
		randomEndpoint := endpoints[rand.Intn(len(endpoints))]

		// Ip and port of the service we chose
		ip := randomEndpoint.Ip

		fmt.Println("ip", ip)

		// The url of the service to contact
		serviceUrl := "http://" + ip + ":" + fmt.Sprint(service.Port)
		
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
}