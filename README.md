# failure

[![CircleCI](https://circleci.com/gh/morikuni/failure/tree/master.svg?style=shield)](https://circleci.com/gh/morikuni/failure/tree/master)
[![GoDoc](https://godoc.org/github.com/morikuni/failure?status.svg)](https://godoc.org/github.com/morikuni/failure)
[![Go Report Card](https://goreportcard.com/badge/github.com/morikuni/failure)](https://goreportcard.com/report/github.com/morikuni/failure)
[![codecov](https://codecov.io/gh/morikuni/failure/branch/master/graph/badge.svg)](https://codecov.io/gh/morikuni/failure)

Package failure provides an error represented as error code and extensible error interface with wrappers.

Use error code instead of error type.

```go
var NotFound failure.StringCode = "NotFound"

err := failure.New(NotFound)

if failure.Is(err, NotFound) { // true
	r.WriteHeader(http.StatusNotFound)
}
```

Wrap errors.

```go
type Wrapper interface {
	WrapError(err error) error
}

err := failure.Wrap(err, MarkTemporary())
```

Unwrap errors with Iterator.

```go
type Unwrapper interface {
	UnwrapError() error
}

i := failure.NewIterator(err)
for i.Next() { // unwrap error
	err := i.Error()
	if e, ok := err.(Temporary); ok {
		return e.IsTemporary()
	}
}
```

## Example

```go
package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/morikuni/failure"
)

// error codes for your application.
const (
	NotFound  failure.StringCode = "not_found"
	Forbidden failure.StringCode = "forbidden"
)

func GetACL(projectID, userID string) (acl interface{}, e error) {
	notFound := true
	if notFound {
		return nil, failure.New(NotFound,
			failure.Debug{"project_id": projectID, "user_id": userID},
		)
	}
	err := errors.New("error")
	if err != nil {
		return nil, failure.Wrap(err)
	}
	return nil, nil
}

func GetProject(projectID, userID string) (project interface{}, e error) {
	_, err := GetACL(projectID, userID)
	if err != nil {
		if failure.Is(err, NotFound) {
			return nil, failure.Translate(err, Forbidden,
				failure.Message("You have no grant to access the project."),
				failure.Debug{"additional_info": "hello"},
			)
		}
		return nil, failure.Wrap(err)
	}
	return nil, nil
}

func Handler(w http.ResponseWriter, r *http.Request) {
	_, err := GetProject(r.FormValue("projectID"), r.FormValue("userID"))
	if err != nil {
		HandleError(w, err)
		return
	}
}

func getHTTPStatus(err error) int {
	c, ok := failure.CodeOf(err)
	if !ok {
		return http.StatusInternalServerError
	}
	switch c {
	case NotFound:
		return http.StatusNotFound
	case Forbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func getMessage(err error) string {
	msg, ok := failure.MessageOf(err)
	if ok {
		return msg
	}
	return "Error"
}

func HandleError(w http.ResponseWriter, err error) {
	w.WriteHeader(getHTTPStatus(err))
	io.WriteString(w, getMessage(err))

	fmt.Println(err)
	// main.GetProject: code(forbidden): main.GetACL: code(not_found)
	fmt.Printf("%+v\n", err)
	// [main.GetProject] /go/src/github.com/morikuni/failure/example/main.go:36
	//     additional_info = hello
	//     message("You have no grant to access the project.")
	//     code(forbidden)
	// [main.GetACL] /go/src/github.com/morikuni/failure/example/main.go:21
	//     project_id =
	//     user_id =
	//     code(not_found)
	// [CallStack]
	//     [main.GetACL] /go/src/github.com/morikuni/failure/example/main.go:21
	//     [main.GetProject] /go/src/github.com/morikuni/failure/example/main.go:33
	//     [main.Handler] /go/src/github.com/morikuni/failure/example/main.go:47
	//     [http.HandlerFunc.ServeHTTP] /usr/local/go/src/net/http/server.go:1964
	//     [http.(*ServeMux).ServeHTTP] /usr/local/go/src/net/http/server.go:2361
	//     [http.serverHandler.ServeHTTP] /usr/local/go/src/net/http/server.go:2741
	//     [http.(*conn).serve] /usr/local/go/src/net/http/server.go:1847
	//     [runtime.goexit] /usr/local/go/src/runtime/asm_amd64.s:1333
	fmt.Println(failure.CodeOf(err))
	// forbidden
	fmt.Println(failure.MessageOf(err))
	// You have no grant to access the project.
	fmt.Println(failure.DebugsOf(err))
	// [map[additional_info:hello] map[project_id:123 user_id:456]]
	fmt.Println(failure.CallStackOf(err))
	// main.GetACL: main.GetProject: main.Handler: http.HandlerFunc.ServeHTTP: http.(*ServeMux).ServeHTTP: http.serverHandler.ServeHTTP: http.(*conn).serve: goexit
	fmt.Println(failure.CauseOf(err))
	// code(not_found)
}

func main() {
	http.HandleFunc("/", Handler)
	http.ListenAndServe(":8080", nil)
	// $ http "localhost:8080/?projectID=123&userID=456"
	// HTTP/1.1 403 Forbidden
	// Content-Length: 40
	// Content-Type: text/plain; charset=utf-8
	// Date: Tue, 19 Jun 2018 02:56:32 GMT
	//
	// You have no grant to access the project.
}
```
