package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func readStateFromBinaryFile(stateFilename string) (t *cairo.Node, err error) {
	readFromBinaryFile(stateFilename, func(statesReader *bufio.Reader) interface{} {
		t, err = cairo.StateFromBinary(statesReader)
		t.GraphAndPicture(outputNameFromInputName(stateFilename))
		return t
	})
	return t, err
}

func readStateChangesFromBinaryFile(stateChangesFileName string) (d *cairo.Dict, err error) {
	readFromBinaryFile(stateChangesFileName, func(statesReader *bufio.Reader) interface{} {
		d, err = cairo.StateChangesFromBinary(statesReader)
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
		t.GraphAndPicture(outputNameFromInputName(stateFileName))
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

func readMappedStateFromCsvFile(stateFileName string) (t *cairo.Node, err error) {
	readFromCsvFile(stateFileName, func(state *bufio.Scanner) interface{} {
		t, err = cairo.MappedStateFromCsv(state)
		t.GraphAndPicture(outputNameFromInputName(stateFileName))
		return t
	})
	return t, err
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("main <stateFilename> <stateChangesFilename> <mappedStateFilename>")
		os.Exit(0)
	}
	stateFileName := os.Args[1]
	stateChangesFileName := os.Args[2]
	mappedStateChangesFileName := os.Args[3]
	stateFileExt := filepath.Ext(stateFileName)
	stateChangesFileExt := filepath.Ext(stateChangesFileName)
	var state *cairo.Node
	var stateChanges *cairo.Dict
	var err error
	if stateFileExt == ".csv" && stateChangesFileExt == ".csv" {
		state, err = readStateFromCsvFile(stateFileName)
		if err != nil {
			fmt.Printf("error reading CSV state file: %s", err)
			os.Exit(1)
		}
		stateChanges, err = readStateChangesFromCsvFile(stateChangesFileName)
		if err != nil {
			fmt.Printf("error reading CSV state change file: %s", err)
			os.Exit(1)
		}
		readMappedStateFromCsvFile(mappedStateChangesFileName)
	} else {
		state, err = readStateFromBinaryFile(stateFileName)
		if err != nil {
			fmt.Printf("error reading BIN state file: %s", err)
			os.Exit(1)
		}
		stateChanges, err = readStateChangesFromBinaryFile(stateChangesFileName)
		if err != nil {
			fmt.Printf("error reading BIN state change file: %s", err)
			os.Exit(1)
		}
	}
	stats := &cairo.Counters{}
	newState := cairo.Union(state, stateChanges, stats)
	newState.GraphAndPicture("new_" + outputNameFromInputName(stateFileName))

	fmt.Printf("Number of current state nodes: %d\n", state.Size())
	fmt.Printf("Number of state change nodes: %d\n", stateChanges.Size())
	fmt.Printf("Number of next state nodes: %d\n", newState.Size())
	fmt.Printf("Number of exposed nodes for next state: %d\n", stats.ExposedCount)
	fmt.Printf("Number of height taken for next state: %d\n", stats.HeightCount)
	fmt.Printf("Number of new hashes for next state: %d\n", newState.CountNewHashes())
}
