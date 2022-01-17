package main

import (
	"flag"
	"os"

	cairo_bptree "github.com/canepat/bst/cairo-bptree"
	log "github.com/sirupsen/logrus"
)

const DEFAULT_GENERATE bool = false
const DEFAULT_ONLY_EXISTING_KEYS bool = false
const DEFAULT_KEY_SIZE uint = 4
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
	flag.BoolVar(&options.onlyExistingKeys, "onlyExistingKeys", DEFAULT_ONLY_EXISTING_KEYS, "flag indicating if only existing keys should be included in state changes or not")
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
	onlyExistingKeys	bool
	stateFileSize		uint64
	stateChangesFileSize	uint64
	stateFileName		string
	stateChangesFileName	string
	keySize			uint
	nested			bool
	logLevel		string
	graph			bool
}

func bulkUpsert(keyFactory cairo_bptree.KeyFactory, kvPairs, stateChanges cairo_bptree.KeyValues) {
	log.Printf("UPSERT: creating tree with #kvPairs=%v\n", kvPairs.Len())
	state := cairo_bptree.NewTree23(kvPairs)
	log.Printf("UPSERT: created tree: %v\n", state)

	if options.graph {
		state.GraphAndPicture("state")
	}

	log.Printf("UPSERT: number of nodes in the current state tree: %d\n", state.Size())
	log.Printf("UPSERT: number of state changes: %d\n", stateChanges.Len())
	log.Debugf("UPSERT: state changes as key-value pairs: %v\n", stateChanges)

	stats := &cairo_bptree.Stats{}
	stateAfterUpsert := state.UpsertWithStats(stateChanges, stats)

	log.Printf("UPSERT: number of nodes in the next state tree: %d\n", stateAfterUpsert.Size())
	log.Printf("UPSERT: number of re-hashed nodes for the next state: %d\n", stats.RehashedCount)
	log.Printf("UPSERT: number of existing nodes exposed: %d\n", stats.ExposedCount)
	log.Printf("UPSERT: number of hashes (opening): %d\n", stats.OpeningHashes)
	log.Printf("UPSERT: number of new nodes exposed: %d\n", stats.RehashedCount-stats.ExposedCount)
	log.Printf("UPSERT: number of created nodes: %d\n", stats.CreatedCount)
	log.Printf("UPSERT: number of updated values: %d\n", stats.UpdatedCount)
	log.Printf("UPSERT: number of hashes (closing): %d\n", stats.ClosingHashes)

	if options.graph {
		stateAfterUpsert.GraphAndPicture("stateAfterUpsert")
	}
}

func bulkDelete(keyFactory cairo_bptree.KeyFactory, kvPairs cairo_bptree.KeyValues, stateDeletes cairo_bptree.Keys) {
	log.Printf("DELETE: creating tree with #kvPairs=%v\n", kvPairs.Len())
	state := cairo_bptree.NewTree23(kvPairs)
	log.Printf("DELETE: created tree: %v\n", state)

	log.Printf("DELETE: number of nodes in the current state tree: %d\n", state.Size())
	log.Printf("DELETE: number of state deletes: %d\n", stateDeletes.Len())
	log.Debugf("DELETE: state deletes as keys: %v\n", stateDeletes)

	stats := &cairo_bptree.Stats{}
	stateAfterDelete := state.DeleteWithStats(stateDeletes, stats)

	log.Printf("DELETE: number of nodes in the next state tree: %d\n", stateAfterDelete.Size())
	log.Printf("DELETE: number of re-hashed nodes for the next state: %d\n", stats.RehashedCount)
	log.Printf("DELETE: number of existing nodes exposed: %d\n", stats.ExposedCount)
	log.Printf("DELETE: number of hashes (opening): %d\n", stats.OpeningHashes)
	log.Printf("DELETE: number of nodes exposed unchanged: %d\n", stats.ExposedCount-stats.RehashedCount)
	log.Printf("DELETE: number of deleted nodes: %d\n", stats.DeletedCount)
	log.Printf("DELETE: number of updated nodes: %d\n", stats.UpdatedCount)
	log.Printf("DELETE: number of hashes (closing): %d\n", stats.ClosingHashes)

	if options.graph {
		stateAfterDelete.GraphAndPicture("stateAfterDelete")
	}
}

