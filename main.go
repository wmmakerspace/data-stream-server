package main

import (
    "fmt"
    "log"
    "flag"
    "net/http"

    "github.com/gorilla/websocket"
)

var addr = flag.String("addr", "127.0.0.1:8080", "http service address")

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

var video = make(chan []byte)

func videoInHandler(w http.ResponseWriter, r *http.Request) {
    // enable CORS
    w.Header().Set("Access-Control-Allow-Origin", "*")

    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }

    log.Println("new video source connected: " + ws.RemoteAddr().String())

    go func() {
        for {
            _, p, err := ws.ReadMessage()
            if err != nil {
                log.Println(err)
                return
            }
            video <- p
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

    log.Println("new video client: " + ws.RemoteAddr().String())

    // https://github.com/phoboslab/jsmpeg/blob/master/stream-server.js#L23-L27
    header := []byte{0x6a, 0x73, 0x6d, 0x70, 0x01, 0x40, 0x00, 0xf0}
    if err := ws.WriteMessage(websocket.BinaryMessage, header); err != nil {
        log.Println(err)
        return
    }

    go func() {
        for {
            if err = ws.WriteMessage(websocket.BinaryMessage, <-video); err != nil {
                log.Println(err)
                log.Println("CLOSED: " + ws.RemoteAddr().String())
                return
            }
        }
    }()
}

func main() {
    http.HandleFunc("/video/in", videoInHandler)
    http.HandleFunc("/video/out", videoOutHandler)
    fmt.Println("Server listening on port :8080")
    fmt.Println("------------------------------")
    log.Fatal(http.ListenAndServe(*addr, nil))
}
