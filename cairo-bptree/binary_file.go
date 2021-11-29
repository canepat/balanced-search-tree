package cairo_bptree

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const BLOCKSIZE int64 = 8 //32 //4096

type BinaryFile struct {
	path		string
	blockSize	int64
	size		int64
	file		*os.File
	opened		bool
}

func CreateRandomBinaryFile(prefix string, size int64) *BinaryFile {
	ensure(size%BLOCKSIZE == 0, fmt.Sprintf("CreateRandomBinaryFile: expected size multiple of 4k bytes, got %d\n", size))
	
	file, err := ioutil.TempFile("testdata/", prefix + strconv.FormatInt(size, 10) + "_")
	ensure(err == nil, fmt.Sprintf("CreateRandomBinaryFile: cannot create file %s, error %s\n", file.Name(), err))

	err = file.Truncate(size)
	ensure(err == nil, fmt.Sprintf("CreateRandomBinaryFile: cannot truncate file %s to %d, error %s\n", file.Name(), size, err))

	log.Infof("CreateRandomBinaryFile: created file %s\n", file.Name())

	buffer := make([]byte, BLOCKSIZE)
	for i := int64(0); i < size; i+= BLOCKSIZE {
		bytesRead, err := io.ReadFull(rand.Reader, buffer)
		ensure(bytesRead == len(buffer), fmt.Sprintf("insufficient bytes read %d, error %s\n", bytesRead, err))
		log.Tracef("CreateRandomBinaryFile: bytesRead=%d\n", bytesRead)
		bytesWritten, err := file.Write(buffer)

		ensure(bytesWritten == len(buffer), fmt.Sprintf("insufficient bytes written %d, error %s\n", bytesWritten, err))
		log.Tracef("CreateRandomBinaryFile: bytesWritten=%d\n", bytesWritten)
	}
	file.Seek(0, 0)

	binaryFile := &BinaryFile{
		path : file.Name(),
		blockSize: BLOCKSIZE,
		size: size,
		file: file,
		opened: true,
	}
	return binaryFile
}

func OpenBinaryFile(path string) *BinaryFile {
	file, err := os.Open(path)
	ensure(err == nil, fmt.Sprintf("OpenBinaryFile: cannot open file %s, error %s\n", path, err))

	info, err := file.Stat()
	ensure(err == nil, fmt.Sprintf("OpenBinaryFile: cannot stat file %s error %s\n", path, err))
	ensure(info.Size() >= 0, fmt.Sprintf("OpenBinaryFile: negative size %d file %s\n", info.Size(), path))

	binaryFile := &BinaryFile{
		path : path,
		blockSize: BLOCKSIZE,
		size: info.Size(),
		file: file,
		opened: true,
	}
	return binaryFile
}

func (f *BinaryFile) NewReader() *bufio.Reader {
	ensure(f.opened, fmt.Sprintf("Close: file %s is not opened\n", f.path))
	return bufio.NewReader(f.file)
}

func (f *BinaryFile) Close() {
	ensure(f.opened, fmt.Sprintf("Close: file %s is not opened\n", f.path))
	err := f.file.Close()
	ensure(err == nil, fmt.Sprintf("Close: cannot close file %s, error %s\n", f.path, err))
}
