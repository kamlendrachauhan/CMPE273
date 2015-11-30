package main

import (
    "fmt"
    "gopkg.in/redis.v3"
    "github.com/julienschmidt/httprouter"
    "net/http"
    "encoding/json"
    "image"
    "bytes"
    "image/jpeg"
    "log"
    "strconv"
)
const (
    redis_cache_one = "54.175.28.88:6379"
    redis_main_db = "52.91.39.197:6379"
    redis_localhost = "54.175.28.88:6379"
)

type data map[string]string

func postData(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    // Stub an user to be populated from the body
    var d data

 // Populate the user data
    json.NewDecoder(r.Body).Decode(&d)

    //First put data into the main db then once successful put it in Cache
    for key,val := range d{
        putError := postDataToServer(key, val, redis_main_db)
        if putError != nil{
            panic(putError)
            return
        }
    }

    // Populate the user data
//    json.NewDecoder(r.Body).Decode(&d)

    client := redis.NewClient(&redis.Options{
        Addr:     "54.175.28.88:6379",
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
        Addr:     "54.175.28.88:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })
    pingresult, err := client.Ping().Result()
    if err != nil {
        panic(err)
    }

    fmt.Println(pingresult)
    fmt.Println("outside the client creation")

     val, err := client.Get(key).Result()
    if err == redis.Nil {
        fmt.Println("key does not exists")

        //Get the values from the main database
        value, err := getFromMainDatabase(key)

        if err == redis.Nil{
            fmt.Println("key does not exists in main db as well")
        }

        

        //update the cache
        postDataToServer(key, value,redis_localhost)
        val = value
    } else if err != nil {
        panic(err)
    } else {
        //fmt.Println("key ", val)
    }

    bytesImage := []byte(val)
    img, _, _ := image.Decode(bytes.NewReader(bytesImage))
    sendImage(rw, &img)

    
}



func getData_old(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
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
        fmt.Println("key does not exists")

        //Get the values from the main database
        value, err := getFromMainDatabase(key)

        if err == redis.Nil{
            fmt.Println("key does not exists in main db as well")
        }
        //update the cache
        postDataToServer(key, value,redis_localhost)
        val = value
    } else if err != nil {
        panic(err)
    } else {
        fmt.Println("key ", val)
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
        fmt.Println("key ", val)
    }
    return val,err
}

func postDataToServer(key string,val string, serverIpAndPort string) (er error){

    client := redis.NewClient(&redis.Options{
        Addr:     serverIpAndPort,
        Password: "", // no password set
        DB:       0,  // use default DB
    })
    pResult, err := client.Ping().Result()
    fmt.Println("result from second ping",pResult)
    if err != nil {
        panic(err)
    }


    err1 := client.Set(key, val, 0).Err()
    if err1 != nil {
        panic(err1)
    }else {
    fmt.Println("Key has been added to main server",key)
    }
    return err
}





func postDataToServer_old(key string,val string, serverIpAndPort string) (er error){

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
    }else {
	fmt.Println("Key has been added to main server",key)
	}
    return err
}

// writeImage encodes an image in jpeg format and writes it into ResponseWriter.
func sendImage(w http.ResponseWriter, img *image.Image) {
    fmt.Println("inside sendIma")
    buffer := new(bytes.Buffer)
    if err := jpeg.Encode(buffer, *img, nil); err != nil {
        log.Println("unable to encode image.")
    }

    w.Header().Set("Content-Type", "image/jpeg")
    w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
    if _, err := w.Write(buffer.Bytes()); err != nil {
        log.Println("unable to write image.")
    }
    //w.WriteHeader(200)
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
