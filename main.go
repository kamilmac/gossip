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
            if strings.Index(k, "subscriber:") == 0 {
                res.Message = fmt.Sprintf("%v /n %v@%v appended as subscriber", res.Message, k, v)
            }
        } 
        res.Status = "success"
        go broadcast(kv)
    }
    json, err := json.Marshal(res)
    if err != nil {
        fmt.Println(err)
    }
    w.Write(json)
}

func broadcast(store map[string]string) {
    time.Sleep(1*time.Second)
    msg := broadcastMsg{
        Data: store,
        Password: pass,
    }
    msg_json, _ := json.Marshal(msg)
    subscribers := []string{}
    for k, v := range kv {
        if strings.Index(k, "subscriber:") == 0 {
            subscribers = append(subscribers, v)  
        } 
    }
    for _, url := range subscribers {
        req, _ := http.NewRequest("POST", url, bytes.NewBuffer(msg_json))
        req.Header.Set("Content-Type", "application/json")
        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
            fmt.Println(err)
        }
        defer resp.Body.Close()
    }
}

func init() {
    flag.StringVar(&pass, "pass", "testpass", "Create password")
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
