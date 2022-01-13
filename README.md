# Balanced search tree prototyping
Python and Go implementation of balanced search tree data structures

## Structure

```
avl
cairo-avl
cairo-bptree
cmd
    cairo-avl
    cairo-bptree
README.md
```

- The `avl` folder contains both Python and Go implementations of the (unnested) self-balancing AVL trees described in the [BFS16.pdf](https://www.cs.cmu.edu/~guyb/papers/BFS16.pdf) paper
- The `cairo-avl` folder contains the Go implementation of (nested and unnested) AVL tree variant suitable for representing contract-based blockchain state
- The `cairo-bptree` folder contains the Go implementation of (unnested) B+ tree variant suitable for representing contract-based blockchain state

## Usage

### AVL variant

This implementation supports building state and state-changes AVL-trees from *existing* binary files.

#### Build

```
$ cd cmd/cairo-avl
$ go build
```

#### Synopsis

```
$ ./cairo-avl --help
Usage of ./cairo-avl:
  -graph
        flag indicating if tree graph should be saved or not
  -keySize int
        the key size in bytes (default 8)
  -logLevel string
        the logging level (default "INFO")
  -nested
        flag indicating if tree should be nested or not
  -stateChangesFileName string
        the state-change file name
  -stateFileName string
        the state file name
```

#### Example
```
./cairo-avl -stateFileName=state30 -stateChangesFileName=statechanges10 -keySize=1 -graph -nested=false
```

### B+tree variant

This implementation supports both generating random state and state-changes binary files and building state and state-changes B+trees from *generated on-the-fly* or *existing* binary files.

#### Build

```
$ cd cmd/cairo-bptree
$ go build
```

#### Synopsis

```
$ ./cairo-bptree
Usage of ./cairo-bptree:
  -generate
        flag indicating if binary files shall be generated or not
  -graph
        flag indicating if tree graph should be saved or not
  -keySize uint
        the key size in bytes (default 8)
  -logLevel string
        the logging level (default "INFO")
  -nested
        flag indicating if tree should be nested or not
  -stateChangesFileName string
        the state-change file name
  -stateChangesFileSize uint
        the state-change file size in bytes
  -stateFileName string
        the state file name
  -stateFileSize uint
        the state file size in bytes
```

#### Example

To generate state and state-changes binary files and use them to execute bulk upsert and bulk delete:

```
./cairo-bptree -generate -stateFileSize=1073741824 -stateChangesFileSize=104857600
```

To build state and state-changes trees and execute bulk upsert and bulk delete from binary files using 1-byte keys:
```
./cairo-bptree -stateFileName=state30 -stateChangesFileName=statechanges10 -keySize=1
```