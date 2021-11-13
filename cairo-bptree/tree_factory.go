package cairo_bptree

import (
	"bufio"
	"encoding/binary"
	"io"
	"sort"

	log "github.com/sirupsen/logrus"
)

type Tree23Factory interface {
	NewTree23(reader *bufio.Reader) *Tree23
	NewUniqueKeyValues(reader *bufio.Reader) []KeyValue
}

type Tree23BinaryFactory struct {
	keySize int
}

func NewTree23BinaryFactory(keySize int) Tree23Factory {
	return &Tree23BinaryFactory{keySize: keySize}
}

func (factory *Tree23BinaryFactory) NewTree23(reader *bufio.Reader) *Tree23 {
	kvPairs := factory.readUniqueKeyValues(reader)
	sort.Sort(KeyValueByKey(kvPairs))
	log.Debugf("Number of keys for bulk loading: %d\n", len(kvPairs))
	return NewTree23(kvPairs)
}

func (factory *Tree23BinaryFactory) NewUniqueKeyValues(reader *bufio.Reader) []KeyValue {
	kvPairs := factory.readUniqueKeyValues(reader)
	sort.Sort(KeyValueByKey(kvPairs))
	return kvPairs
}

func (factory *Tree23BinaryFactory) readUniqueKeyValues(reader *bufio.Reader) []KeyValue {
	kvPairs := make([]KeyValue, 0)
	keyRegistry := make(map[Felt]bool)
	buffer := make([]byte, BufferSize)
	for {
		bytes_read, err := reader.Read(buffer)
		log.Tracef("BINARY state bytes read: %d err: %v\n", bytes_read, err)
		if err == io.EOF {
			break
		}
		key_bytes_count := factory.keySize * (bytes_read / factory.keySize)
		duplicated_keys := 0
		log.Tracef("BINARY state key_bytes_count: %d\n", key_bytes_count)
		for i := 0; i < key_bytes_count; i += factory.keySize {
			key := Felt(binary.BigEndian.Uint64(buffer[i:i+factory.keySize]))
			log.Debugf("BINARY state key: %d\n", key)
			if _, duplicated := keyRegistry[key]; duplicated {
				duplicated_keys++
				continue
			}
			value := key // Shortcut: value equal to key
			kvPairs = append(kvPairs, KeyValue{key, value})
		}
		log.Tracef("BINARY state duplicated_keys: %d\n", duplicated_keys)
	}
	return kvPairs
}
