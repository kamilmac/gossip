package main

import (
    "fmt"
    "log"
    "net/http"
    "encoding/json"
    "bytes"
    "flag"
)

type response struct {
    Status      string              `json:"status"`
    Message     string              `json:"message"`
}

type attachReq struct {
    Password    string              `json:"password"`
    Name        string              `json:"name"`
    Callback    string              `json:"callback"`
    Data        map[string]string   `json:"data"`
}

type attachRes struct {
    response    
}

var (
    pass string
    port int
    kv = make(map[string]string)
    containers = make(map[string]string)
) 

func init() {
    flag.StringVar(&pass, "pass", "qpwoeiruty", "Create password")
    flag.IntVar(&port, "port", 3000, "Specify port")
    flag.Parse()
    fmt.Println("Running on port: ", port)
}

func main() {
    http.HandleFunc("/attach", handleAttach)
    err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

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
        if req.Callback != "" {
            containers[req.Name] = req.Callback
            res.Message = fmt.Sprintf("Container (%v: %v) appended as observer", req.Name, req.Callback)
        }
        for k, v := range req.Data {
            kv[string(k)] = string(v)
        } 
        res.Status = "success"
        data, _ := json.Marshal(kv)
        go broadcast(data)
    }
    json, err := json.Marshal(res)
    if err != nil {
        fmt.Println(err)
    }
    w.Write(json)
}

func broadcast(store []byte) {
    for _, url := range containers {
        req, _ := http.NewRequest("POST", url, bytes.NewBuffer(store))
        req.Header.Set("Content-Type", "application/json")
        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
            fmt.Println(err)
        }
        defer resp.Body.Close()
    }
}

func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}
