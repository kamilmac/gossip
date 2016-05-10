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

type request struct {
    Password    string              `json:"password"`
}

type attachReq struct {
    request
    Data        map[string]string   `json:"data"`
}

type attachRes struct {
    response    
}

type getReq struct {
    request
}

type getRes struct {
    Data        map[string]string   `json:"data"`
    response
}

type setReq struct {
    Key         string              `json:"key"`
    Value       string              `json:"value"`
    request
}

type setRes struct {
    response
}

type delReq struct {
    Key         string              `json:"key"`
    request
}

type delRes struct {
    response
}

type broadcastReq struct {
    request
}

type broadcastRes struct {
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
    res.Status = "error"
    defer func() {
        json, err := json.Marshal(res)
        if err != nil {
            log.Println(err)
        }
        w.Write(json)
    }()
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&req)
    if err != nil {
        res.Message = "Json req decoding error"
        return
    }
    if req.Password != pass {
        res.Message = "Wrong password"
        return
    }
    for k, v := range req.Data {
        kv[string(k)] = string(v)
    }
    deadSubscribers := validateSubscribers(kv)
    log.Println("validateSubscribers done. Dead subscribers: ", deadSubscribers)
    kv = removeDeadSubscribers(kv, deadSubscribers)
    log.Println("Dead subscribers removed")
    broadcast(kv)
    log.Println("Broadcast done")
    res.Status = "success"
    res.Message = "Broadcast done"
}

func validateSubscribers(store map[string]string) map[string]string {
    time.Sleep(1*time.Second)
    subscribers := extractSubscribers(store)
    deadSubscribers := make(map[string]string)
    var wg sync.WaitGroup
    for name, callback := range subscribers {
        wg.Add(1)
        go func(name, callback string) {
            req, err := http.NewRequest("POST", callback, bytes.NewBuffer([]byte("{}")))
            if err != nil {
                log.Println("validateSubscribers request err: ", err)
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
                log.Println("validateSubscribers response err: ", err)
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

func handleGet(w http.ResponseWriter, r *http.Request) {
    var (
        res getRes
        req getReq
    )
    res.Status = "error"
    defer func() {
        json, err := json.Marshal(res)
        if err != nil {
            log.Println(err)
        }
        w.Write(json)
    }()
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&req)
    if err != nil {
        res.Message = "Json req decoding error"
        return
    }
    if req.Password != pass {
        res.Message = "Wrong password"
        return
    }
    res.Data = kv
    res.Status = "success"
}

func handleSet(w http.ResponseWriter, r *http.Request) {
    var (
        res setRes
        req setReq
    )
    res.Status = "error"
    defer func() {
        json, err := json.Marshal(res)
        if err != nil {
            log.Println(err)
        }
        w.Write(json)
    }()
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&req)
    if err != nil {
        res.Message = "Json req decoding error"
        return
    }
    if req.Password != pass {
        res.Message = "Wrong password"
        return
    }
    kv[req.Key] = req.Value
    res.Status = "success"
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
    var (
        res delRes
        req delReq
    )
    res.Status = "error"
    defer func() {
        json, err := json.Marshal(res)
        if err != nil {
            log.Println(err)
        }
        w.Write(json)
    }()
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&req)
    if err != nil {
        res.Message = "Json req decoding error"
        return
    }
    if req.Password != pass {
        res.Message = "Wrong password"
        return
    }
    delete(kv, req.Key)
    res.Status = "success"
}

func handleBroadcastNow(w http.ResponseWriter, r *http.Request) {
    var (
        req broadcastReq
        res broadcastRes
    )
    res.Status = "error"
    defer func() {
        json, err := json.Marshal(res)
        if err != nil {
            log.Println(err)
        }
        w.Write(json)
    }()
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&req)
    if err != nil {
        res.Message = "Json req decoding error"
        return
    }
    if req.Password != pass {
        res.Message = "Wrong password"
        return
    }
    deadSubscribers := validateSubscribers(kv)
    log.Println("validateSubscribers done. Dead subscribers: ", deadSubscribers)
    kv = removeDeadSubscribers(kv, deadSubscribers)
    log.Println("Dead subscribers removed")
    broadcast(kv)
    log.Println("Broadcast done")
    res.Status = "success"
    res.Message = "Broadcast done"
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
    http.HandleFunc("/attach", handleAttach)
    http.HandleFunc("/broadcastnow", handleBroadcastNow)
    http.HandleFunc("/get", handleGet)
    http.HandleFunc("/set", handleSet)
    http.HandleFunc("/del", handleDelete)
    err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}