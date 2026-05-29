package main

import (
	"encoding/json"
	"fmt"
	"log"
	"openaccess-sync/pkg/client"
	"openaccess-sync/pkg/config"
	"os"

	"github.com/magiconair/properties"
	"github.com/spf13/pflag"
)

func main() {
	fs := pflag.NewFlagSet("querytool", pflag.ContinueOnError)

	var configFile string
	var typeName, filter string
	var endpoint, application, user, password, directory string
	var file string
	var pageSize int
	var insecure bool

	fs.StringVarP(&configFile, "config", "c", "", "Configuration file")
	fs.StringVarP(&typeName, "type", "t", "", "Object type name (e.g. Lnl_Badge, Lnl_Cardholder)")
	fs.StringVarP(&filter, "filter", "F", "", "Filter expression")
	fs.StringVarP(&endpoint, "endpoint", "e", "", "API endpoint")
	fs.StringVarP(&application, "application", "a", "", "Application ID")
	fs.StringVarP(&user, "user", "u", "", "Username")
	fs.StringVarP(&password, "password", "p", "", "Password")
	fs.StringVarP(&directory, "directory", "d", "", "Directory ID")
	fs.BoolVarP(&insecure, "insecure", "i", false, "Disable SSL certificate validation")
	fs.IntVarP(&pageSize, "pagesize", "P", config.DefaultPageSize, "Page size (1-100)")
	fs.StringVarP(&file, "file", "f", "", "Output file path (default: stdout)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: querytool -t TYPE [flags]\n\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(os.Args[1:]); err != nil {
		if err == pflag.ErrHelp {
			os.Exit(0)
		}
		log.Fatalf("Error parsing arguments: %v", err)
	}

	cfg := config.AppConfig{PageSize: config.DefaultPageSize}

	if fs.Changed("config") {
		props, err := properties.LoadFile(configFile, properties.UTF8)
		if err != nil {
			log.Fatalf("Error loading configuration file: %v", err)
		}
		cfg.Endpoint = props.GetString("endpoint", "")
		cfg.Application = props.GetString("application", "")
		cfg.User = props.GetString("user", "")
		cfg.Password = props.GetString("password", "")
		cfg.Directory = props.GetString("directory", "")
		cfg.Insecure = props.GetBool("insecure", false)
		if ps := props.GetInt("pagesize", 0); ps > 0 && ps <= config.MaxPageSize {
			cfg.PageSize = ps
		}
	}

	if fs.Changed("endpoint") {
		cfg.Endpoint = endpoint
	}
	if fs.Changed("application") {
		cfg.Application = application
	}
	if fs.Changed("user") {
		cfg.User = user
	}
	if fs.Changed("password") {
		cfg.Password = password
	}
	if fs.Changed("directory") {
		cfg.Directory = directory
	}
	if fs.Changed("insecure") {
		cfg.Insecure = insecure
	}
	if fs.Changed("pagesize") {
		cfg.PageSize = pageSize
	}
	if fs.Changed("file") {
		cfg.File = file
	}

	if typeName == "" {
		fmt.Fprintln(os.Stderr, "Error: --type (-t) is required")
		fs.Usage()
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	cl, err := client.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}
	defer func() {
		if err := cl.Close(); err != nil {
			log.Printf("Failed to close client session: %v", err)
		}
	}()

	items, err := cl.GetInstances(typeName, filter)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	out, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal results: %v", err)
	}

	if cfg.File == "" {
		fmt.Println(string(out))
		return
	}

	if err := os.WriteFile(cfg.File, append(out, '\n'), 0644); err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	log.Printf("Wrote %d items to %s", len(items), cfg.File)
}
