package brotli

import (
	"bufio"
	"io/ioutil"
	"net"
	"net/http"
	"sync"

	"github.com/andybalholm/brotli"
	"github.com/gin-gonic/gin"
)

const (
	BestSpeed          = brotli.BestSpeed
	BestCompression    = brotli.BestCompression
	DefaultCompression = brotli.DefaultCompression
	DefalutContentLen  = 1024
)

// Config is used in Handler initialization
type Config struct {
	// 压缩等级
	CompressionLevel int
	// 响应内容长度
	MinContentLength int64
	// 根据请求校验是否过滤
	RequestFilter []RequestFilter
	// 根据响应校验是否过滤
	ResponseHeaderFilter []ResponseHeaderFilter
}

// Handler implement brotli compression for gin
type Handler struct {
	compressionLevel     int
	minContentLength     int64
	requestFilter        []RequestFilter
	responseHeaderFilter []ResponseHeaderFilter
	brotliWriterPool     sync.Pool
	wrapperPool          sync.Pool
}

func NewHandler(config Config) *Handler {
	// 设置压缩等级不符合则，使用默认等级
	if config.CompressionLevel < BestSpeed || config.CompressionLevel > BestCompression {
		config.CompressionLevel = DefaultCompression
	}
	// 设置默认压缩长度限制
	if config.MinContentLength <= 0 {
		config.MinContentLength = DefalutContentLen
	}

	handler := Handler{
		compressionLevel:     config.CompressionLevel,
		minContentLength:     config.MinContentLength,
		requestFilter:        config.RequestFilter,
		responseHeaderFilter: config.ResponseHeaderFilter,
	}

	// brotli writer
	handler.brotliWriterPool.New = func() interface{} {
		return brotli.NewWriterLevel(ioutil.Discard, handler.compressionLevel)
	}
	handler.wrapperPool.New = func() interface{} {
		return newWriterWrapper(handler.responseHeaderFilter,
			handler.minContentLength,
			nil,
			handler.getBrotliWriter,
			handler.putBrotliWriter)
	}

	return &handler
}

// 默认配置
var defaultConfig = Config{
	CompressionLevel: DefaultCompression,
	MinContentLength: DefalutContentLen,
	RequestFilter: []RequestFilter{
		NewCommonRequestFilter(),
	},
	ResponseHeaderFilter: []ResponseHeaderFilter{
		DefaultContentTypeFilter(),
	},
}

// DefaultHandler 创建一个默认handler
func DefaultHandler() *Handler {
	return NewHandler(defaultConfig)
}

// getBrotliWriter 获取一个brotli writer
func (h *Handler) getBrotliWriter() *brotli.Writer {
	return h.brotliWriterPool.Get().(*brotli.Writer)
}

// putBrotliWriter 回收brotli writer
func (h *Handler) putBrotliWriter(w *brotli.Writer) {
	if w == nil {
		return
	}

	_ = w.Close()
	w.Reset(ioutil.Discard)
	h.brotliWriterPool.Put(w)
}

// getWriteWrapper 获取Wrapper
func (h *Handler) getWriteWrapper() *writerWrapper {
	return h.wrapperPool.Get().(*writerWrapper)
}

// putWriteWrapper
func (h *Handler) putWriteWrapper(w *writerWrapper) {
	if w == nil {
		return
	}

	// 回收资源
	w.FinishWriting()
	w.OriginWriter = nil
	h.wrapperPool.Put(w)
}

type ginBrotliWriter struct {
	wrapper      *writerWrapper
	originWriter gin.ResponseWriter
}

// interface verification
var _ gin.ResponseWriter = &ginBrotliWriter{}

// WriteHeaderNow implement the gin.ResponseWriter interface.
func (g *ginBrotliWriter) WriteHeaderNow() {
	g.wrapper.WriteHeaderNow()
}

// Hijack implements the http.Hijacker interface.
func (g *ginBrotliWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return g.originWriter.Hijack()
}

// CloseNotify implements the http.CloseNotify interface.
func (g *ginBrotliWriter) CloseNotify() <-chan bool {
	return g.originWriter.CloseNotify()
}

// Status implement the gin.ResponseWriter interface.
func (g *ginBrotliWriter) Status() int {
	return g.wrapper.Status()
}

// Size implement the gin.ResponseWriter interface.
func (g *ginBrotliWriter) Size() int {
	return g.wrapper.Size()
}

// Written implement the gin.ResponseWriter interface.
func (g *ginBrotliWriter) Written() bool {
	return g.wrapper.Written()
}

// Pusher implement the gin.ResponseWriter interface.
func (g *ginBrotliWriter) Pusher() http.Pusher {
	return g.originWriter.Pusher()
}

// WriteString implements the gin.ResponseWriter interface.
func (g *ginBrotliWriter) WriteString(s string) (int, error) {
	return g.wrapper.Write([]byte(s))
}

// Write implements the http.ResponseWriter interface.
func (g *ginBrotliWriter) Write(data []byte) (int, error) {
	return g.wrapper.Write(data)
}

// WriteHeader implements the http.ResponseWriter interface.
func (g *ginBrotliWriter) WriteHeader(code int) {
	g.wrapper.WriteHeader(code)
}

// Header implements the http.ResponseWriter interface.
func (g *ginBrotliWriter) Header() http.Header {
	return g.wrapper.Header()
}

// Flush implements the http.Flusher interface.
func (g *ginBrotliWriter) Flush() {
	g.wrapper.Flush()
}

// Gin implement gin's middleware
func (h *Handler) Gin(ctx *gin.Context) {
	var shouldCompress = true

	// 根据请求信息校验是否进行压缩
	for _, filter := range h.requestFilter {
		shouldCompress = filter.ShouldCompress(ctx.Request)
		if !shouldCompress {
			break
		}
	}

	if shouldCompress {
		wrapper := h.getWriteWrapper()
		wrapper.Reset(ctx.Writer)

		originWriter := ctx.Writer
		ctx.Writer = &ginBrotliWriter{
			originWriter: ctx.Writer,
			wrapper:      wrapper,
		}
		defer func() {
			// 资源回收
			h.putWriteWrapper(wrapper)
			ctx.Writer = originWriter
		}()
	}

	ctx.Next()
}
