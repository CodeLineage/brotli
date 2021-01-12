package brotli

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	smallPayload = []byte("this is a message")
	bigPayload   = []byte(`{"err_no":0,"err_msg":"success","data":[{"article_id":"6910032983059775502","article_info":{"article_id":"6910032983059775502","user_id":"641770519800781","category_id":"6809637769959178254","tag_ids":[6809640364677267000],"visible_level":0,"link_url":"","cover_image":"","is_gfw":0,"title":"Golang—literal copies lock value from gzPool: sync.Pool contains sync.noCopy","brief_content":"","is_english":0,"is_original":1,"user_index":0,"original_type":0,"original_author":"","content":"","ctime":"1608867519","mtime":"1608877626","rtime":"1608877626","draft_id":"6910032808006123533","view_count":27,"collect_count":0,"digg_count":0,"comment_count":0,"hot_index":1,"is_hot":0,"rank_index":0.00011862,"status":2,"verify_status":1,"audit_status":2,"mark_content":""},"author_user_info":{"user_id":"641770519800781","user_name":"欧阳俊","company":"互联网公司","job_title":"R \\u0026 D(Gopher|PHPer)","avatar_large":"https://sf6-ttcdn-tos.pstatp.com/img/user-avatar/828afd4d8e0b520b854be4a813c2ae72~300x300.image","level":1,"description":"互联网的搬运工，喜欢骑车、读书、跑步 | 人生只要用心，无论输赢都是精彩","followee_count":15,"follower_count":6,"post_article_count":60,"digg_article_count":76,"got_digg_count":21,"got_view_count":4959,"post_shortmsg_count":0,"digg_shortmsg_count":0,"isfollowed":false,"favorable_author":0,"power":70,"study_point":0,"university":{"university_id":"0","name":"","logo":""},"major":{"major_id":"0","parent_id":"0","name":""},"student_status":0,"select_event_count":0,"select_online_course_count":0,"identity":0},"category":{"category_id":"6809637769959178254","category_name":"后端","category_url":"backend","rank":1,"ctime":1457483880,"mtime":1432503193,"show_type":3},"tags":[{"id":2546494,"tag_id":"6809640364677267469","tag_name":"Go","color":"#64D7E3","icon":"https://lc-gold-cdn.xitu.io/1aae38ab22d12b654cfa.png","back_ground":"","show_navi":0,"tag_alias":"","post_article_count":6383,"concern_user_count":79309}],"user_interact":{"id":6910032983059775000,"omitempty":2,"user_id":641770519800781,"is_digg":false,"is_follow":false,"is_collect":false}},{"article_id":"6909475495650426894","article_info":{"article_id":"6909475495650426894","user_id":"641770519800781","category_id":"6809637769959178254","tag_ids":[6809640364677267000],"visible_level":0,"link_url":"","cover_image":"","is_gfw":0,"title":"Golang—invalid operation: m[1] (type *string does not support indexing)","brief_content":"","is_english":0,"is_original":1,"user_index":0,"original_type":0,"original_author":"","content":"","ctime":"1608737857","mtime":"1608789198","rtime":"1608789198","draft_id":"6909470897296375821","view_count":37,"collect_count":0,"digg_count":0,"comment_count":0,"hot_index":1,"is_hot":0,"rank_index":0.0001088,"status":2,"verify_status":1,"audit_status":2,"mark_content":""},"author_user_info":{"user_id":"641770519800781","user_name":"欧阳俊","company":"互联网公司","job_title":"R \\u0026 D(Gopher|PHPer)","avatar_large":"https://sf6-ttcdn-tos.pstatp.com/img/user-avatar/828afd4d8e0b520b854be4a813c2ae72~300x300.image","level":1,"description":"互联网的搬运工，喜欢骑车、读书、跑步 | 人生只要用心，无论输赢都是精彩","followee_count":15,"follower_count":6,"post_article_count":60,"digg_article_count":76,"got_digg_count":21,"got_view_count":4959,"post_shortmsg_count":0,"digg_shortmsg_count":0,"isfollowed":false,"favorable_author":0,"power":70,"study_point":0,"university":{"university_id":"0","name":"","logo":""},"major":{"major_id":"0","parent_id":"0","name":""},"student_status":0,"select_event_count":0,"select_online_course_count":0,"identity":0},"category":{"category_id":"6809637769959178254","category_name":"后端","category_url":"backend","rank":1,"ctime":1457483880,"mtime":1432503193,"show_type":3},"tags":[{"id":2546494,"tag_id":"6809640364677267469","tag_name":"Go","color":"#64D7E3","icon":"https://lc-gold-cdn.xitu.io/1aae38ab22d12b654cfa.png","back_ground":"","show_navi":0,"tag_alias":"","post_article_count":6383,"concern_user_count":79309}],"user_interact":{"id":6909475495650427000,"omitempty":2,"user_id":641770519800781,"is_digg":false,"is_follow":false,"is_collect":false}}],"cursor":"10","count":60,"has_more":true}`)
)

