package streamserver

import (
    "log"
    "strings"
    "strconv"
    "net/http"
    "encoding/json"

    "github.com/gorilla/websocket"
)

var sources = make(map[int]*DataSource)
// TODO: protect this, not thread safe
var sourceId = 0
var endpoint = ""

// allow CORS
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func Start(ep string) {
    endpoint = ep
    http.HandleFunc(endpoint + "/in", DataInHandler)
    http.HandleFunc(endpoint + "/list", ListStreams)
}

type Client struct {
    ws *websocket.Conn
    data chan []byte
    source *DataSource
}

func (c *Client) run() {
    if err := c.ws.WriteMessage(websocket.BinaryMessage, header); err != nil {
        log.Println(err)
        return
    }
    c.source.register <- c
    defer func() {
        c.source.unregister <- c
    }()
    for {
        d := <- c.data
        if err := c.ws.WriteMessage(websocket.BinaryMessage, d); err != nil {
            log.Println(err)
            log.Println("CLOSED: " + c.ws.RemoteAddr().String())
            c.ws.Close()
            return
        }
    }
}

type DataSource struct {
    data chan []byte
    clients map[*Client]bool
    register chan *Client
    unregister chan *Client
    id int
}

func (v *DataSource) run() {
    for {
        select {
        case client := <-v.register:
            v.clients[client] = true
        case client := <-v.unregister:
            delete(v.clients, client)
            close(client.data)
        case data := <- v.data:
            for client := range v.clients {
                client.data <- data
            }
        }
    }
}

func newDataSource() *DataSource {
    sourceId++
    return &DataSource{
        data: make(chan []byte),
        clients: make(map[*Client]bool),
        register: make(chan *Client),
        unregister: make(chan *Client),
        id: sourceId,
    }
}

func DataInHandler(w http.ResponseWriter, r *http.Request) {
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }
    ds := newDataSource()
    // TODO: remove handler when the source goes away
    http.HandleFunc(endpoint + "/out/" + strconv.Itoa(ds.id), DataOutHandler)
    sources[ds.id] = ds
    log.Println("new data source connected: " + strconv.Itoa(ds.id))
    go ds.run()
    go func() {
        for {
            _, p, err := ws.ReadMessage()
            if err != nil {
                log.Println(err)
                delete(sources, ds.id)
                return
            }
            ds.data <- p
        }
    }()
}

func DataOutHandler(w http.ResponseWriter, r *http.Request) {
    url := strings.Split(r.URL.String(), "/")
    reqSourceId, err := strconv.Atoi(url[len(url) - 1])
    if err != nil {
        log.Println(err)
        return
    }
    log.Println("new client connected to: " + strconv.Itoa(reqSourceId))
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }
    client := Client{
        ws: ws,
        data: make(chan []byte),
        source: sources[reqSourceId],
    }
    client.run()
}

func ListStreams(w http.ResponseWriter, r *http.Request) {
    sourceIds := make([]int, 0, len(sources))
    for k := range sources {
        sourceIds = append(sourceIds, k)
    }
    encoder := json.NewEncoder(w)
    encoder.Encode(sourceIds)
}
