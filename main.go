package main

import (
    "flag"
    "log"
    "time"
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

func videoOutHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }
    go func() {
        for {
            if err = ws.WriteMessage(websocket.TextMessage, []byte{1, 2, 3}); err != nil {
                return
            }
            time.Sleep(time.Millisecond * 3000)
        }
    }()
}

func main() {
    http.HandleFunc("/test", testPage)
    http.HandleFunc("/video/in", videoInHandler)
    http.HandleFunc("/video/out", videoOutHandler)
    log.Fatal(http.ListenAndServe(*addr, nil))
}
