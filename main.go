package main

import (
    "fmt"
    "log"
    "net/http"
    "encoding/json"
    "bytes"
    "flag"
    "strings"
    "time"
    "sync"
)

type response struct {
    Status      string              `json:"status"`
    Message     string              `json:"message"`
}

type attachReq struct {
    Password    string              `json:"password"`
    Data        map[string]string   `json:"data"`
}

type attachRes struct {
    response    
}

type broadcastMsg struct {
    Password    string              `json:"password"`
    Data        map[string]string   `json:"data"`
}

const (
    connectionTimeout = 10
)
var (
    pass string
    port int
    kv = make(map[string]string)
) 

func handleAttach(w http.ResponseWriter, r *http.Request) {
    var (
        res attachRes
        req attachReq
    )
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&req)
    if err != nil {
        res.Status = "error"
        res.Message = "Json req decoding error"
    } else if req.Password != pass {
        res.Status = "error"
        res.Message = "Wrong password"
    } else {
        res.Message = ""
        for k, v := range req.Data {
            kv[string(k)] = string(v)
        } 
        res.Status = "success"
        deadSubscribers := handshake(kv)
        log.Println("Handshake done. Dead subscribers: ", deadSubscribers)
        kv = removeDeadSubscribers(kv, deadSubscribers)
        log.Println("Dead subscribers removed")
        broadcast(kv)
        log.Println("Broadcast done")
    }
    json, err := json.Marshal(res)
    if err != nil {
        log.Println(err)
    }
    w.Write(json)
}

func handshake(store map[string]string) map[string]string {
    time.Sleep(1*time.Second)
    subscribers := extractSubscribers(store)
    deadSubscribers := make(map[string]string)
    var wg sync.WaitGroup
    for name, callback := range subscribers {
        wg.Add(1)
        go func(name, callback string) {
            req, err := http.NewRequest("POST", callback, bytes.NewBuffer([]byte("{}")))
            if err != nil {
                log.Println("Handshake request err: ", err)
                log.Println("Removing subscriber: ", callback)
                deadSubscribers[name] = callback
                wg.Done()
                return
            }
            req.Header.Set("Content-Type", "application/json")
            client := &http.Client{
                Timeout: time.Duration(connectionTimeout * time.Second),
            }
            resp, err := client.Do(req)
            if err != nil {
                log.Println("Handshake response err: ", err)
                log.Println("Removing subscriber: ", callback)
                deadSubscribers[name] = callback
                wg.Done()
                return
            }
            // TODO: VALIDATE RESPONSE WITH PASSWORD
            defer resp.Body.Close()
            wg.Done()
        }(name, callback)
    }
    wg.Wait()
    return deadSubscribers
}

func broadcast(store map[string]string) {
    msg, _ := json.Marshal(broadcastMsg{
        Data: store,
        Password: pass,
    })
    subscribers := extractSubscribers(store)
    var wg sync.WaitGroup
    for _, callback := range subscribers {
        wg.Add(1)
        go func(callback string) {
            req, err := http.NewRequest("POST", callback, bytes.NewBuffer(msg))
            if err != nil {
                log.Println("Broadcast request err: ", err)
                wg.Done()
                return
            }
            req.Header.Set("Content-Type", "application/json")
            client := &http.Client{
                Timeout: time.Duration(connectionTimeout * time.Second),
            }
            resp, err := client.Do(req)
            if err != nil {
                log.Println("Broadcast response err: ", err)
                wg.Done()
                return
            }
            defer resp.Body.Close()
            wg.Done()
        }(callback)
    }
    wg.Wait()
}

func extractSubscribers(store map[string]string) map[string]string {
    subscribers := make(map[string]string)
    for k, v := range store {
        if strings.Index(k, "subscriber:") == 0 {
            subscribers[k] = v  
        } 
    }
    return subscribers
}

func removeDeadSubscribers(store, deadSubscribers map[string]string) map[string]string {
    for k := range store {
        for name := range deadSubscribers {
            if name == k {
                delete(store, name)
            }
        }
    }
    return store
}

func init() {
    flag.StringVar(&pass, "pass", "testpass", "Create password")
    flag.IntVar(&port, "port", 3000, "Specify port")
    flag.Parse()
    log.Println("Running on port: ", port)
}

func main() {
    http.HandleFunc("/set", handleAttach)
    err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
