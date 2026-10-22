package ingest

import (
	"bufio"
	"io"
	"os"
	"sync/atomic"

	"github.com/alexandrmotologa/timberlens/internal/model"
	"github.com/alexandrmotologa/timberlens/internal/parser"
)

const maxLineBuffer = 1024 * 1024 // 1 MB line max

// LineConsumer is a callback function invoked for every parsed log record.
type LineConsumer func(rec *model.LogRecord)

// StdinReader reads and parses log streams from STDIN or any io.Reader.
type StdinReader struct {
	reader   io.Reader
	consumer LineConsumer
	seqID    uint64
	running  int32
}

// NewStdinReader creates a new StdinReader.
func NewStdinReader(r io.Reader, consumer LineConsumer) *StdinReader {
	if r == nil {
		r = os.Stdin
	}
	return &StdinReader{
		reader:   r,
		consumer: consumer,
	}
}

// Start begins reading and parsing in the background. It returns when the stream hits EOF or errors.
func (sr *StdinReader) Start() error {
	atomic.StoreInt32(&sr.running, 1)
	defer atomic.StoreInt32(&sr.running, 0)

	reader := bufio.NewReaderSize(sr.reader, 64*1024)

	for atomic.LoadInt32(&sr.running) == 1 {
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		// Handle long lines that exceed the buffer
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

		seq := atomic.AddUint64(&sr.seqID, 1)
		rec, _ := parser.ParseLine(fullLine, seq)
		if rec != nil && sr.consumer != nil {
			sr.consumer(rec)
		}
	}

	return nil
}

// Stop signals the reader to stop consuming lines.
func (sr *StdinReader) Stop() {
	atomic.StoreInt32(&sr.running, 0)
}

// TotalLines returns the number of lines processed.
func (sr *StdinReader) TotalLines() uint64 {
	return atomic.LoadUint64(&sr.seqID)
}
