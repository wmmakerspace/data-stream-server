package main

import (
    "fmt"
    "log"
    "flag"
    "strings"
    "strconv"
    "net/http"

    "github.com/gorilla/websocket"
)

var addr = flag.String("addr", "127.0.0.1:8080", "http service address")
var sources = make(map[int]*VideoSource)
// TODO: protect this
var sourceId = 0

// allow CORS
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

type Client struct {
    ws *websocket.Conn
    video chan []byte
    source *VideoSource
}

func (c *Client) run() {
    // https://github.com/phoboslab/jsmpeg/blob/master/stream-server.js#L23-L27
    header := []byte{0x6a, 0x73, 0x6d, 0x70, 0x01, 0x40, 0x00, 0xf0}
    if err := c.ws.WriteMessage(websocket.BinaryMessage, header); err != nil {
        log.Println(err)
        return
    }
    c.source.register <- c
    for {
        if err := c.ws.WriteMessage(websocket.BinaryMessage, <-c.video); err != nil {
            log.Println(err)
            log.Println("CLOSED: " + c.ws.RemoteAddr().String())
            c.source.unregister <- c
            return
        }
    }
}

type VideoSource struct {
    video chan []byte
    clients map[*Client]bool
    register chan *Client
    unregister chan *Client
    id int
}

func (v *VideoSource) run() {
    for {
        select {
        case client := <- v.register:
            v.clients[client] = true
        case client := <- v.unregister:
            delete(v.clients, client)
            close(client.video)
        case video := <- v.video:
            for client := range v.clients {
                client.video <- video
            }
        }
    }
}

func newVideoSource() *VideoSource {
    sourceId++
    return &VideoSource{
        video: make(chan []byte),
        clients: make(map[*Client]bool),
        register: make(chan *Client),
        unregister: make(chan *Client),
        id: sourceId,
    }
}

func videoInHandler(w http.ResponseWriter, r *http.Request) {
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }
    vs := newVideoSource()
    http.HandleFunc("/video/out/" + strconv.Itoa(vs.id), videoOutHandler)
    sources[vs.id] = vs
    go vs.run()
    go func() {
        for {
            _, p, err := ws.ReadMessage()
            if err != nil {
                log.Println(err)
                return
            }
            vs.video <- p
        }
    }()
}

func videoOutHandler(w http.ResponseWriter, r *http.Request) {
    url := strings.Split(r.URL.String(), "/")
    reqSourceId, err := strconv.Atoi(url[len(url) - 1])
    if err != nil {
        log.Println(err)
        return
    }
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }
    client := Client{
        ws: ws,
        video: make(chan []byte),
        source: sources[reqSourceId],
    }
    go client.run()
}

func main() {
    http.HandleFunc("/video/in", videoInHandler)

    fmt.Println("Server listening on port :8080")
    fmt.Println("------------------------------")
    log.Fatal(http.ListenAndServe(*addr, nil))
}
