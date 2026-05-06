package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/magiconair/properties"
	"github.com/spf13/pflag"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// AppConfig holds all validated configuration for the application.
type AppConfig struct {
	Endpoint    string
	Application string
	User        string
	Password    string
	Directory   string
	Insecure    bool
	PageSize    int
	File        string
	Export      bool
	Sync        bool
	FullExport  bool
	Cleanup     bool
	DiffFile    string
	Verbose     bool
}

// Validate returns an error if any required field is missing or invalid.
func (c AppConfig) Validate() error {
	if strings.TrimSpace(c.Endpoint) == "" {
		return errors.New("endpoint is required")
	}
	if !isValidURL(c.Endpoint) {
		return errors.New("endpoint must be a valid URL")
	}
	if strings.TrimSpace(c.Application) == "" {
		return errors.New("application is required")
	}
	if strings.TrimSpace(c.User) == "" {
		return errors.New("user is required")
	}
	if strings.TrimSpace(c.Password) == "" {
		return errors.New("password is required")
	}
	if strings.TrimSpace(c.Directory) == "" {
		return errors.New("directory is required")
	}
	if c.PageSize <= 0 || c.PageSize > MaxPageSize {
		return fmt.Errorf("invalid pageSize - must be greater than zero and less than or equal to %d", MaxPageSize)
	}
	return nil
}

// parseConfig parses CLI arguments into an AppConfig.
// Config file values (if -c is given) are loaded first; CLI flags override them.
func parseConfig(args []string) (AppConfig, error) {
	fs := pflag.NewFlagSet("openaccess-sync", pflag.ContinueOnError)

	var configFile string
	var endpoint, application, user, password, directory string
	var insecureStr string
	var pageSize int
	var file, diffFile string
	var export, sync, fullExport, cleanup, verbose bool

	fs.StringVarP(&configFile, "config", "c", "", "Configuration file")
	fs.StringVarP(&endpoint, "endpoint", "e", "", "API endpoint")
	fs.StringVarP(&application, "application", "a", "", "Application ID")
	fs.StringVarP(&user, "user", "u", "", "Username")
	fs.StringVarP(&password, "password", "p", "", "Password")
	fs.StringVarP(&directory, "directory", "d", "", "Directory ID")
	fs.StringVarP(&insecureStr, "insecure", "i", "false", "Disable SSL certificate validation (true/false)")
	fs.IntVarP(&pageSize, "pagesize", "P", DefaultPageSize, "Page size (1-100)")
	fs.StringVarP(&file, "file", "f", "", "File path for the active mode")
	fs.BoolVarP(&export, "export", "x", false, "Export records to CSV (use with --file)")
	fs.BoolVarP(&sync, "sync", "s", false, "Sync/compare CSV against API (use with --file)")
	fs.BoolVarP(&fullExport, "fullexport", "X", false, "Full XLSX export (use with --file)")
	fs.BoolVarP(&cleanup, "cleanup", "k", false, "Cleanup")
	fs.BoolVarP(&verbose, "verbose", "v", false, "Verbose output (debug)")
	fs.StringVarP(&diffFile, "diff", "D", "", "File to write ContentEquals diff output for debugging")

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage of openaccess-sync:\n")
		fs.PrintDefaults()
	}

	if len(args) == 0 {
		fs.Usage()
		return AppConfig{}, pflag.ErrHelp
	}

	if err := fs.Parse(args); err != nil {
		return AppConfig{}, err
	}

	cfg := AppConfig{PageSize: DefaultPageSize}

	// Load properties file first if -c was provided
	if fs.Changed("config") {
		props, err := properties.LoadFile(configFile, properties.UTF8)
		if err != nil {
			return AppConfig{}, fmt.Errorf("error loading configuration file: %w", err)
		}
		cfg.Endpoint = props.GetString("endpoint", "")
		cfg.Application = props.GetString("application", "")
		cfg.User = props.GetString("user", "")
		cfg.Password = props.GetString("password", "")
		cfg.Directory = props.GetString("directory", "")
		cfg.Insecure = props.GetBool("insecure", false)
		if ps := props.GetInt("pagesize", 0); ps > 0 && ps <= MaxPageSize {
			cfg.PageSize = ps
		}
	}

	// CLI flags override file values (only when explicitly provided)
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
		cfg.Insecure = strings.EqualFold(insecureStr, "true")
	}
	if fs.Changed("pagesize") {
		if pageSize > 0 && pageSize <= MaxPageSize {
			cfg.PageSize = pageSize
		} else {
			cfg.PageSize = DefaultPageSize
		}
	}
	if fs.Changed("file") {
		cfg.File = file
	}
	if fs.Changed("export") {
		cfg.Export = export
	}
	if fs.Changed("sync") {
		cfg.Sync = sync
	}
	if fs.Changed("fullexport") {
		cfg.FullExport = fullExport
	}
	if fs.Changed("cleanup") {
		cfg.Cleanup = cleanup
	}
	if fs.Changed("diff") {
		cfg.DiffFile = diffFile
	}
	if fs.Changed("verbose") {
		cfg.Verbose = verbose
	}

	// Require exactly one mode flag
	modeCount := 0
	for _, flag := range []string{"export", "sync", "fullexport", "cleanup"} {
		if fs.Changed(flag) {
			modeCount++
		}
	}
	if modeCount == 0 {
		return AppConfig{}, errors.New("one of --sync (-s), --export (-x), --fullexport (-X), or --cleanup (-k) is required")
	}
	if (cfg.Export || cfg.Sync || cfg.FullExport) && cfg.File == "" {
		return AppConfig{}, errors.New("--file (-f) is required with --export, --sync, and --fullexport")
	}

	return cfg, nil
}
