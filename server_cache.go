package main

import (
    "fmt"
    "gopkg.in/redis.v3"
    "github.com/julienschmidt/httprouter"
    "net/http"
    "encoding/json"
)
const (
    redis_cache_one = "54.175.28.88:6379"
    redis_main_db = "52.91.39.197:6379"
    redis_localhost = "localhost:6379"
)

type data map[string]string

func postData(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    // Stub an user to be populated from the body
    var d data

    //First put data into the main db then once successful put it in Cache
    for key,val := range d{
        putError := postDataToServer(key, val, redis_main_db)
        if putError != nil{
            panic(putError)
            return
        }
    }

    // Populate the user data
    json.NewDecoder(r.Body).Decode(&d)

    client := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })
    _, err := client.Ping().Result()
    if err != nil {
        panic(err)
    }

    for key,val :=range d {
        err1 := client.Set(key, val, 0).Err()
        if err1 != nil {
            panic(err1)
        }
    }

    // Marshal provided interface into JSON structure
    uj, _ := json.Marshal(d)

    // Write content-type, statuscode, payload
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(201)
    fmt.Fprintf(w, "%s", uj)
}
func getData(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    key :=  p.ByName("key")
    client := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })
    _, err := client.Ping().Result()
    if err != nil {
        panic(err)
    }

    val, err := client.Get(key).Result()
    if err == redis.Nil {
        fmt.Println("key does not exists in cache")

        //Get the values from the main database
        value, err := getFromMainDatabase(key)

        if err == redis.Nil{
            fmt.Println("key does not exists in main db as well")
        }
        //update the cache
        postDataToServer(key, value,redis_localhost)
        val = value
        rw.Header().Set("System-Type", "MainDB")

    } else if err != nil {
        panic(err)
    } else {
        //fmt.Println("key ", val)
        fmt.Println("key exists in cache")
        rw.Header().Set("System-Type", "Cache")
    }
    uj, _ := json.Marshal(val)

    // Write content-type, statuscode, payload
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(200)
    fmt.Fprintf(rw, "%s", uj)

}

func getFromMainDatabase(missingKey string) (value string, errr error){
    client := redis.NewClient(&redis.Options{
        Addr:     redis_main_db,
        Password: "", // no password set
        DB:       0,  // use default DB
    })
    _, err := client.Ping().Result()
    if err != nil {
        panic(err)
    }

    val, err := client.Get(missingKey).Result()
    if err == redis.Nil {
        fmt.Println("key does not exists")
    } else if err != nil {
        panic(err)
    } else {
        //fmt.Println("key ", val)
    }
    return val,err
}

func postDataToServer(key string,val string, serverIpAndPort string) (er error){

    client := redis.NewClient(&redis.Options{
        Addr:     serverIpAndPort,
        Password: "", // no password set
        DB:       0,  // use default DB
    })
    _, err := client.Ping().Result()
    if err != nil {
        panic(err)
    }

    err1 := client.Set(key, val, 0).Err()
    if err1 != nil {
        panic(err1)
    }

    return err
}
func main() {
    //server code start
    r := httprouter.New()

    r.POST("/keyvals",postData)
    r.GET("/keyvals/:key",getData)

    fmt.Println("Server Started ...")

    server := http.Server{
        Addr:        "0.0.0.0:3031",
        Handler: r,
    }

    server.ListenAndServe()
    //server code ends here
}

