package streamserver

import (
    "log"
    "sync"
    "strings"
    "strconv"
    "net/http"
    "encoding/json"

    "github.com/gorilla/websocket"
)

var HEADER_DELIMITER byte = '|'

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

var endpoint string
var header []byte // the first message sent to every new client before the data stream
var sourceId = 1
var sourceIdMutex = &sync.Mutex{}

var sources = make(map[string]map[*websocket.Conn]bool)
var metadata = make(map[string][]byte)

/**
 * a websocket that receives a data stream in from sources
 */
func DataInHandler(w http.ResponseWriter, r *http.Request) {
    sourceIdMutex.Lock()
    idStr := strconv.Itoa(sourceId);
    sourceId++
    sourceIdMutex.Unlock()
    sources[idStr] = make(map[*websocket.Conn]bool)

    log.Println("new data source: " +  idStr)

    ws, err := upgrader.Upgrade(w, r, nil)

    // websocket endpoint for data source
    http.HandleFunc(endpoint + "/out/" + idStr, DataOutHandler)
    // metadata endpoint for data source
    http.HandleFunc(endpoint + "/out/" + idStr + "/metadata", Metadata)

    if err != nil {
        log.Println(err)
        return
    }

    /**
     * read source stream header
     * delimited from its data by a character
     */
    delimiterFound := false
    var sourceHeader []byte
    for {
        _, p, err := ws.ReadMessage()
        if err != nil {
            delete(sources, idStr)
            log.Println(err)
            return
        }
        for i := 0; i < len(p); i++ {
            if p[i] == HEADER_DELIMITER {
                delimiterFound = true

                // write the remainder of this message to the client
                message := p[i+1:]
                for client := range sources[idStr] {
                    if err := client.WriteMessage(websocket.BinaryMessage, message); err != nil {
                        delete(sources[idStr], client)
                        log.Println(err)
                    }
                }
                // save the source metadata to our dictionary
                metadata[idStr] = sourceHeader
                log.Println("source (" + idStr + ") header: " + string(sourceHeader))
                break
            } else {
                // keep reading the header
                sourceHeader = append(sourceHeader, p[i])
            }

        }

        if (delimiterFound) {
            break
        }
    }

    /**
     *  start broadcasting data to clients
     */
    for {
        _, p, err := ws.ReadMessage()
        if err != nil {
            delete(sources, idStr) // delete the source
            delete(metadata, idStr) // delete the metadata
            ws.Close() // close the socket
            log.Println(err)
            return
        }
        for client := range sources[idStr] {
            if err := client.WriteMessage(websocket.BinaryMessage, p); err != nil {
                delete(sources[idStr], client)
                client.Close() // close the socket
                log.Println(err)
            }
        }
    }
}

/**
 * a websocket that streams data received from  sources out
 */
func DataOutHandler(w http.ResponseWriter, r *http.Request) {
    url := strings.Split(r.URL.String(), "/")
    sourceId := url[len(url) - 1]
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
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

/**
 * list all endpoints streaming to the service in a JSON array
 */
func ListStreams(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*") // CORS

    sourceIds := make([]string, 0, len(sources))
    for k := range sources {
        sourceIds = append(sourceIds, k)
    }
    encoder := json.NewEncoder(w)
    encoder.Encode(sourceIds)
}

/**
 * Get the metadata for a data source
 */
func Metadata(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*") // CORS

    url := strings.Split(r.URL.String(), "/")
    sourceId := url[len(url) - 2]

    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(metadata[sourceId]))
}

func Start(e string, h []byte) {
    // set globals
    endpoint = e
    header = h

    http.HandleFunc(endpoint + "/list", ListStreams)
    http.HandleFunc(endpoint + "/in", DataInHandler)
}
