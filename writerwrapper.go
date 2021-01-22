package brotli

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
)

type writerWrapper struct {
	Filters          []ResponseHeaderFilter
	MinContentLength int64
	OriginWriter     http.ResponseWriter
	brotliWriter     *brotli.Writer
	GetBrotliWriter  func() *brotli.Writer
	PutBrotliWriter  func(*brotli.Writer)

	shouldCompress        bool
	bodyBigEnough         bool
	headerFlushed         bool
	responseHeaderChecked bool
	statusCode            int
	size                  int
	bodyBuffer            []byte
}

// interface verification
var _ http.ResponseWriter = &writerWrapper{}
var _ http.Flusher = &writerWrapper{}

func newWriterWrapper(filters []ResponseHeaderFilter,
	minContentLength int64,
	originWriter http.ResponseWriter,
	getBrotliWriter func() *brotli.Writer,
	putBrotliWriter func(*brotli.Writer)) *writerWrapper {

	return &writerWrapper{
		shouldCompress:   true,
		bodyBuffer:       make([]byte, 0, minContentLength),
		Filters:          filters,
		MinContentLength: minContentLength,
		OriginWriter:     originWriter,
		GetBrotliWriter:  getBrotliWriter,
		PutBrotliWriter:  putBrotliWriter,
	}
}

// Reset the wrapper into a fresh one,
// writing to originWriter
func (w *writerWrapper) Reset(originWriter http.ResponseWriter) {
	w.OriginWriter = originWriter

	// internal below

	// reset status with caution
	// all internal fields should be taken good care
	w.shouldCompress = true
	w.headerFlushed = false
	w.responseHeaderChecked = false
	w.bodyBigEnough = false
	w.statusCode = 0
	w.size = 0

	if w.brotliWriter != nil {
		w.PutBrotliWriter(w.brotliWriter)
		w.brotliWriter = nil
	}
	if w.bodyBuffer != nil {
		w.bodyBuffer = w.bodyBuffer[:0]
	}
}

// Status implement the gin.ResponseWriter interface.
func (w *writerWrapper) Status() int {
	return w.statusCode
}

// Size implement the gin.ResponseWriter interface.
func (w *writerWrapper) Size() int {
	return w.size
}

// Written implement the gin.ResponseWriter interface.
func (w *writerWrapper) Written() bool {
	return w.headerFlushed || len(w.bodyBuffer) > 0
}

func (w *writerWrapper) WriteHeaderCalled() bool {
	return w.statusCode != 0
}

// initBrotliWriter
func (w *writerWrapper) initBrotliWriter() {
	w.brotliWriter = w.GetBrotliWriter()
	w.brotliWriter.Reset(w.OriginWriter)
}

// Header implements the http.ResponseWriter interface.
func (w *writerWrapper) Header() http.Header {
	return w.OriginWriter.Header()
}

// Write implements the http.ResponseWriter interface.
func (w *writerWrapper) Write(data []byte) (int, error) {
	w.size += len(data)

	if !w.WriteHeaderCalled() {
		w.WriteHeader(http.StatusOK)
	}

	if !w.shouldCompress {
		return w.OriginWriter.Write(data)
	}
	if w.bodyBigEnough {
		return w.brotliWriter.Write(data)
	}

	// fast check
	if !w.responseHeaderChecked {
		w.responseHeaderChecked = true

		// 响应数据校验
		header := w.Header()
		for _, filter := range w.Filters {
			w.shouldCompress = filter.ShouldCompress(header)
			if !w.shouldCompress {
				w.WriteHeaderNow()
				return w.OriginWriter.Write(data)
			}
		}
	}

	// check buffer length
	if !w.writeBuffer(data) {
		w.bodyBigEnough = true

		// detect Content-Type if there's none
		if header := w.Header(); header.Get("Content-Type") == "" {
			header.Set("Content-Type", http.DetectContentType(w.bodyBuffer))
		}

		w.WriteHeaderNow()
		w.initBrotliWriter()
		if len(w.bodyBuffer) > 0 {
			written, err := w.brotliWriter.Write(w.bodyBuffer)
			if err != nil {
				err = fmt.Errorf("w.brotliWriter.Write: %w", err)
				return written, err
			}
		}
		return w.brotliWriter.Write(data)
	}

	return len(data), nil
}

// writeBuffer
func (w *writerWrapper) writeBuffer(data []byte) bool {
	if int64(len(data)+len(w.bodyBuffer)) > w.MinContentLength {
		return false
	}

	w.bodyBuffer = append(w.bodyBuffer, data...)
	return true
}

// WriteHeader implements the http.ResponseWriter interface.
//
// WriteHeader does not really calls originalHandler's WriteHeader,
// and the calling will actually be handler by WriteHeaderNow().
//
// http.ResponseWriter does not specify clearly whether permitting
// updating status code on second call to WriteHeader(), and it's
// conflicting between http and gin's implementation.
// Here, brotli consider second(and furthermore) calls to WriteHeader()
// valid. WriteHeader() is disabled after flushing header.
// Do note setting status not 200 marks content uncompressable,
// and a later status code change does not revert this.
func (w *writerWrapper) WriteHeader(statusCode int) {
	if w.headerFlushed {
		return
	}

	w.statusCode = statusCode
	// write http status
	w.OriginWriter.WriteHeader(w.statusCode)

	if !w.shouldCompress {
		return
	}

	if statusCode != http.StatusOK {
		w.shouldCompress = false
		return
	}
}

// WriteHeaderNow implement the gin.ResponseWriter interface.
// WriteHeaderNow Forces to write the http header (status code + headers).
//
// WriteHeaderNow must always be called and called after
// WriteHeader() is called and
// w.shouldCompress is decided.
//
// This method is usually called by gin's AbortWithStatus()
func (w *writerWrapper) WriteHeaderNow() {
	if w.headerFlushed {
		return
	}

	// if neither WriteHeader() or Write() are called,
	// do nothing
	if !w.WriteHeaderCalled() {
		return
	}

	if w.shouldCompress {
		header := w.Header()
		header.Del("Content-Length")
		header.Set("Content-Encoding", "br")
		header.Add("Vary", "Accept-Encoding")
		originalEtag := w.Header().Get("ETag")
		if originalEtag != "" && !strings.HasPrefix(originalEtag, "W/") {
			w.Header().Set("ETag", "W/"+originalEtag)
		}
	}

	// write http status
	w.OriginWriter.WriteHeader(w.statusCode)

	w.headerFlushed = true
}

// FinishWriting flushes header and closed brotli writer
//
// Write() and WriteHeader() should not be called
// after FinishWriting()
func (w *writerWrapper) FinishWriting() {
	// still buffering
	if w.shouldCompress && !w.bodyBigEnough {
		w.shouldCompress = false
		w.WriteHeaderNow()
		if len(w.bodyBuffer) > 0 {
			_, _ = w.OriginWriter.Write(w.bodyBuffer)
		}
	}

	w.WriteHeaderNow()
	if w.brotliWriter != nil {
		w.PutBrotliWriter(w.brotliWriter)
		w.brotliWriter = nil
	}
}

// Flush implements the http.Flusher interface.
func (w *writerWrapper) Flush() {
	w.FinishWriting()

	if flusher, ok := w.OriginWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