func TestNewHandler(t *testing.T) {
	assert.NotPanics(t, func() {
		NewHandler(Config{
			CompressionLevel: 5,
			MinContentLength: 100,
		})
	})

	assert.NotPanics(t, func() {
		NewHandler(Config{
			CompressionLevel: -4,
			MinContentLength: 100,
		})
	})

	assert.NotPanics(t, func() {
		NewHandler(Config{
			CompressionLevel: 10,
			MinContentLength: 100,
		})
	})

	assert.NotPanics(t, func() {
		NewHandler(Config{
			CompressionLevel: 5,
			MinContentLength: 0,
		})
	})

	assert.NotPanics(t, func() {
		NewHandler(Config{
			CompressionLevel: 5,
			MinContentLength: -1,
		})
	})
}

func TestFunc(t *testing.T) {
	t.Log(callerName(1))
	t.Log(callerName(0))
}

// callerName gives the function name (qualified with a package path)
// for the caller after skip frames (where 0 means the current function).
func callerName(skip int) string {
	// Make room for the skip PC.
	var pc [1]uintptr
	n := runtime.Callers(skip+2, pc[:]) // skip + runtime.Callers + callerName
	if n == 0 {
		panic("testing: zero callers found")
	}
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.Function
}

const handlerTestSize = 256

func newGinInstance(payload []byte, middleware ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	g := gin.New()
	g.HandleMethodNotAllowed = true
	g.Use(middleware...)

	g.POST("/", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/plain; charset=utf8", payload)
	})

	return g
}

func newEchoGinInstance(payload []byte, middleware ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	g := gin.New()
	g.Use(middleware...)

	g.POST("/", func(ctx *gin.Context) {
		var buf bytes.Buffer

		_, _ = io.Copy(&buf, ctx.Request.Body)
		_, _ = buf.Write(payload)

		ctx.Data(http.StatusOK, "text/plain; charset=utf8", buf.Bytes())
	})

	return g
}

type NopWriter struct {
	header http.Header
}

func NewNopWriter() *NopWriter {
	return &NopWriter{
		header: make(http.Header),
	}
}

func (n *NopWriter) Header() http.Header {
	return n.header
}

