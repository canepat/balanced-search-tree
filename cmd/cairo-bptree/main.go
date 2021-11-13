package main

import (
	"flag"

	cairo_bptree "github.com/canepat/bst/cairo-bptree"
	log "github.com/sirupsen/logrus"
)

const DEFAULT_STATE_FILE_SIZE uint64 = 256//16384
const DEFAULT_STATE_CHANGES_FILE_SIZE uint64 = 64//4096
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
	flag.Uint64Var(&options.stateFileSize, "s", DEFAULT_STATE_FILE_SIZE, "the state file size in bytes")
	flag.Uint64Var(&options.stateChangesFileSize, "c", DEFAULT_STATE_CHANGES_FILE_SIZE, "the state-change file size in bytes")
	flag.UintVar(&options.keySize, "k", DEFAULT_KEY_SIZE, "the key size in bytes")
	flag.BoolVar(&options.nested, "n", DEFAULT_NESTED, "flag indicating if tree should be nested or not")
	flag.StringVar(&options.logLevel, "l", DEFAULT_LOG_LEVEL, "the logging level")
}

type Options struct {
	stateFileSize		uint64
	stateChangesFileSize	uint64
	keySize			uint
	nested			bool
	logLevel		string
}

func main() {
	flag.Parse()

	level, _ := log.ParseLevel(options.logLevel)
	log.SetLevel(level)

	log.Printf("Size of the state file in bytes: %d\n", options.stateFileSize)
	log.Printf("Size of the state-changes file in bytes: %d\n", options.stateChangesFileSize)
	log.Printf("Size of the key in bytes: %d\n", options.keySize)
	log.Printf("Trees are nested: %t\n", options.nested)
	log.Printf("Log level: %s\n", options.logLevel)

	stateFile := cairo_bptree.CreateRandomBinaryFile("state", int64(options.stateFileSize))
	defer stateFile.Close()
	stateChangesFile := cairo_bptree.CreateRandomBinaryFile("statechanges", int64(options.stateChangesFileSize))
	defer stateChangesFile.Close()

	treeFactory := cairo_bptree.NewTree23BinaryFactory(int(options.keySize))
	state := treeFactory.NewTree23(stateFile.NewReader())
	stateChanges := treeFactory.NewUniqueKeyValues(stateChangesFile.NewReader())

	stats := &cairo_bptree.Stats{}
	newState := state.Upsert(stateChanges, stats)

	log.Printf("UPSERT: Number of nodes in the current state tree: %d\n", state.Size())
	log.Printf("UPSERT: Number of state changes: %d\n", len(stateChanges))
	log.Printf("UPSERT: Number of nodes in the next state tree: %d\n", newState.Size())
	log.Printf("UPSERT: Number of re-hashes for the next state: %d\n", newState.CountNewHashes())
	log.Printf("UPSERT: Number of nodes exposed: %d\n", stats.ExposedCount)
}
