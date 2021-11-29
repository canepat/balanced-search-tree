package main

import (
	"flag"
	"os"

	cairo_bptree "github.com/canepat/bst/cairo-bptree"
	log "github.com/sirupsen/logrus"
)

const DEFAULT_GENERATE bool = true
//const DEFAULT_STATE_FILE_SIZE uint64 = 16384 //256
//const DEFAULT_STATE_CHANGES_FILE_SIZE uint64 = 4096 //64
const DEFAULT_KEY_SIZE uint = 8 // TODO: 4 and change Felt to []byte
const DEFAULT_NESTED bool = false
const DEFAULT_LOG_LEVEL string = "INFO"

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
	flag.BoolVar(&options.generate, "generate", DEFAULT_GENERATE, "flag indicating if binary files shall be generated or not")
	flag.Uint64Var(&options.stateFileSize, "stateFileSize", 0, "the state file size in bytes")
	flag.Uint64Var(&options.stateChangesFileSize, "stateChangesFileSize", 0, "the state-change file size in bytes")
	flag.StringVar(&options.stateFileName, "stateFileName", "", "the state file name")
	flag.StringVar(&options.stateChangesFileName, "stateChangesFileName", "", "the state-change file name")
	flag.UintVar(&options.keySize, "keySize", DEFAULT_KEY_SIZE, "the key size in bytes")
	flag.BoolVar(&options.nested, "nested", DEFAULT_NESTED, "flag indicating if tree should be nested or not")
	flag.StringVar(&options.logLevel, "logLevel", DEFAULT_LOG_LEVEL, "the logging level")
}

type Options struct {
	generate		bool
	stateFileSize		uint64
	stateChangesFileSize	uint64
	stateFileName		string
	stateChangesFileName	string
	keySize			uint
	nested			bool
	logLevel		string
}

func main() {
	flag.Parse()

	if options.generate {
		if options.stateFileSize == 0 || options.stateChangesFileSize == 0 {
			flag.Usage()
			os.Exit(0)
		}
	} else {
		if options.stateFileName == "" || options.stateChangesFileName == "" {
			flag.Usage()
			os.Exit(0)
		}
	}

	level, _ := log.ParseLevel(options.logLevel)
	log.SetLevel(level)

	log.Printf("Generate state and state-changes files: %t\n", options.generate)
	if options.generate {
		log.Printf("Size of the state file in bytes: %d\n", options.stateFileSize)
		log.Printf("Size of the state-changes file in bytes: %d\n", options.stateChangesFileSize)
	} else {
		log.Printf("Name of the state file: %s\n", options.stateFileName)
		log.Printf("Name of the state-changes file: %s\n", options.stateChangesFileName)
	}
	log.Printf("Size of the key in bytes: %d\n", options.keySize)
	log.Printf("Trees are nested: %t\n", options.nested)
	log.Printf("Log level: %s\n", options.logLevel)

	var stateFile, stateChangesFile *cairo_bptree.BinaryFile
	if options.generate {
		stateFile = cairo_bptree.CreateRandomBinaryFile("state", int64(options.stateFileSize))
		stateChangesFile = cairo_bptree.CreateRandomBinaryFile("statechanges", int64(options.stateChangesFileSize))
	} else {
		stateFile = cairo_bptree.OpenBinaryFile(options.stateFileName)
		stateChangesFile = cairo_bptree.OpenBinaryFile(options.stateChangesFileName)
	}
	defer stateFile.Close()
	defer stateChangesFile.Close()


	treeFactory := cairo_bptree.NewTree23BinaryFactory(int(options.keySize))
	state := treeFactory.NewTree23(stateFile.NewReader())
	state.GraphAndPicture("state")
	stateChanges := treeFactory.NewUniqueKeyValues(stateChangesFile.NewReader())

	stats := &cairo_bptree.Stats{}
	newState := state.Upsert(stateChanges, stats)
	newState.GraphAndPicture("newState")

	log.Printf("UPSERT: Number of nodes in the current state tree: %d\n", state.Size())
	log.Printf("UPSERT: Number of state changes: %d\n", len(stateChanges))
	log.Printf("UPSERT: Number of nodes in the next state tree: %d\n", newState.Size())
	log.Printf("UPSERT: Number of re-hashes for the next state: %d\n", newState.CountNewHashes())
	log.Printf("UPSERT: Number of nodes exposed: %d\n", stats.ExposedCount)
}
