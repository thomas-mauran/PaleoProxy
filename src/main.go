package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-yaml/yaml"
)

func main(){
	domain := "localhost"
	fmt.Println("Proxy is running on port 8080")

	config, err := ReadConfig("./config.yaml")
    fmt.Printf("%#v %#v", config, err)

	services := config.Services

	handlers := map[string]http.HandlerFunc{}
	// We make a map for all our routes creating handlers
	for _, service := range services {
		if !service.Enabled {
			continue
		}
		// name := service.Name
		port := service.Port
		subdomain := service.Subdomain + "." + domain + ":" + fmt.Sprint(port)
		ip := service.Ip
		service := ip + ":" + fmt.Sprint(port)
		// description := service.Description

		handler := func (w http.ResponseWriter, req *http.Request){
			res , err := http.Get("http://" + service)
			fmt.Println("SENDING THE REQUEST TO" + service)
			if err != nil {
				io.WriteString(w, "[ERROR], an error occured when trying to reach GET" + service)
			}
			body, err := io.ReadAll(res.Body)
			if err != nil {
				io.WriteString(w, "[ERROR], failed to read response body from "+service)
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

type Conf struct {
	Services []Service `yaml:"services"`
}

type Service struct {
	Name string `yaml:"name"`
	Description string `yaml:"description"`
	Enabled bool `yaml:"enabled"`
	Subdomain string `yaml:"subdomain"`
	Ip string `yaml:"ip"`
	Port int64 `yaml:"port"`
}

func ReadConfig(filename string) (*Conf, error){
	buf, err := os.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	c := &Conf{}
	err = yaml.Unmarshal(buf, c)

	if err != nil {
		return nil, fmt.Errorf("in file %q: %w", filename, err)
	}
	return c, err
}