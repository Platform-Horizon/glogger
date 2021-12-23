<div align="center">

# Glogger

[![Build Status](https://github.com/riccardotzr/glogger/workflows/build/badge.svg)](https://github.com/riccardotzr/gconfig/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/riccardotzr/glogger)](https://goreportcard.com/report/github.com/riccardotzr/glogger)
[![Go Reference](https://pkg.go.dev/badge/github.com/riccardotzr/glogger.svg)](https://pkg.go.dev/github.com/riccardotzr/glogger)

</div>

Glogger is a go logging library. It uses [logrus](https://github.com/sirupsen/logrus) library and implements a middleware to be used with [Gorilla Mux](https://github.com/gorilla/mux).

## Install

```ssh
go get -u github.com/riccardotzr/glogger
```

## Usage

### Logger initialization
```go
log, err := glogger.Init(glogger.InitOptions{Level: "info"})

if err != nil {
    panic(err.Error())
}

```

### Middleware initialization
```go
r := mux.NewRouter()
r.Use(glogger.LoggingMiddleware(log))
```

and to retrieve logger injected in request context:

```go
func (w http.ResponseWriter, r *http.Request) {
    logger := glogger.Get(r.Context())
    logger.Info("My log message")
}
```

### Logging Error Message

To log error message using default field

```go
func (w http.ResponseWriter, r *http.Request) {
    logger := glogger.Get(r.Context())
    
    _, err := fn()

    if err != nil {
        logger.WithError(err).Error("My error message")
    }
}
```

### Logging Custom Fields
To log error message using default field

```go
func (w http.ResponseWriter, r *http.Request) {
    logger := glogger.Get(r.Context())
    logger.WithFields(&logrus.Fields{
        "key": "my_key",
    }).Info("Log with custom fields")
}
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE.md](LICENSE.md)
file for details