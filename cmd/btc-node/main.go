package main

import (
	"flag"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

func main() {
	configPath := flag.String("config", "./example_config.yaml", "path to the configuration file")
	logsPath := flag.String("logs_path", "/tmp/app.log", "path to the configuration file")
	flag.Parse()

	file, err := os.OpenFile(*logsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

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
