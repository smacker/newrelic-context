# newrelic-context

Contains different helpers to make life easier with NewRelic and Context.

## Installation

`go get github.com/smacker/newrelic-context`

## In this package:

* `ContextWithTxn` - Set NewRelic transaction to context
* `GetTnxFromContext` - Get NewRelic transaction from context anywhere
* `NewRelicMiddleware` - Reports time in NewRelic and sets transaction in context
* `WrapHTTPClient` - Wraps client transport with newrelic RoundTripper with transaction from context
* `SetTxnToGorm` - Sets transaction from Context to gorm settings
* `AddGormCallbacks` - Adds callbacks to NewRelic, you should call SetTxnToGorm to make them work
* `RedisWrapper` - Logs gopkg.in/redis.v5 time in newrelic
* `WrapRedisClient` - Returns copy of redis client with newrelic for transaction

API documentation is available on [godoc.org](https://godoc.org/github.com/smacker/newrelic-context)

## Examples:

Use NewRelicMiddleware:

```go
package main

import (
    "github.com/smacker/newrelic-context"
    "log"
    "net/http"
)

func indexHandlerFunc(rw http.ResponseWriter, req *http.Request) {
    rw.Write([]byte("I'm an index page!"))
}

func main() {
    var handler http.Handler

    handler = http.HandlerFunc(indexHandlerFunc)
    nrmiddleware, err := nrcontext.NewMiddleware("test-app", "my-license-key")
    if err != nil {
        log.Print("Can't create newrelic middleware: ", err)
    } else {
        handler = nrmiddleware.Handler(handler)
    }

    http.ListenAndServe(":8000", handler)
}

```

Use HTTPClientWrap:

```go
func Consume(ctx context.Context, query string) {
    client := &http.Client{Timeout: 10}
    nrcontext.WrapHTTPClient(ctx, client)
    _, err := client.Get(fmt.Sprintf("https://www.google.com.vn/?q=%v", query))
    if err != nil {
        log.Println("Can't fetch google :(")
        return
    }
    log.Println("Google fetched successfully!")
}
```

Use with Gorm:

```go
var db *gorm.DB

func initDB() *gorm.DB {
    db, err := gorm.Open("sqlite3", "./foo.db")
    if err != nil {
        panic(err)
    }
    nrgorm.AddGormCallbacks(db)
    return db
}

func catalogPage(db *gorm.DB) http.Handler {
    return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
        var products []Product
        db := nrcontext.SetTxnToGorm(req.Context(), db)
        db.Find(&products)
        for i, v := range products {
            rw.Write([]byte(fmt.Sprintf("%v. %v\n", i, v.Name)))
        }
    })
}

func main() {
    db = initDB()
    defer db.Close()

    handler := catalogPage(db)
    nrmiddleware, _ := nrcontext.NewMiddleware("test-app", "my-license-key")
    handler = nrmiddleware.Handler(handler)

    http.ListenAndServe(":8000", handler)
}
```

Use with Redis:

```go
func catalogPage(redisClient *redis.Client) http.Handler {
    return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
        redisClient := WrapRedisClient(req.Context(), redisClient)

        redisClient.Set("key", "value", 0)
        // ...
        redisClient.Get("other-key")
    })
}

func main() {
    client := redis.NewClient(...)

    handler := catalogPage(db)
    nrmiddleware, _ := nrcontext.NewMiddleware("test-app", "my-license-key")
    handler = nrmiddleware.Handler(handler)

    http.ListenAndServe(":8000", handler)
}

```
