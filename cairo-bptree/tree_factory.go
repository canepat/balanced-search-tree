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
			key := factory.readKey(buffer, i)
			log.Tracef("BINARY state key: %d\n", key)
			if _, duplicated := keyRegistry[key]; duplicated {
				duplicated_keys++
				continue
			}
			keyRegistry[key] = true
			log.Debugf("BINARY state unique key: %d\n", key)
			value := key // Shortcut: value equal to key
			kvPairs = append(kvPairs, KeyValue{key, value})
		}
		log.Tracef("BINARY state duplicated_keys: %d\n", duplicated_keys)
	}
	return kvPairs
}

func (factory *Tree23BinaryFactory) readKey(buffer []byte, offset int) Felt {
	keySlice := buffer[offset:offset+factory.keySize]
	switch factory.keySize {
	case 1:
		return Felt(keySlice[0])
	case 2:
		return Felt(binary.BigEndian.Uint16(keySlice))
	case 4:
		return Felt(binary.BigEndian.Uint32(keySlice))
	default:
		return Felt(binary.BigEndian.Uint64(keySlice))
	}
}
