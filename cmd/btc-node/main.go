package main

import (
	"flag"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

func main() {
	file, err := os.OpenFile("/tmp/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	configPath := flag.String("config", "/home/emil/go/src/github.com/EmilGeorgiev/btc-node/cmd/btc-node/exampe_config.yaml", "path to the configuration file")
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
