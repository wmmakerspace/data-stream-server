# Testing

This directly includes useful scripts for developing and testing the library.

- `server.go`: a stream server with the following usage

```sh
go run server.go \
    -port=<PORT_FOR_SERVICE> \ # default: '8080'
    -enpoint=<ENDPOINT_OF_SERVICE> \ # default: 'data'
    -video-header # boolean indicating to use 8 byte 
                  # video header required for jsmpeg
```

- `source.go`: streams data into the example service; usage:

```sh
go run source.go \
    -host=<SERVER_HOST> \ # default: 'localhost'
    -port=<SERVER_PORT> \ # default: '8080'
    -endpoint=<SERVICE_ENDPOINT> \ # default: 'data'
```

- `ingest.go`: ingests data from the example service

### Running

First start the example service:

```sh
go run example/server.go
```

Then start the data source:

```sh
go run test/source.go
```

Then start the client:

```sh
go run test/ingest.go
```

