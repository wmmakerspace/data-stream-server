package main

import (
    "flag"
    "log"
    "net/http"
    "text/template"

    "github.com/gorilla/websocket"
)

var addr = flag.String("addr", "127.0.0.1:8080", "http service address")

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func testPage(w http.ResponseWriter, r *http.Request) {
    var pageTemplate = template.Must(template.ParseFiles("index.html"))
    pageTemplate.Execute(w, r.Host)
}

func videoInHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    log.Println("here")
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }
    go func() {
        for {
            _, p, err := ws.ReadMessage()
            if err != nil {
                log.Println(err)
                return
            }
            log.Println(p)
        }
    }()
}

func main() {
    http.HandleFunc("/test", testPage)
    http.HandleFunc("/", videoInHandler)
    log.Fatal(http.ListenAndServe(*addr, nil))
}
