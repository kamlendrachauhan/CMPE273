package main

import (
    "gopkg.in/redis.v3"
    "fmt"
    "strings"
    "github.com/julienschmidt/httprouter"
    "net/http"
    "encoding/json"
    "io/ioutil"
    "strconv"
)
const (
    client_host_url = "http://54.164.93.210:3004"
    client_get_all_nodes_url = client_host_url + "/nodes"
    main_db_redis_ip_port = "52.91.39.197:6379"
)
type NodeData struct {
    NodeNumber string `json:"node_number"`
    NodeType string `json:"node_type"`
    NodeIP string `json:"node_ip"`
    MemoryUsed string `json:"memory_used"`
    NoOfKeys string `json:"no_of_keys"`
    NoOfKeyHit string `json:"no_of_key_hit"`
    NoOfKeyMiss string `json:"no_of_key_miss"`
}
type AllNodes struct
{
    Values []NodeUrl `json:"value"`
}
type NodeUrl struct
{
    Url string `json:"url"`
}

func main() {
	fmt.Println("Client Get All Node Info starting at 3005")
    mux := httprouter.New()
    mux.GET("/getAllNodeInfo", getAllDBData)
    server := http.Server{
        Addr:        "0.0.0.0:3005",
        Handler: mux,
    }
    server.ListenAndServe()
}

func getAllDBData(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    fmt.Println("inside Get", client_get_all_nodes_url)
    var node_info[] NodeData

    response, err := http.Get(client_get_all_nodes_url)
    if err != nil {
        return
    }
    defer response.Body.Close()
    body, _ := ioutil.ReadAll(response.Body)
    allNodes := AllNodes{}
    err = json.Unmarshal(body, &allNodes)
    if err != nil {
        return
    }
     mainNodeNum := 0
    for i := 0; i<len(allNodes.Values); i++ {
        dbUrl := allNodes.Values[i].Url
        nodeInfo := getInfoFromCacheAndMDB(dbUrl, allNodes,i,"Cache")
        node_info = append(node_info, nodeInfo)
        mainNodeNum ++
    }
    nodeInfo := getInfoFromCacheAndMDB(main_db_redis_ip_port, allNodes,mainNodeNum,"Main Database")
    node_info = append(node_info, nodeInfo)
    uj, _ := json.Marshal(node_info)

    // Write content-type, statuscode, payload
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(200)
    fmt.Fprintf(rw, "%s", uj)
}

func getInfoFromCacheAndMDB(dbUrl string, allNodes AllNodes, count int, dbtype string) (nodeI NodeData){
    if(count != len(allNodes.Values)) {
        dbUrl = strings.Split(dbUrl, ":3")[0]
        dbUrl = dbUrl + ":6379"
        dbUrl = strings.Split(dbUrl, "//")[1]
    }
    fmt.Println(dbUrl)
    keyValMap := make(map[string]string)
    client := redis.NewClient(&redis.Options{
        Addr:     dbUrl,
        Password: "", // no password set
        DB:       0, // use default DB
    })
    _, err1 := client.Ping().Result()
    if err1 != nil {
        panic(err1)
    }
    clientInfo := client.Info().String()
    //clientInfo := client.Info().String()
    splitted := strings.Split(clientInfo, "\n")
    for i := 0; i<len(splitted); i++ {
        if (!strings.Contains(splitted[i], "#") || !strings.Contains(splitted[i], "")) {
            if (strings.Contains(splitted[i], "used_memory_human") || strings.Contains(splitted[i], "keyspace_hits") || strings.Contains(splitted[i], "keyspace_misses") || strings.Contains(splitted[i], "db0")) {
                moreSplits := strings.Split(splitted[i], ":")
                if (strings.Contains(moreSplits[1], "=")) {
                    keySplits := strings.Split(strings.Split(moreSplits[1], ",")[0], "=")
                    keyValMap[keySplits[0]]=keySplits[1]
                }else {
                    keyValMap[moreSplits[0]]= moreSplits[1]
                }
            }
        }
    }
    for key, value := range keyValMap {
        fmt.Println(key, value)
    }
    nodeInfo := NodeData{
        NodeNumber: "Node "+strconv.Itoa(count+1),
        NodeIP: strings.Split(dbUrl, ":")[0],
        NodeType: dbtype,
        MemoryUsed:keyValMap["used_memory_human"],
        NoOfKeys:keyValMap["keys"],
        NoOfKeyHit:keyValMap["keyspace_hits"],
        NoOfKeyMiss:keyValMap["keyspace_misses"],
    }
return nodeInfo
}