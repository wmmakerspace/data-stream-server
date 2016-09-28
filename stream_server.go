package streamserver

import (
    "fmt"
    "log"
    "sync"
    "strings"
    "strconv"
    "net/http"
    "encoding/json"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

var endpoint string
var header []byte
var sourceId = 1;
var sourceIdMutex = &sync.Mutex{}
var sources = make(map[string]map[*websocket.Conn]bool)

func DataInHandler(w http.ResponseWriter, r *http.Request) {
    sourceIdMutex.Lock()
    idStr := strconv.Itoa(sourceId);
    sourceId++
    sourceIdMutex.Unlock()
    sources[idStr] = make(map[*websocket.Conn]bool)

    log.Println("new data source: " +  idStr)

    ws, err := upgrader.Upgrade(w, r, nil)

    http.HandleFunc(endpoint + "/out/" + idStr, DataOutHandler)

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

    if header != nil {
        if err := ws.WriteMessage(websocket.BinaryMessage, header); err != nil {
            log.Println(err)
            return
        }
    }

    sources[sourceId][ws] = true
    log.Println("new client connected to: " + sourceId)

}

func ListStreams(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*") // CORS

    sourceIds := make([]string, 0, len(sources))
    for k := range sources {
        sourceIds = append(sourceIds, k)
    }
    encoder := json.NewEncoder(w)
    encoder.Encode(sourceIds)
}

func Start(e string, h []byte) {
    endpoint = e
    header = h
    http.HandleFunc(endpoint + "/list", ListStreams)
    http.HandleFunc(endpoint + "/in", DataInHandler)
}
