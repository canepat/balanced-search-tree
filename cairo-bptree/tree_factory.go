package cairo_bptree

import (
	"bufio"
	"encoding/binary"
	"io"

	log "github.com/sirupsen/logrus"
)

type Tree23Factory interface {
	NewTree23(reader *bufio.Reader) *Tree23
}

type Tree23BinaryFactory struct {
	keySize int
}

func (factory *Tree23BinaryFactory) NewTree23(reader *bufio.Reader) *Tree23 {
	kvPairs := make([]KeyValue, 0)
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
			key := binary.BigEndian.Uint64(buffer[i:i+factory.keySize])
			log.Debugf("BINARY state key: %d\n", key)
			value := key // Shortcut: value equal to key
			kv := KeyValue{&key, &value}
			kvPairs = append(kvPairs, kv)
		}
		log.Tracef("BINARY state duplicated_keys: %d\n", duplicated_keys)
	}
	log.Debugf("Number of keys for bulk loading: %d\n", len(kvPairs))
	tree := &Tree23{}
	return tree.Upsert(kvPairs)
}
