package main

import (
	"fmt"
	"gopkg.in/redis.v3" 
	"github.com/julienschmidt/httprouter"
	"net/http"
	"encoding/json"
)

/*
type Data struct{
	Key string `json:"key"`
	Value string `json:"value"`
}

type DataArray struct{
	RedisData []Data
}

func (data *DataArray) AddKeyVal(keyVal Data) []Data {
	data.RedisData = append(data.RedisData, keyVal)
	return data.RedisData
}
*/

type data map[string]string

func postData(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {  
    // Stub an user to be populated from the body
	fmt.Println("In post")
    var d data

    // Populate the user data
    json.NewDecoder(r.Body).Decode(&d)

    client := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "redis", // no password set
        DB:       0,  // use default DB
    })  
    pong, err := client.Ping().Result()
    fmt.Println(pong, err)
    // Output: PONG <nil>

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

func main() {
	//server code start
	r := httprouter.New()
	r.POST("/keyvals",postData)
	fmt.Println("in Main()")
    	 server := http.Server{

            Addr:        "0.0.0.0:3030",

            Handler: r,

    }

    server.ListenAndServe()
 	//server code ends here
}
