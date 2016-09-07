# Video

Streaming data server


### Usage

See `example/main.go` for an example

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

- `list`: returns a JSON array or currently available source ids


### Developing

See `[/test](/test)` for helper scripts