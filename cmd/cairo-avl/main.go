package main

import (
	"bufio"
	"flag"
	"os"
	"path/filepath"
	"strings"

	cairo "github.com/canepat/bst/cairo-avl"
	log "github.com/sirupsen/logrus"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func outputNameFromInputName(inputFileName string) string {
	return strings.ReplaceAll(inputFileName, ".", "_")
}

func readFromBinaryFile(binaryFilename string, readFunction func(*bufio.Reader) interface{}) interface{} {
	binaryFile, err := os.Open(binaryFilename)
	check(err)
	defer binaryFile.Close()
	stateReader := bufio.NewReader(binaryFile)
	return readFunction(stateReader)
}

func readStateFromBinaryFile(stateFilename string, keySize int, nested bool) (t *cairo.Node, err error) {
	readFromBinaryFile(stateFilename, func(statesReader *bufio.Reader) interface{} {
		t, err = cairo.StateFromBinary(statesReader, keySize, nested)
		t.GraphAndPicture(outputNameFromInputName(stateFilename), /*debug=*/false)
		return t
	})
	return t, err
}

func readStateChangesFromBinaryFile(stateChangesFileName string, keySize int, nested bool) (d *cairo.Dict, err error) {
	readFromBinaryFile(stateChangesFileName, func(statesReader *bufio.Reader) interface{} {
		d, err = cairo.StateChangesFromBinary(statesReader, keySize, nested)
		d.GraphAndPicture(outputNameFromInputName(stateChangesFileName))
		return d
	})
	return d, err
}

func readFromCsvFile(csvFileName string, scanFromCsv func(*bufio.Scanner) interface{}) interface{} {
	csvFile, err := os.Open(csvFileName)
	check(err)
	defer csvFile.Close()
	stateScanner := bufio.NewScanner(csvFile)
	return scanFromCsv(stateScanner)
}

func readStateFromCsvFile(stateFileName string) (t *cairo.Node, err error) {
	readFromCsvFile(stateFileName, func(state *bufio.Scanner) interface{} {
		t, err = cairo.StateFromCsv(state)
		t.GraphAndPicture(outputNameFromInputName(stateFileName), /*debug=*/false)
		return t
	})
	return t, err
}

func readStateChangesFromCsvFile(stateChangesFileName string) (d *cairo.Dict, err error) {
	readFromCsvFile(stateChangesFileName, func(stateChanges *bufio.Scanner) interface{} {
		d, err = cairo.StateChangesFromCsv(stateChanges)
		d.GraphAndPicture(outputNameFromInputName(stateChangesFileName))
		return d
	})
	return d, err
}

var options Options

func init() {
	const hasCustomFormatter = false
	if hasCustomFormatter {
		customFormatter := new(log.TextFormatter)
		customFormatter.TimestampFormat = "2006-01-02 15:04:05"
		customFormatter.FullTimestamp = true
		log.SetFormatter(customFormatter)
	}

	options = Options{}
	flag.StringVar(&options.stateFileName, "stateFileName", "", "the state file name")
	flag.StringVar(&options.stateChangesFileName, "stateChangesFileName", "", "the state-change file name")
	flag.IntVar(&options.keySize, "keySize", 4, "the key size in bytes")
	flag.BoolVar(&options.nested, "nested", true, "flag indicating if tree should be nested or not")
	flag.StringVar(&options.logLevel, "logLevel", "INFO", "the logging level")
	flag.BoolVar(&options.graph, "graph", false, "flag indicating if tree graph should be saved or not")
}

type Options struct {
	stateFileName		string
	stateChangesFileName	string
	keySize			int
	nested			bool
	logLevel		string
	graph			bool
}

func main() {
	flag.Parse()

	level, err := log.ParseLevel(options.logLevel)
	if err != nil {
		log.Fatalln("logLevel argument error: ", err)
		os.Exit(0)
	}
	log.SetLevel(level)

	log.Printf("Name of the state file: %s\n", options.stateFileName)
	log.Printf("Name of the state changes file: %s\n", options.stateChangesFileName)
	log.Printf("Size of the key in bytes: %d\n", options.keySize)
	log.Printf("Trees are nested: %t\n", options.nested)
	log.Printf("Log level: %s\n", options.logLevel)

	stateFileExt := filepath.Ext(options.stateFileName)
	stateChangesFileExt := filepath.Ext(options.stateChangesFileName)
	var state *cairo.Node
	var stateChanges *cairo.Dict
	if stateFileExt == ".csv" && stateChangesFileExt == ".csv" {
		state, err = readStateFromCsvFile(options.stateFileName)
		if err != nil {
			log.Fatalln("error reading CSV state file: %s", err)
			os.Exit(1)
		}
		stateChanges, err = readStateChangesFromCsvFile(options.stateChangesFileName)
		if err != nil {
			log.Fatalln("error reading CSV state change file: %s", err)
			os.Exit(1)
		}
	} else {
		state, err = readStateFromBinaryFile(options.stateFileName, options.keySize, options.nested)
		if err != nil {
			log.Fatalln("error reading BIN state file: %s", err)
			os.Exit(1)
		}
		stateChanges, err = readStateChangesFromBinaryFile(options.stateChangesFileName, options.keySize, options.nested)
		if err != nil {
			log.Fatalln("error reading BIN state change file: %s", err)
			os.Exit(1)
		}
	}

	if options.graph {
		state.GraphAndPicture("state_" + outputNameFromInputName(options.stateFileName), /*debug=*/false)
	}

	unionStats := &cairo.Counters{}
	newState := cairo.Union(state, stateChanges, unionStats)
	if options.graph {
		newState.GraphAndPicture("stateAfterUnion_" + outputNameFromInputName(options.stateFileName), /*debug=*/false)
	}

	log.Printf("UNION: Number of nodes in the current state tree: %d\n", state.Size())
	log.Printf("UNION: Number of nodes in the state update tree: %d\n", stateChanges.Size())
	log.Printf("UNION: Number of nodes in the next state tree: %d\n", newState.Size())
	log.Printf("UNION: Number of re-hashes for the next state: %d\n", newState.CountNewHashes())
	log.Printf("UNION: Number of nodes exposed: %d\n", unionStats.ExposedCount)
	log.Printf("UNION: Number of nodes with height taken: %d\n", unionStats.HeightCount)

	diffStats := &cairo.Counters{}
	newState = cairo.Difference(state, stateChanges, diffStats)
	if options.graph {
		newState.GraphAndPicture("stateAfterDiff_" + outputNameFromInputName(options.stateFileName), /*debug=*/false)
	}

	log.Printf("DIFFERENCE: Number of nodes in the current state tree: %d\n", state.Size())
	log.Printf("DIFFERENCE: Number of nodes in the state update tree: %d\n", stateChanges.Size())
	log.Printf("DIFFERENCE: Number of nodes in the next state tree: %d\n", newState.Size())
	log.Printf("DIFFERENCE: Number of re-hashes for the next state: %d\n", newState.CountNewHashes())
	log.Printf("DIFFERENCE: Number of nodes exposed: %d\n", diffStats.ExposedCount)
	log.Printf("DIFFERENCE: Number of nodes with height taken: %d\n", diffStats.HeightCount)
}
