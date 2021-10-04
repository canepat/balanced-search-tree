package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	cairo "github.com/canepat/bst/cairo-avl"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func readFromBinaryFile(binaryFilename string, readFunction func(*bufio.Reader) interface{}) interface{} {
	binaryFile, err := os.Open(binaryFilename)
	check(err)
	defer binaryFile.Close()
	stateReader := bufio.NewReader(binaryFile)
	return readFunction(stateReader)
}

func readStateFromBinaryFile(stateFilename string) (t *cairo.Node) {
	readFromBinaryFile(stateFilename, func(statesReader *bufio.Reader) interface{} {
		buffer := make([]byte, 4096)
		for {
			bytes_read, err := statesReader.Read(buffer)
			fmt.Println("BINARY state bytes read: ", bytes_read, " err: ", err)
			if err == io.EOF {
				break
			}
		}
		return t
	})
	return t
}

func readStateChangesFromBinaryFile(stateChangesFileName string) (t *cairo.Node) {
	readFromBinaryFile(stateChangesFileName, func(statesReader *bufio.Reader) interface{} {
		buffer := make([]byte, 4096)
		for {

			bytes_read, err := statesReader.Read(buffer)
			fmt.Println("CSV state bytes read: ", bytes_read, " err: ", err)
			if err == io.EOF {
				break
			}
		}
		return t
	})
	return t
}

func readFromCsvFile(csvFileName string, scanFromCsv func(*bufio.Scanner) interface{}) interface{} {
	csvFile, err := os.Open(csvFileName)
	check(err)
	defer csvFile.Close()
	stateScanner := bufio.NewScanner(csvFile)
	return scanFromCsv(stateScanner)
}

/*func readStateChangesFromCsvFile(stateChangesFilename string) (d *cairo.Dict, err error) {
	readFromCsvFile(stateChangesFilename, func(stateChanges *bufio.Scanner) interface{} {
		d, err = cairo.StateChangesFromCsv(stateChanges)
		d.GraphAndPicture("statechanges")
		return d
	})
	return d, err
}*/

func readStateFromCsvFile(stateFilename string) (t *cairo.Node, err error) {
	readFromCsvFile(stateFilename, func(state *bufio.Scanner) interface{} {
		t, err = cairo.StateFromCsv(state)
		t.GraphAndPicture("state")
		return t
	})
	return t, err
}

/*func readMappedStateFromCsvFile(stateFilename string) (t *cairo.Node, err error) {
	readFromCsvFile(stateFilename, func(state *bufio.Scanner) interface{} {
		t, err = cairo.MappedStateFromCsv(state)
		t.GraphAndPicture("mapped_state")
		return t
	})
	return t, err
}*/

func main() {
	if len(os.Args) < 4 {
		fmt.Println("main <stateFilename> <stateChangesFilename> <mappedStateFilename>")
		os.Exit(0)
	}
	stateFileName := os.Args[1]
	stateChangesFileName := os.Args[2]
	//mappedStateChangesFileName := os.Args[3]
	stateFileExt := filepath.Ext(stateFileName)
	stateChangesFileExt := filepath.Ext(stateChangesFileName)
	if stateFileExt == ".csv" && stateChangesFileExt == ".csv" {
		readStateFromCsvFile(stateFileName)
		//readStateChangesFromCsvFile(stateChangesFileName)
		//readMappedStateFromCsvFile(mappedStateChangesFileName)
	} else {
		readStateFromBinaryFile(stateFileName)
		readStateChangesFromBinaryFile(stateChangesFileName)
	}
}
