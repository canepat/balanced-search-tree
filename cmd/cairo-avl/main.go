package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	cairo "github.com/canepat/bst/cairo-avl"
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

func init() {
	const hasCustomFormatter = false
	if hasCustomFormatter {
		customFormatter := new(log.TextFormatter)
		customFormatter.TimestampFormat = "2006-01-02 15:04:05"
		customFormatter.FullTimestamp = true
		log.SetFormatter(customFormatter)
	}
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
	if len(os.Args) < 3 {
		fmt.Println("usage: main <stateFilename> <stateChangesFilename> [<keySize>] [<nested>] [<logLevel>]")
		os.Exit(0)
	}
	var err error
	stateFileName := os.Args[1]
	stateChangesFileName := os.Args[2]
	keySize := parseKeySize()
	nested := parseNested()
	logLevel := parseLogLevel()

	log.SetLevel(logLevel)

	log.Printf("Name of the state file: %s\n", stateFileName)
	log.Printf("Name of the state changes file: %s\n", stateChangesFileName)
	log.Printf("Size of the key in bytes: %d\n", keySize)
	log.Printf("Trees are nested: %t\n", nested)
	log.Printf("Log level: %s\n", logLevel)

	stateFileExt := filepath.Ext(stateFileName)
	stateChangesFileExt := filepath.Ext(stateChangesFileName)
	var state *cairo.Node
	var stateChanges *cairo.Dict
	if stateFileExt == ".csv" && stateChangesFileExt == ".csv" {
		state, err = readStateFromCsvFile(stateFileName)
		if err != nil {
			log.Fatalln("error reading CSV state file: %s", err)
			os.Exit(1)
		}
		stateChanges, err = readStateChangesFromCsvFile(stateChangesFileName)
		if err != nil {
			log.Fatalln("error reading CSV state change file: %s", err)
			os.Exit(1)
		}
	} else {
		state, err = readStateFromBinaryFile(stateFileName, keySize, nested)
		if err != nil {
			log.Fatalln("error reading BIN state file: %s", err)
			os.Exit(1)
		}
		stateChanges, err = readStateChangesFromBinaryFile(stateChangesFileName, keySize, nested)
		if err != nil {
			log.Fatalln("error reading BIN state change file: %s", err)
			os.Exit(1)
		}
	}
	unionStats := &cairo.Counters{}
	newState := cairo.Union(state, stateChanges, unionStats)
	newState.GraphAndPicture("union_new_" + outputNameFromInputName(stateFileName), /*debug=*/false)

	log.Printf("UNION: Number of nodes in the current state tree: %d\n", state.Size())
	log.Printf("UNION: Number of nodes in the state update tree: %d\n", stateChanges.Size())
	log.Printf("UNION: Number of nodes in the next state tree: %d\n", newState.Size())
	log.Printf("UNION: Number of re-hashes for the next state: %d\n", newState.CountNewHashes())
	log.Printf("UNION: Number of nodes exposed: %d\n", unionStats.ExposedCount)
	log.Printf("UNION: Number of nodes with height taken: %d\n", unionStats.HeightCount)

	diffStats := &cairo.Counters{}
	newState = cairo.Difference(state, stateChanges, diffStats)
	newState.GraphAndPicture("diff_new_" + outputNameFromInputName(stateFileName), /*debug=*/false)

	log.Printf("DIFFERENCE: Number of nodes in the current state tree: %d\n", state.Size())
	log.Printf("DIFFERENCE: Number of nodes in the state update tree: %d\n", stateChanges.Size())
	log.Printf("DIFFERENCE: Number of nodes in the next state tree: %d\n", newState.Size())
	log.Printf("DIFFERENCE: Number of re-hashes for the next state: %d\n", newState.CountNewHashes())
	log.Printf("DIFFERENCE: Number of nodes exposed: %d\n", diffStats.ExposedCount)
	log.Printf("DIFFERENCE: Number of nodes with height taken: %d\n", diffStats.HeightCount)
}
