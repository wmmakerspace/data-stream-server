package streamserver

import (
    "net/http"
    "github.com/gorilla/websocket"
    "fmt"
    "strings"
    "strconv"
    "sync"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

var sourceId = 1;
var sourceIdMutex = &sync.Mutex{}
var sources = make(map[string]map[*websocket.Conn]bool)

func DataInHandler(w http.ResponseWriter, r *http.Request) {
    sourceIdMutex.Lock()
    idStr := strconv.Itoa(sourceId);
    sourceId++
    sourceIdMutex.Unlock()
    sources[idStr] = make(map[*websocket.Conn]bool)

    ws, err := upgrader.Upgrade(w, r, nil)

    http.HandleFunc("/data/out/" + idStr, DataOutHandler)

    if err != nil {
        fmt.Println(err)
        return
    }
    for {
        _, p, err := ws.ReadMessage()
        if err != nil {
            delete(sources, idStr)
            fmt.Println(err)
            return
        }
        for client := range sources[idStr] {
            if err := client.WriteMessage(websocket.BinaryMessage, p); err != nil {
                delete(sources[idStr], client)
                fmt.Println(err)
            }
        }
    }
}

func DataOutHandler(w http.ResponseWriter, r *http.Request) {
    url := strings.Split(r.URL.String(), "/")
    sourceId := url[len(url) - 1]
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    sources[sourceId][ws] = true
}

func Start(endpoint string) {
    http.HandleFunc(endpoint + "/in", DataInHandler)
}
