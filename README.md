# Brotli Middleware for Gin

## 功能

针对http输出数据进行压缩

## 使用案例

### 默认

```golang
handler.Use(brotli.DefaultHandler().Gin)
```

### 自定义

```golang
 handler.Use(brotli.NewHandler(brotli.Config{
  CompressionLevel: brotli.DefaultCompression,
  MinContentLength: brotli.DefalutContentLen,
  RequestFilter: []brotli.RequestFilter{
   brotli.NewCommonRequestFilter(),
   brotli.NewRequestApiFilter([]string{
    "/blog/cache/http",
    "/blog/detail/index",
   }),
  },
  ResponseHeaderFilter: []brotli.ResponseHeaderFilter{
   brotli.DefaultContentTypeFilter(),
  },
 }).Gin)
```

## 测试

### 压测速率

```brotli
bash-4.3# go test -v  -run="GinWithLevelsHandler"
=== RUN   TestGinWithLevelsHandler
=== RUN   TestGinWithLevelsHandler/level_0
=== RUN   TestGinWithLevelsHandler/level_1
=== RUN   TestGinWithLevelsHandler/level_2
=== RUN   TestGinWithLevelsHandler/level_3
=== RUN   TestGinWithLevelsHandler/level_4
=== RUN   TestGinWithLevelsHandler/level_5
=== RUN   TestGinWithLevelsHandler/level_6
=== RUN   TestGinWithLevelsHandler/level_7
=== RUN   TestGinWithLevelsHandler/level_8
=== RUN   TestGinWithLevelsHandler/level_9
=== RUN   TestGinWithLevelsHandler/level_10
=== RUN   TestGinWithLevelsHandler/level_11
--- PASS: TestGinWithLevelsHandler (0.06s)
    --- PASS: TestGinWithLevelsHandler/level_0 (0.00s)
        handler_test.go:215: level_0: compressed 4267 => 1456 ratio=>0.66
    --- PASS: TestGinWithLevelsHandler/level_1 (0.00s)
        handler_test.go:215: level_1: compressed 4267 => 1419 ratio=>0.67
    --- PASS: TestGinWithLevelsHandler/level_2 (0.00s)
        handler_test.go:215: level_2: compressed 4267 => 1410 ratio=>0.67
    --- PASS: TestGinWithLevelsHandler/level_3 (0.00s)
        handler_test.go:215: level_3: compressed 4267 => 1340 ratio=>0.69
    --- PASS: TestGinWithLevelsHandler/level_4 (0.00s)
        handler_test.go:215: level_4: compressed 4267 => 1257 ratio=>0.71
    --- PASS: TestGinWithLevelsHandler/level_5 (0.00s)
        handler_test.go:215: level_5: compressed 4267 => 1187 ratio=>0.72
    --- PASS: TestGinWithLevelsHandler/level_6 (0.01s)
        handler_test.go:215: level_6: compressed 4267 => 1187 ratio=>0.72
    --- PASS: TestGinWithLevelsHandler/level_7 (0.00s)
        handler_test.go:215: level_7: compressed 4267 => 1186 ratio=>0.72
    --- PASS: TestGinWithLevelsHandler/level_8 (0.01s)
        handler_test.go:215: level_8: compressed 4267 => 1186 ratio=>0.72
    --- PASS: TestGinWithLevelsHandler/level_9 (0.01s)
        handler_test.go:215: level_9: compressed 4267 => 1185 ratio=>0.72
    --- PASS: TestGinWithLevelsHandler/level_10 (0.01s)
        handler_test.go:215: level_10: compressed 4268 => 1049 ratio=>0.75
    --- PASS: TestGinWithLevelsHandler/level_11 (0.02s)
        handler_test.go:215: level_11: compressed 4268 => 1042 ratio=>0.76
PASS
ok      github.com/CodeLineage/brotli   0.076s
bash-4.3#
```

### 性能测试

```brotli
bash-4.3# go test -benchmem -bench .
goos: linux
goarch: amd64
pkg: github.com/CodeLineage/brotli
BenchmarkGin_SmallPayload-2                              2917842               435 ns/op              96 B/op          3 allocs/op
BenchmarkGinWithDefaultHandler_SmallPayload-2            1271218               997 ns/op             128 B/op          4 allocs/op
BenchmarkGin_BigPayload-2                                2912677               424 ns/op              96 B/op          3 allocs/op
BenchmarkGinWithDefaultHandler_BigPayload-2                 3063            405511 ns/op             882 B/op          6 allocs/op
PASS
ok      github.com/CodeLineage/brotli   7.122s
```

## 依赖

github.com/andybalholm/brotli v1.0.1

## 参考

[https://github.com/nanmu42/gzip](https://github.com/nanmu42/gzip)
