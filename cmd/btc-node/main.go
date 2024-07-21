package main

import (
	"flag"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

func main() {
	configPath := flag.String("config", "example_config.yaml", "path to the configuration file")
	flag.Parse()

	data, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Error reading YAML file: %s", err)
	}

	// Unmarshal the YAML file into the Config struct
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("Error parsing YAML file: %s", err)
	}

	if err = cfg.Validate(); err != nil {
		log.Fatalf("invalid config file: %s", err)
	}

	Run(cfg)
}
