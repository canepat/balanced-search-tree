package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

func readStateFromBinaryFile(stateFilename string, keySize int, nested bool) (t *cairo.Node, err error) {
	readFromBinaryFile(stateFilename, func(statesReader *bufio.Reader) interface{} {
		t, err = cairo.StateFromBinary(statesReader, keySize, nested)
		t.GraphAndPicture(outputNameFromInputName(stateFilename))
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

func main() {
	if len(os.Args) < 3 {
		fmt.Println("main <stateFilename> <stateChangesFilename> [<keySize>] [<nested>]")
		os.Exit(0)
	}
	var err error
	stateFileName := os.Args[1]
	stateChangesFileName := os.Args[2]
	keySize := 4
	if len(os.Args) > 3 {
		keySize, err = strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Println("keySize argument error: ", err)
			os.Exit(0)
		}
	}
	nested := true
	if len(os.Args) > 4 {
		nested, err = strconv.ParseBool(os.Args[4])
		if err 	 	 != nil {
			fmt.Println("nested argument error: ", err)
			os.Exit(0)
		}
	}
	stateFileExt := filepath.Ext(stateFileName)
	stateChangesFileExt := filepath.Ext(stateChangesFileName)
	var state *cairo.Node
	var stateChanges *cairo.Dict
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
	} else {
		state, err = readStateFromBinaryFile(stateFileName, keySize, nested)
		if err != nil {
			fmt.Printf("error reading BIN state file: %s", err)
			os.Exit(1)
		}
		stateChanges, err = readStateChangesFromBinaryFile(stateChangesFileName, keySize, nested)
		if err != nil {
			fmt.Printf("error reading BIN state change file: %s", err)
			os.Exit(1)
		}
	}
	unionStats := &cairo.Counters{}
	newState := cairo.Union(state, stateChanges, unionStats)
	newState.GraphAndPicture("union_new_" + outputNameFromInputName(stateFileName))

	fmt.Printf("#UNION:\n")
	fmt.Printf("Number of nodes in the current state tree: %d\n", state.Size())
	fmt.Printf("Number of nodes in the state update tree: %d\n", stateChanges.Size())
	fmt.Printf("Number of nodes in the next state tree: %d\n", newState.Size())
	fmt.Printf("Number of re-hashes for the next state: %d\n", newState.CountNewHashes())
	fmt.Printf("Number of nodes exposed: %d\n", unionStats.ExposedCount)
	fmt.Printf("Number of nodes with height taken: %d\n", unionStats.HeightCount)

	diffStats := &cairo.Counters{}
	newState = cairo.Difference(state, stateChanges, diffStats)
	newState.GraphAndPicture("diff_new_" + outputNameFromInputName(stateFileName))

	fmt.Printf("\n#DIFFERENCE:\n")
	fmt.Printf("Number of nodes in the current state tree: %d\n", state.Size())
	fmt.Printf("Number of nodes in the state update tree: %d\n", stateChanges.Size())
	fmt.Printf("Number of nodes in the next state tree: %d\n", newState.Size())
	fmt.Printf("Number of re-hashes for the next state: %d\n", newState.CountNewHashes())
	fmt.Printf("Number of nodes exposed: %d\n", diffStats.ExposedCount)
	fmt.Printf("Number of nodes with height taken: %d\n", diffStats.HeightCount)
}
