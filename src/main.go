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
	fmt.Println("Proxy is running on port 8080")

	config, err := ReadConfig("./config.yaml")
    fmt.Printf("%#v %#v", config, err)

	services := config.Services
	for _, service := range services {
		fmt.Printf("sevices:::::::::::::::::: %#v", service.Description)

		handler := func (w http.ResponseWriter, req *http.Request){
			io.WriteString(w, service.Description + ": " + service.Description)
		}

		http.HandleFunc("/" + service.Name, handler)
	}


	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, world!\n")
		fmt.Println("New request", req)
	}

	http.HandleFunc("/hello", helloHandler)



	log.Fatal(http.ListenAndServe(":8080", nil))
}

type Conf struct {
	Services []Service `yaml:"services"`
}

type Service struct {
	Name string `yaml:"name"`
	Description string `yaml:"description"`
	Enabled bool `yaml:"enabled"`
	Url string `yaml:"url"`
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