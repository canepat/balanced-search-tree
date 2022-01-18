package cairo_bptree

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
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

type RandomBinaryReader struct {
	sourceFile	*BinaryFile
	keySize		int
}
func (r RandomBinaryReader) Read(b []byte) (n int, err error) {
	numKeys := len(b) / r.keySize
	for i := 0; i < numKeys; i++ {
		bytesRead, err := r.readAtRandomOffset(b[i*r.keySize:i*r.keySize+r.keySize])
		if err != nil {
			return i*r.keySize + bytesRead, fmt.Errorf("cannot random read at iteration %d: %v", i, err)
		}
		n += bytesRead
	}
	remainderSize := len(b) % r.keySize
	bytesRead, err := r.readAtRandomOffset(b[numKeys*r.keySize:numKeys*r.keySize+remainderSize])
	if err != nil {
		return numKeys*r.keySize + bytesRead, fmt.Errorf("cannot random read remainder %d: %v", remainderSize, err)
	}
	n += bytesRead
	return n, nil
}

func (r RandomBinaryReader) readAtRandomOffset(b []byte) (n int, err error) {
	randomValue, err := rand.Int(rand.Reader, big.NewInt(r.sourceFile.size - int64(len(b))))
	if err != nil {
		return 0, fmt.Errorf("cannot generate random offset: %v", err)
	}
	randomOffset := randomValue.Int64()
	_, err = r.sourceFile.file.Seek(randomOffset, io.SeekStart)
	if err != nil {
		return 0, fmt.Errorf("cannot seek to offset %d: %v", randomOffset, err)
	}
	bytesRead, err := r.sourceFile.file.Read(b)
	if err != nil {
		return 0, fmt.Errorf("cannot read from source file: %v", err)
	}
	return bytesRead, nil
}

func CreateBinaryFileByRandomSampling(path string, size int64, sourceFile *BinaryFile, keySize int) *BinaryFile {
	return CreateBinaryFileFromReader(path, "_onlyexisting", size, RandomBinaryReader{sourceFile, keySize})
}

func CreateBinaryFileByPRNG(path string, size int64) *BinaryFile {
	return CreateBinaryFileFromReader(path, "", size, rand.Reader)
}

func CreateBinaryFileFromReader(path, suffix string, size int64, reader io.Reader) *BinaryFile {
	file, err := os.OpenFile(path + strconv.FormatInt(size, 10) + suffix, os.O_RDWR|os.O_CREATE, 0644)
	ensure(err == nil, fmt.Sprintf("CreateBinaryFileFromReader: cannot create file %s, error %s\n", file.Name(), err))

	err = file.Truncate(size)
	ensure(err == nil, fmt.Sprintf("CreateBinaryFileFromReader: cannot truncate file %s to %d, error %s\n", file.Name(), size, err))

	bufferedFile := bufio.NewWriter(file)
	numBlocks := size / BLOCKSIZE
	remainderSize := size % BLOCKSIZE
	buffer := make([]byte, BLOCKSIZE)
	for i := int64(0); i <= numBlocks; i++ {
		if i == numBlocks {
			buffer = make([]byte, remainderSize)
		}
		bytesRead, err := io.ReadFull(reader, buffer)
		ensure(bytesRead == len(buffer), fmt.Sprintf("CreateBinaryFileFromReader: insufficient bytes read %d, error %s\n", bytesRead, err))
		bytesWritten, err := bufferedFile.Write(buffer)
		ensure(bytesWritten == len(buffer), fmt.Sprintf("CreateBinaryFileFromReader: insufficient bytes written %d, error %s\n", bytesWritten, err))
	}

	err = bufferedFile.Flush()
	ensure(err == nil, fmt.Sprintf("CreateBinaryFileFromReader: error during flushing %s\n", err))

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
