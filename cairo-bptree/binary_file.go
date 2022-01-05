package cairo_bptree

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"strconv"
)

const BLOCKSIZE int64 = 4096

type BinaryFile struct {
	path		string
	blockSize	int64
	size		int64
	file		*os.File
	opened		bool
}

func CreateRandomBinaryFile(path string, size int64) *BinaryFile {
	file, err := os.OpenFile(path + strconv.FormatInt(size, 10), os.O_RDWR|os.O_CREATE, 0644)
	ensure(err == nil, fmt.Sprintf("CreateRandomBinaryFile: cannot create file %s, error %s\n", file.Name(), err))

	err = file.Truncate(size)
	ensure(err == nil, fmt.Sprintf("CreateRandomBinaryFile: cannot truncate file %s to %d, error %s\n", file.Name(), size, err))

	bufferedFile := bufio.NewWriter(file)
	numBlocks := size / BLOCKSIZE
	remainderSize := size % BLOCKSIZE
	buffer := make([]byte, BLOCKSIZE)
	for i := int64(0); i <= numBlocks; i++ {
		if i == numBlocks {
			buffer = make([]byte, remainderSize)
		}
		bytesRead, err := io.ReadFull(rand.Reader, buffer)
		ensure(bytesRead == len(buffer), fmt.Sprintf("CreateRandomBinaryFile: insufficient bytes read %d, error %s\n", bytesRead, err))
		bytesWritten, err := bufferedFile.Write(buffer)
		ensure(bytesWritten == len(buffer), fmt.Sprintf("CreateRandomBinaryFile: insufficient bytes written %d, error %s\n", bytesWritten, err))
	}

	err = bufferedFile.Flush()
	ensure(err == nil, fmt.Sprintf("CreateRandomBinaryFile: error during flushing %s\n", err))

	binaryFile := &BinaryFile{
		path : file.Name(),
		blockSize: BLOCKSIZE,
		size: size,
		file: file,
		opened: true,
	}
	binaryFile.rewind()
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

func (f *BinaryFile) rewind() {
	offset, err := f.file.Seek(0, io.SeekStart)
	ensure(err == nil, fmt.Sprintf("rewind: error during seeking %s\n", err))
	ensure(offset == 0, fmt.Sprintf("rewind: unexpected offset after seeking: %d\n", offset))
}

func (f *BinaryFile) Name() string {
	return f.path
}

func (f *BinaryFile) Size() int64 {
	return f.size
}

func (f *BinaryFile) NewReader() *bufio.Reader {
	ensure(f.opened, fmt.Sprintf("NewReader: file %s is not opened\n", f.path))
	f.rewind()
	return bufio.NewReader(f.file)
}

func (f *BinaryFile) Close() {
	ensure(f.opened, fmt.Sprintf("Close: file %s is not opened\n", f.path))
	err := f.file.Close()
	ensure(err == nil, fmt.Sprintf("Close: cannot close file %s, error %s\n", f.path, err))
	f.opened = false
}