func main() {
	flag.Parse()

	generate := options.generate
	onlyExistingKeys := options.onlyExistingKeys
	stateFileSize := options.stateFileSize
	stateChangesFileSize := options.stateChangesFileSize
	stateFileName := options.stateFileName
	stateChangesFileName:= options.stateChangesFileName
	keySize := options.keySize
	nested := options.nested
	logLevel := options.logLevel

	if generate {
		if stateFileSize == 0 || stateChangesFileSize == 0 {
			log.Errorln("both -stateFileSize and -stateChangesFileSize must be present when -generate=true")
			flag.Usage()
			os.Exit(0)
		}
	} else {
		if stateFileName == "" || stateChangesFileName == "" {
			log.Errorln("both -stateFileName and -stateChangesFileName must be present when -generate=false")
			flag.Usage()
			os.Exit(0)
		}
	}

	level, _ := log.ParseLevel(logLevel)
	log.SetLevel(level)

	log.Printf("Generate state and state-changes files: %t\n", generate)
	if generate {
		log.Printf("Size of the state file in bytes: %d\n", stateFileSize)
		log.Printf("Size of the state-changes file in bytes: %d\n", stateChangesFileSize)
	} else {
		log.Printf("Name of the state file: %s\n", stateFileName)
		log.Printf("Name of the state-changes file: %s\n", stateChangesFileName)
	}
	log.Printf("Size of the key in bytes: %d\n", keySize)
	log.Printf("Trees are nested: %t\n", nested)

	var stateFile, stateChangesFile *cairo_bptree.BinaryFile

	if generate {
		log.Printf("Creating random binary state file...\n")
		stateFile = cairo_bptree.CreateBinaryFileByPRNG("state", int64(stateFileSize))
		defer stateFile.Close()
		log.Printf("Random binary state file created: %s\n", stateFile.Name())
		if onlyExistingKeys {
			log.Printf("Creating random binary state-changes file from PRNG...\n")
			stateChangesFile = cairo_bptree.CreateBinaryFileByPRNG("statechanges", int64(stateChangesFileSize))
		} else {
			log.Printf("Creating random binary state-changes file from state file...\n")
			stateChangesFile = cairo_bptree.CreateBinaryFileByRandomSampling("statechanges", int64(stateChangesFileSize), stateFile, int(keySize))
		}
		defer stateChangesFile.Close()
		log.Printf("Random binary state-changes file created: %s\n", stateChangesFile.Name())
	} else {
		stateFile = cairo_bptree.OpenBinaryFile(stateFileName)
		defer stateFile.Close()
		log.Printf("Random binary state file opened: %s, size=%d\n", stateFile.Name(), stateFile.Size())

		stateChangesFile = cairo_bptree.OpenBinaryFile(stateChangesFileName)
		defer stateChangesFile.Close()
		log.Printf("Random binary state-changes file opened: %s, size=%d\n", stateChangesFile.Name(), stateChangesFile.Size())
	}

	keyFactory := cairo_bptree.NewKeyBinaryFactory(int(keySize))
	log.Printf("Reading unique key-value pairs from: %s\n", stateFile.Name())
	kvPairs := keyFactory.NewUniqueKeyValues(stateFile.NewReader())

	log.Printf("Reading unique key-value pairs from: %s\n", stateChangesFile.Name())
	stateChanges := keyFactory.NewUniqueKeyValues(stateChangesFile.NewReader())
	bulkUpsert(keyFactory, kvPairs, stateChanges)

	log.Printf("Reading unique keys from: %s\n", stateChangesFile.Name())
	stateDeletes := keyFactory.NewUniqueKeys(stateChangesFile.NewReader())
	bulkDelete(keyFactory, kvPairs, stateDeletes)
}
