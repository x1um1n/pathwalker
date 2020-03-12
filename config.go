package main

import (
	"log"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"

	"github.com/x1um1n/checkerr"
)

// K is the global koanf instance
var K = koanf.New(".")

// LoadKoanf populates k with default values from configs/default.yml, then
// overrides/appends to those values with environment variables prefixed with KOANF_
func LoadKoanf() {
	log.Println("Reading default config")
	err := K.Load(file.Provider("config/default.yml"), yaml.Parser())
	checkerr.CheckFatal(err, "Error reading default config file")

	log.Println("Checking for local secrets")
	err = K.Load(file.Provider("config/secrets.yml"), yaml.Parser())
	checkerr.Check(err, "Error reading local secrets file")

	log.Println("Checking environment for overrides")
	K.Load(env.Provider("KOANF_", ".", func(s string) string {
		return strings.ToLower(strings.TrimPrefix(s, "KOANF_"))
	}), nil)

	log.Printf("Checking for %s config", K.String("environment"))
	err = K.Load(file.Provider("config/"+K.String("environment")+".yml"), yaml.Parser())
	if !checkerr.Check(err, "Error reading "+K.String("environment")+" config file") {
		log.Printf("Using config for %s environment", K.String("environment"))
	}
}
