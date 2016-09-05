package main

import (
    "fmt"
    "log"
    "flag"
    "net/http"

    "github.com/gorilla/websocket"
)

var addr = flag.String("addr", "127.0.0.1:8080", "http service address")
var sources = make(map[*VideoSource]bool)

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

func videoInHandler(w http.ResponseWriter, r *http.Request) {
    // enable CORS
    w.Header().Set("Access-Control-Allow-Origin", "*")
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }
    vs := VideoSource{
        video: make(chan []byte),
        clients: make(map[*Client]bool),
        register: make(chan *Client),
        unregister: make(chan *Client),
    }
    sources[&vs] = true
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
    // enable CORS
    w.Header().Set("Access-Control-Allow-Origin", "*")
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }
    // TODO CHANGE THIS!!!
    var source *VideoSource
    for k := range sources {
        source = k
    }
    client := Client{
        ws: ws,
        video: make(chan []byte),
        source: source,
    }
    go client.run()
}

func main() {
    http.HandleFunc("/video/in", videoInHandler)
    http.HandleFunc("/video/out", videoOutHandler)

    fmt.Println("Server listening on port :8080")
    fmt.Println("------------------------------")
    log.Fatal(http.ListenAndServe(*addr, nil))
}