func (n *NopWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (n *NopWriter) WriteHeader(_ int) {
	// relax
}

func TestGinHandler(t *testing.T) {
	var (
		// DefaultHandler().Gin
		g = newGinInstance(bigPayload)
		r = httptest.NewRequest(http.MethodPost, "/", nil)
		w = NewNopWriter()
	)

	r.Header.Set("Accept-Encoding", "br")

	g.ServeHTTP(w, r)

	assert.Empty(t, w.Header().Get("Content-Encoding"))
}

func TestGinWithDefaultHandler(t *testing.T) {
	var (
		g = newEchoGinInstance(bigPayload, DefaultHandler().Gin)
	)

	for i := 0; i < handlerTestSize; i++ {
		var seq = strconv.Itoa(i)
		t.Run(seq, func(t *testing.T) {
			t.Parallel()

			var (
				w = httptest.NewRecorder()
				r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(seq))
			)

			r.Header.Set("Accept-Encoding", "br")
			g.ServeHTTP(w, r)

			result := w.Result()
			require.EqualValues(t, http.StatusOK, result.StatusCode)
			require.Equal(t, "br", result.Header.Get("Content-Encoding"))

			reader := brotli.NewReader(result.Body)
			body, err := ioutil.ReadAll(reader)
			require.NoError(t, err)
			require.True(t, bytes.HasPrefix(body, []byte(seq)))
		})
	}
}

func TestGinWithLevelsHandler(t *testing.T) {
	// 测试不同等级
	for i := BestSpeed; i <= BestCompression; i++ {
		var seq = "level_" + strconv.Itoa(i)
		i := i
		t.Run(seq, func(t *testing.T) {
			g := newEchoGinInstance(bigPayload, NewHandler(Config{
				CompressionLevel: i,
				MinContentLength: 1,
			}).Gin)

			var (
				w = httptest.NewRecorder()
				r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(seq))
			)

			r.Header.Set("Accept-Encoding", "br")
			g.ServeHTTP(w, r)

			result := w.Result()
			// http状态
			require.EqualValues(t, http.StatusOK, result.StatusCode)
			// 响应编码
			require.Equal(t, "br", result.Header.Get("Content-Encoding"))
			comp, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			// 解压验证
			reader := brotli.NewReader(bytes.NewReader(comp))
			body, err := ioutil.ReadAll(reader)
			require.NoError(t, err)
			require.True(t, bytes.HasPrefix(body, []byte(seq)))
			ratio, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", (float64(len(body))-float64(len(comp)))/float64(len(body))), 64)
			t.Logf("%s: compressed %d => %d ratio=>%.2f", seq, len(body), len(comp), ratio)
		})
	}
}

func TestGinWithDefaultHandler_404(t *testing.T) {
	var (
		g = newGinInstance(bigPayload, DefaultHandler().Gin)
		r = httptest.NewRequest(http.MethodPost, "/404", nil)
		w = httptest.NewRecorder()
	)

	r.Header.Set("Accept-Encoding", "br")

	g.ServeHTTP(w, r)

	result := w.Result()

	assert.EqualValues(t, http.StatusNotFound, result.StatusCode)
	assert.Equal(t, "404 page not found", w.Body.String())
	t.Log(w.Body.String())
}

func TestGinWithDefaultHandler_405(t *testing.T) {
	var (
		g = newGinInstance(bigPayload, DefaultHandler().Gin)
		r = httptest.NewRequest(http.MethodPatch, "/", nil)
		w = httptest.NewRecorder()
	)

	r.Header.Set("Accept-Encoding", "br")

	g.ServeHTTP(w, r)

	result := w.Result()

	assert.EqualValues(t, http.StatusMethodNotAllowed, result.StatusCode)
	assert.Equal(t, "405 method not allowed", w.Body.String())
}

// corsMiddleware allows CORS request
func corsMiddleware(ctx *gin.Context) {
	ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	ctx.Writer.Header().Set("Access-Control-Allow-Methods", "POST")

	if ctx.Request.Method == http.MethodOptions {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}

	ctx.Next()
}

func TestGinCORSMiddleware(t *testing.T) {
	var (
		g = newGinInstance(bigPayload, DefaultHandler().Gin, corsMiddleware)
		r = httptest.NewRequest(http.MethodOptions, "/", nil)
		w = httptest.NewRecorder()
	)
	r.Header.Set("Accept-Encoding", "br")
	g.ServeHTTP(w, r)
	result := w.Result()

	assert.EqualValues(t, http.StatusNoContent, result.StatusCode)
	assert.Equal(t, "*", result.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "POST", result.Header.Get("Access-Control-Allow-Methods"))
	assert.EqualValues(t, 0, w.Body.Len())
}

