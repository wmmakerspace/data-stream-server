# Video

Streaming data server


### Usage

See `example/server.go` for an example server

```go
import "github.com/wmmakerspace/data-stream-server"

// the argument to the Start function is the endpoint that 
// exposes the data stream server
streamserver.Start("/data")

http.ListenAndServe(":8080", nil)
```

The library exposes 3 endpoints under the user defined endpoint:

- `/in`: data sources stream data to this endpoint

- `/out/<source_id>`: clients stream data from this enpoint

- `/list`: returns a JSON array or currently available source ids

- `/out/<source_id>/metadata`: returns whatever metadata a data source supplies


### Developing

See [`/test`](/test) for helper scripts