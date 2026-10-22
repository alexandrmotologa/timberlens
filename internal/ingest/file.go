package ingest

import (
	"bufio"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/alexandrmotologa/timberlens/internal/parser"
)

// FileReader tails a file on disk continuously.
type FileReader struct {
	path     string
	consumer LineConsumer
	seqID    uint64
	running  int32
}

// NewFileReader creates a file tailer for the specified path.
func NewFileReader(path string, consumer LineConsumer) *FileReader {
	return &FileReader{
		path:     path,
		consumer: consumer,
	}
}

// Start tails the file until stopped.
func (fr *FileReader) Start() error {
	atomic.StoreInt32(&fr.running, 1)
	defer atomic.StoreInt32(&fr.running, 0)

	file, err := os.Open(fr.path)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for atomic.LoadInt32(&fr.running) == 1 {
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				// Wait for new writes
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return err
		}

		var fullLine []byte
		if isPrefix {
			fullLine = make([]byte, len(line), maxLineBuffer)
			copy(fullLine, line)
			for isPrefix && err == nil {
				var chunk []byte
				chunk, isPrefix, err = reader.ReadLine()
				fullLine = append(fullLine, chunk...)
				if len(fullLine) >= maxLineBuffer {
					break
				}
			}
		} else {
			fullLine = line
		}

		if len(fullLine) == 0 {
			continue
		}

		seq := atomic.AddUint64(&fr.seqID, 1)
		rec, _ := parser.ParseLine(fullLine, seq)
		if rec != nil && fr.consumer != nil {
			fr.consumer(rec)
		}
	}

	return nil
}

// Stop terminates the file tailer.
func (fr *FileReader) Stop() {
	atomic.StoreInt32(&fr.running, 0)
}
