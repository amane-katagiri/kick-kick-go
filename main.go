package main

import (
    "fmt"
    "html/template"
    "log"
    "net/http"
    "runtime"

    "github.com/amane-katagiri/kick-kick-go/config"
    "github.com/amane-katagiri/kick-kick-go/storage"
    "github.com/amane-katagiri/kick-kick-go/storage/redis"
    "github.com/amane-katagiri/kick-kick-go/websocket"
)

var wsUrl = "wss?://host/path/to/ws"
var tmpl *template.Template

func IndexHandler(w http.ResponseWriter, r *http.Request) {
    tmpl.ExecuteTemplate(w, "index", map[string]interface{}{"WsUrl": wsUrl})
}

func main() {
    cpus := runtime.NumCPU()
    runtime.GOMAXPROCS(cpus)

    storage.LoadFlag()
    config.LoadFlag()
    err := storage.LoadConfig()
    if err != nil {
        panic(err)
    }
    config, err := config.LoadConfig()
    if err != nil {
        panic(err)
    }
    if config.Server.Key != "" {
        wsUrl = fmt.Sprintf("wss://%s:%d%s", config.Server.Host, config.Server.Port, config.Server.WsPath)
    } else {
        wsUrl = fmt.Sprintf("ws://%s:%d%s", config.Server.Host, config.Server.Port, config.Server.WsPath)
    }
    tmpl, err = template.New("").ParseFiles(config.TemplateFiles...)
    if err != nil {
        log.Println(err)
        tmpl, _ = template.New("err").Parse("template file is not found")
    }

    http.HandleFunc("/", IndexHandler)
    http.HandleFunc(config.Server.WsPath, websocket.WsHandler)
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(config.StaticDir))))

    r := redis.NewRedisStorage()
    go websocket.ServeCount(r.GetCount(), r.SetCount)
    go websocket.ServeClients()
    go websocket.ServeId()

    if config.Server.Key != "" {
        log.Printf("Serving at https://%s:%d", config.Server.Host, config.Server.Port)
        log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", config.Server.Port), config.Server.Key, config.Server.Cert, nil))
    } else {
        log.Printf("Serving at http://%s:%d", config.Server.Host, config.Server.Port)
        log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Server.Port), nil))
    }
}
