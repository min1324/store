# store 

[![Build Status](https://travis-ci.com/min1324/store.svg?branch=main)](https://travis-ci.com/min1324/store) [![codecov](https://codecov.io/gh/min1324/store/branch/main/graph/badge.svg)](https://codecov.io/gh/min1324/store) [![Go Report Card](https://goreportcard.com/badge/github.com/min1324/store)](https://goreportcard.com/report/github.com/min1324/store) [![GoDoc](https://godoc.org/github.com/min1324/store?status.png)](https://godoc.org/github.com/min1324/store)

 store 是 **go** atomic.Value衍生出来的结构体。由于原生value不支持储存nil和不同类型数据，store修改原生value使其支持nil和不同类型数据。

## usage

Import the package:

```go
import (
	"github.com/min1324/store"
)

```

```bash
go get "github.com/min1324/store"
```

The package is now imported under the "store" namespace.

## example

```go
	// Create a new store.
	var m store.Entry

	// Stores item within store
	m.Store("foo")

```
