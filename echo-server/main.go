package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

var (
	DefaultEndpoint = []string{"/"}
)

type Config struct {
	Endpoints []string `yaml:"endpoints"`
}

func echo(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": c.Request.URL.Path,
	})
}

func main() {
	var configPath string
	var Port string

	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.StringVar(&Port, "port", "8080", "Port to listen on")
	flag.Parse()

	config := Config{}
	if configPath != "" {
		content, err := os.ReadFile(configPath)
		if err != nil {
			panic(err)
		}

		err = yaml.Unmarshal(content, &config)
		if err != nil {
			panic(err)
		}
	}

	if len(config.Endpoints) == 0 {
		config.Endpoints = DefaultEndpoint
	}

	router := gin.Default()
	for _, endpoint := range config.Endpoints {
		router.GET(endpoint, echo)
	}

	router.Run(fmt.Sprintf(":%s", Port))
}
