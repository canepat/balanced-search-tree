package main

import (
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func init() {
	const hasCustomFormatter = false
	if hasCustomFormatter {
		customFormatter := new(log.TextFormatter)
		customFormatter.TimestampFormat = "2006-01-02 15:04:05"
		customFormatter.FullTimestamp = true
		log.SetFormatter(customFormatter)
	}
}

type Options struct {
	stateFileName		string
	stateChangesFileName	string
	keySize			int
	nested			bool
	logLevel		log.Level
}

func parseOptions(args []string) *Options {
	if len(args) < 3 {
		fmt.Println("usage: main <stateFilename> <stateChangesFilename> [<keySize>] [<nested>] [<logLevel>]")
		os.Exit(0)
	}
	stateFileName := os.Args[1]
	stateChangesFileName := os.Args[2]
	keySize := parseKeySize()
	nested := parseNested()
	logLevel := parseLogLevel()
	options := &Options{
		stateFileName: stateFileName,
		stateChangesFileName: stateChangesFileName,
		keySize: keySize,
		nested: nested,
		logLevel: logLevel,
	}
	return options
}

func parseKeySize() int {
	keySize := 4
	if len(os.Args) > 3 {
		var err error
		keySize, err = strconv.Atoi(os.Args[3])
		if err != nil {
			log.Fatalln("keySize argument error: ", err)
			os.Exit(0)
		}
	}
	return keySize
}

func parseNested() bool {
	nested := true
	if len(os.Args) > 4 {
		var err error
		nested, err = strconv.ParseBool(os.Args[4])
		if err != nil {
			log.Fatalln("nested argument error: ", err)
			os.Exit(0)
		}
	}
	return nested
}

func parseLogLevel() log.Level {
	level := log.InfoLevel
	if len(os.Args) > 5 {
		var err error
		level, err = log.ParseLevel(os.Args[5])
		if err != nil {
			log.Fatalln("logLevel argument error: ", err)
			os.Exit(0)
		}
	}
	return level
}

func main() {
	options := parseOptions(os.Args)

	log.SetLevel(options.logLevel)

	log.Printf("Name of the state file: %s\n", options.stateFileName)
	log.Printf("Name of the state changes file: %s\n", options.stateChangesFileName)
	log.Printf("Size of the key in bytes: %d\n", options.keySize)
	log.Printf("Trees are nested: %t\n", options.nested)
	log.Printf("Log level: %s\n", options.logLevel)
}