func TestGinCORSMiddlewareWithDummyConfig(t *testing.T) {
	var (
		g = newGinInstance(bigPayload, NewHandler(Config{
			CompressionLevel:     DefaultCompression,
			MinContentLength:     100,
			RequestFilter:        nil,
			ResponseHeaderFilter: nil,
		}).Gin, corsMiddleware)
		r = httptest.NewRequest(http.MethodOptions, "/", nil)
		w = httptest.NewRecorder()
	)
	r.Header.Set("Accept-Encoding", "br")
	g.ServeHTTP(w, r)
	result := w.Result()

	assert.EqualValues(t, http.StatusNoContent, result.StatusCode)
	assert.Equal(t, "*", result.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "POST", result.Header.Get("Access-Control-Allow-Methods"))
	assert.EqualValues(t, 0, w.Body.Len())
}

func BenchmarkGin_SmallPayload(b *testing.B) {
	var (
		g = newGinInstance(smallPayload)
		r = httptest.NewRequest(http.MethodPost, "/", nil)
		w = NewNopWriter()
	)

	r.Header.Set("Accept-Encoding", "br")

	b.ResetTimer()
	h := map[string][]string(w.header)
	for i := 0; i < b.N; i++ {
		// Delete header between calls.
		for k := range h {
			delete(h, k)
		}
		g.ServeHTTP(w, r)
	}

	b.StopTimer()
	if encoding := w.Header().Get("Content-Encoding"); encoding != "" {
		b.Fatalf("Content-Encoding is not empty, but %s", encoding)
	}
}

func BenchmarkGinWithDefaultHandler_SmallPayload(b *testing.B) {
	var (
		g = newGinInstance(smallPayload, DefaultHandler().Gin)
		r = httptest.NewRequest(http.MethodPost, "/", nil)
		w = NewNopWriter()
	)

	r.Header.Set("Accept-Encoding", "br")

	b.ResetTimer()
	h := map[string][]string(w.header)
	for i := 0; i < b.N; i++ {
		// Delete header between calls.
		for k := range h {
			delete(h, k)
		}
		g.ServeHTTP(w, r)
	}

	b.StopTimer()
	if encoding := w.Header().Get("Content-Encoding"); encoding != "" {
		b.Fatalf("Content-Encoding is not empty, but %s", encoding)
	}
}

func BenchmarkGin_BigPayload(b *testing.B) {
	var (
		g = newGinInstance(bigPayload)
		r = httptest.NewRequest(http.MethodPost, "/", nil)
		w = NewNopWriter()
	)

	r.Header.Set("Accept-Encoding", "br")

	b.ResetTimer()
	h := map[string][]string(w.header)
	for i := 0; i < b.N; i++ {
		// Delete header between calls.
		for k := range h {
			delete(h, k)
		}
		g.ServeHTTP(w, r)
	}

	b.StopTimer()
	if encoding := w.Header().Get("Content-Encoding"); encoding != "" {
		b.Fatalf("Content-Encoding is not empty, but %s", encoding)
	}
}

func BenchmarkGinWithDefaultHandler_BigPayload(b *testing.B) {
	var (
		g = newGinInstance(bigPayload, DefaultHandler().Gin)
		r = httptest.NewRequest(http.MethodPost, "/", nil)
		w = NewNopWriter()
	)

	r.Header.Set("Accept-Encoding", "br")

	b.ResetTimer()
	h := map[string][]string(w.header)
	for i := 0; i < b.N; i++ {
		// Delete header between calls.
		for k := range h {
			delete(h, k)
		}
		g.ServeHTTP(w, r)
	}

	b.StopTimer()
	if encoding := w.Header().Get("Content-Encoding"); encoding != "br" {
		b.Fatalf("Content-Encoding is not brotli, but %q", encoding)
	}
}
