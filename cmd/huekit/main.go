package main

import (
	"fmt"
	"io"
	"os"

	badger "github.com/dgraph-io/badger/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/dj95/huekit/pkg/hue"
	"github.com/dj95/huekit/pkg/store"
)

func init() {
	// set the config name
	viper.SetConfigName("config")

	// add config paths
	viper.AddConfigPath(".")
	viper.AddConfigPath("/")

	// add command line flags
	initializeCommandFlags()

	// override the config file when the commandline flag is set
	if viper.IsSet("config") {
		viper.SetConfigFile(viper.GetString("config"))
	}

	// read the config file
	if err := viper.ReadInConfig(); err != nil {
		log.Warnf("Cannot read config file: %s", err.Error())
	}

	// set the default log level and mode
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{})

	// activate the debug mode
	if viper.GetString("log_level") == "debug" {
		log.SetLevel(log.DebugLevel)
	}

	// set the json formatter if configured
	if viper.GetString("log_format") == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	}

	// open the io writer for the log file
	file, err := os.OpenFile(
		"huekit.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)

	// create the log output
	logOutput := io.MultiWriter(os.Stdout, file)

	// if no error occurred...
	if err != nil {
		logOutput = os.Stdout
		log.Info("failed to log to file, using default stderr")
	}

	// set the stdout + file logger
	log.SetOutput(logOutput)
}

func main() {
	// open the database
	db, err := badger.Open(
		badger.
			DefaultOptions("./huekit_data").
			WithLogger(log.StandardLogger()),
	)

	// error handling
	if err != nil {
		log.Fatal(err)
	}

	// close the database on exit
	defer db.Close()

	// create a new storage with the database as backend
	store := store.NewBadger(db)

	// create a new bridge connection and authenticate,
	// if no authentication is saved in the storage
	bridge, err := hue.NewBridge(
		viper.GetString("bridge_address"),
		store,
	)

	// error handling
	if err != nil {
		log.Fatal(err.Error())
	}

	// fetch all lights
	lights, err := bridge.Lights()

	// error handling
	if err != nil {
		log.Fatal(err.Error())
	}

	// iterate through all the lights
	for _, light := range lights {
		fmt.Printf("%v\n", light)
	}
}

func initializeCommandFlags() {
	// create a new flag for docker health checks
	pflag.String("config", "", "choose the config file")

	// parse the pflags
	pflag.Parse()

	// bind the pflags
	viper.BindPFlags(pflag.CommandLine)
}
