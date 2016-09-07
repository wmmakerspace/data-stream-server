# Testing

This directly includes useful scripts for developing and testing the library.

- `source.go`: streams data into the example service

- `ingest.go`: injests data from the example service

### Running

First start the example service:

```sh
go run example/main.go
```

Then start the data source:

```sh
go run test/source.go
```

Then start the client:

```sh
go run test/ingest.go
```

