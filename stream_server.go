package streamserver

import (
    "log"
    "sync"
    "strings"
    "strconv"
    "net/http"
    "encoding/json"
    "encoding/base64"

    "github.com/gorilla/websocket"
)

/**
 * Header delimiter
 * if the first byte of data from a source is this character then
 * data streaming from a source is read in as a metadata header
 * until we find this character and then the remainder of data
 * streamed in is sent to the clients
 */
var HEADER_DELIMITER byte = '|'

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

var endpoint string
var sourceId = 1
var sourceIdMutex = &sync.Mutex{}

var sources = make(map[string]map[*websocket.Conn]bool)
var metadata = make(map[string]string)
var magicBytes = make(map[string][]byte)

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
     * read source stream header delimited from
     * its data by the `HEADER_DELIMITER` character
     */
    openingDelimiterFound := false
    closingDelimiterFound := false
    var sourceHeader []byte

    // read the first message and look for the header delimiter
    _, p, err := ws.ReadMessage();
    if err != nil {
        delete(sources, idStr)
        log.Println(err)
        return
    }

    // if a header is detected
    if p[0] == HEADER_DELIMITER {
        openingDelimiterFound = true
        // iterate over the remainder of the first message
        for i := 1; i < len(p); i++ {
            if p[i] == HEADER_DELIMITER {
                // if we have found the header closing delimiter in the
                // first message broadcast the rest of the first message
                closingDelimiterFound = true
                log.Println("source (" + idStr + ") header: " + string(sourceHeader))
                for client := range sources[idStr] {
                    if err := client.WriteMessage(websocket.BinaryMessage, p[i+1:]); err != nil {
                        delete(sources[idStr], client)
                        log.Println(err)
                    }
                }
                break
            }

            // TODO: don't use append here. Its probably slow as shit
            sourceHeader = append(sourceHeader, p[i])
        }
    }

    if openingDelimiterFound && !closingDelimiterFound {
        // keep reading until we find the closing delimiter
        // TODO: limit the possible size of the header
        for {
            _, p, err := ws.ReadMessage()
            if err != nil {
                delete(sources, idStr)
                log.Println(err)
                return
            }

            for i := 0; i < len(p); i++ {
                // if we find the closing delimiter
                if p[i] == HEADER_DELIMITER {
                    closingDelimiterFound = true
                    // broadcast the rest of this message
                    for client := range sources[idStr] {
                        if err := client.WriteMessage(websocket.BinaryMessage, p[i+1:]); err != nil {
                            delete(sources[idStr], client)
                        }
                    }
                    break // stop reading the header
                }
                sourceHeader = append(sourceHeader, p[i])
            }

            if closingDelimiterFound {
                break
            }
        }
    }

    /**
     * parse header JSON
     */
    if openingDelimiterFound {
        var dat map[string]interface{}
        if err := json.Unmarshal(sourceHeader, &dat); err != nil {
            // can't parse header JSON. Bail
            log.Println("Cannot decode JSON header")
            log.Println(err)
            return
        }

        decodedData, err := base64.StdEncoding.DecodeString((dat["magicBytes"].(string)))
        if err != nil {
            log.Println("Cannot base64 decode header magic bytes")
            log.Println(err)
            return
        }
        magicBytes[idStr] = decodedData
        metadata[idStr] = dat["metadata"].(string)
        log.Println("source (" + idStr + ") magic bytes: " + string(magicBytes[idStr]))
        log.Println("source (" + idStr + ") metadata: " + metadata[idStr])
    }


    /**
     *  start broadcasting data to clients
     */
    for {
        _, p, err := ws.ReadMessage()
        if err != nil {
            delete(sources, idStr) // delete the source
            delete(metadata, idStr) // delete the metadata
            delete(magicBytes, idStr)
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

    /* send initial header to client (such as magic bytes for jsmpeg video) */
    if mb, ok := magicBytes[sourceId]; ok {
        if err := ws.WriteMessage(websocket.BinaryMessage, mb); err != nil {
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

func Start(e string/*, h []byte*/) {
    // set globals
    endpoint = e

    http.HandleFunc(endpoint + "/list", ListStreams)
    http.HandleFunc(endpoint + "/in", DataInHandler)
}
