package main

import (
	"flag"
	"os"

	cairo_bptree "github.com/canepat/bst/cairo-bptree"
	log "github.com/sirupsen/logrus"
)

const DEFAULT_GENERATE bool = false
const DEFAULT_KEY_SIZE uint = 8 // TODO: 4 and change Felt to []byte
const DEFAULT_NESTED bool = false
const DEFAULT_LOG_LEVEL string = "INFO"
const DEFAULT_GRAPH bool = false

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
	flag.BoolVar(&options.graph, "graph", DEFAULT_GRAPH, "flag indicating if tree graph should be saved or not")
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
	graph			bool
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

	var stateFile, stateChangesFile *cairo_bptree.BinaryFile
	defer func () {
		if stateFile != nil {
			stateFile.Close()
		}
		if stateChangesFile != nil {
			stateChangesFile.Close()
		}
	}()
	if options.generate {
		log.Printf("Creating random binary state file...\n")
		stateFile = cairo_bptree.CreateRandomBinaryFile("state", int64(options.stateFileSize))
		log.Printf("Random binary state file created: %s\n", stateFile.Name())
		log.Printf("Creating random binary state-changes file...\n")
		stateChangesFile = cairo_bptree.CreateRandomBinaryFile("statechanges", int64(options.stateChangesFileSize))
		log.Printf("Random binary state-changes file created: %s\n", stateChangesFile.Name())
	} else {
		stateFile = cairo_bptree.OpenBinaryFile(options.stateFileName)
		log.Printf("Random binary state file opened: %s, size=%d\n", stateFile.Name(), stateFile.Size())
		stateChangesFile = cairo_bptree.OpenBinaryFile(options.stateChangesFileName)
		log.Printf("Random binary state-changes file opened: %s, size=%d\n", stateChangesFile.Name(), stateChangesFile.Size())

		keyFactory := cairo_bptree.NewKeyBinaryFactory(int(options.keySize))
		log.Printf("Reading unique key-value pairs from: %s\n", stateFile.Name())
		kvPairs := keyFactory.NewUniqueKeyValues(stateFile.NewReader())
		log.Printf("Creating tree with #kvPairs=%v\n", kvPairs.Len())
		state := cairo_bptree.NewTree23(kvPairs)
		log.Printf("Created tree: %v\n", state)
		log.Printf("Reading unique key-value pairs from: %s\n", stateChangesFile.Name())
		stateChanges := keyFactory.NewUniqueKeyValues(stateChangesFile.NewReader())
	
		log.Printf("UPSERT: number of nodes in the current state tree: %d\n", state.Size())
		log.Printf("UPSERT: number of state changes: %d\n", stateChanges.Len())
	
		if options.graph {
			state.GraphAndPicture("state")
		}
	
		stats := &cairo_bptree.Stats{}
		newState := state.Upsert(stateChanges, stats)
		rehashedNodes := newState.CountNewHashes()
	
		log.Printf("UPSERT: number of nodes in the next state tree: %d\n", newState.Size())
		log.Printf("UPSERT: number of re-hashed nodes for the next state: %d\n", rehashedNodes)
		log.Printf("UPSERT: number of existing nodes exposed: %d\n", stats.ExposedCount)
		log.Printf("UPSERT: number of new nodes exposed: %d\n", rehashedNodes-stats.ExposedCount)
	
		if options.graph {
			newState.GraphAndPicture("newState")
		}
	}
}
