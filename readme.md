# postgres-api : a very simple example of a golang API on top of a Postgres table

[![GoDoc](https://godoc.org/github.com/dstroot/postgres-api?status.svg)](https://godoc.org/github.com/dstroot/postgres-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/dstroot/postgres-api)](https://goreportcard.com/report/github.com/dstroot/postgres-api)

## Overview 

This API uses:
* Julien Schmidt's httprouter [package](https://github.com/julienschmidt/httprouter).  In contrast to the default mux of Go's net/http package, this router supports variables in the routing pattern and matches against the request method. The router is optimized for high performance and a small memory footprint.
* Negroni middleware [package](https://github.com/urfave/negroni).  Negroni is an idiomatic approach to web middleware in Go. It is tiny, non-intrusive, and encourages use of net/http Handlers.
* Godotenv [package](https://github.com/joho/godotenv) loads env vars from a .env file. Storing configuration in the environment is one of the tenets of a twelve-factor app. But it is not always practical to set environment variables on development machines or continuous integration servers where multiple projects are run. Godotenv load variables from a .env file into ENV when the environment is bootstrapped.
* Envdecode [package](https://github.com/joeshaw/envdecode). Envdecode uses struct tags to map environment variables to fields, allowing you you use any names you want for environment variables. In this way you load the environment variables into a config struct once and can then use them throughout your program.
* Errors [package](https://github.com/pkg/errors).  The errors package allows you to add context to the failure path of your code in a way that does not destroy the original value of the error.

This program is written for go 1.8 and takes advantage of the ability to drain connections and do a graceful shutdown.  


## Install

```
❯ go get github.com/dstroot/postgres-api
❯ cd $GOPATH/src/github.com/dstroot/postgres-api
❯ dep ensure -update
❯ go test -v
❯ go build && ./postgres-api
```

## License

MIT.

### Operating

You need a postgres database and a database created to use.  Set the .env parameters to point to your postgres installation.  Run `go test` to initialize the table. After that you should be able to build and run the program.

Run psql cli:

`$ docker run -it --rm --link postgres:postgres postgres psql -h postgres -U postgres`


Here's how you create a database in Postgres:

```
psql (9.6.2)
Type "help" for help.

postgres=# CREATE DATABASE products;
CREATE DATABASE
postgres=# \q
```

How to build the program:

```
$ go build -ldflags "-X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.commit=`git rev-parse HEAD` -w -s" && ./postgres-api
```

Or, just (without build flags)

```
$ go build && ./postgres-api
```

### References

* https://tylerchr.blog/golang-18-whats-coming/
* https://dave.cheney.net/2016/06/12/stack-traces-and-the-errors-package
