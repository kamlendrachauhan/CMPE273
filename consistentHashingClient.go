package main

import (
"crypto/md5"
"fmt"
"math"
"sort"
"net/http"
"errors"
"encoding/json"
"io/ioutil"
"bytes"
"github.com/julienschmidt/httprouter"
"time"
"os"
"strconv"
"strings"
"github.com/rs/cors"
)

type HashKey uint32
type HashKeyOrder []HashKey

var ring *HashRing

var nodesList []string 
var LOG_FILE_NAME="dat.log"

func (h HashKeyOrder) Len() int           { return len(h) }
func (h HashKeyOrder) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h HashKeyOrder) Less(i, j int) bool { return h[i] < h[j] }

type HashRing struct {
	ring       map[HashKey]string
	sortedKeys []HashKey
	nodes      []string
	weights    map[string]int
}

func New(nodes []string) *HashRing {
	hashRing := &HashRing{
		ring:       make(map[HashKey]string),
		sortedKeys: make([]HashKey, 0),
		nodes:      nodes,
		weights:    make(map[string]int),
	}
	//adding to the nodeList array
	nodesList=nodes
	hashRing.generateCircle()
	return hashRing
}

func NewWithWeights(weights map[string]int) *HashRing {
	nodes := make([]string, 0, len(weights))
	for node, _ := range weights {
		nodes = append(nodes, node)
	}
	hashRing := &HashRing{
		ring:       make(map[HashKey]string),
		sortedKeys: make([]HashKey, 0),
		nodes:      nodes,
		weights:    weights,
	}
	hashRing.generateCircle()
	return hashRing
}

func (h *HashRing) generateCircle() {
	totalWeight := 0
	for _, node := range h.nodes {
		if weight, ok := h.weights[node]; ok {
			totalWeight += weight
		} else {
			totalWeight += 1
		}
	}

	for _, node := range h.nodes {
		weight := 1

		if _, ok := h.weights[node]; ok {
			weight = h.weights[node]
		}

		factor := math.Floor(float64(40*len(h.nodes)*weight) / float64(totalWeight))

		for j := 0; j < int(factor); j++ {
			nodeKey := fmt.Sprintf("%s-%d", node, j)
			bKey := hashDigest(nodeKey)

			for i := 0; i < 3; i++ {
				key := hashVal(bKey[i*4 : i*4+4])
				h.ring[key] = node
				h.sortedKeys = append(h.sortedKeys, key)
			}
		}
	}

	sort.Sort(HashKeyOrder(h.sortedKeys))
}

func (h *HashRing) GetNode(stringKey string) (node string, ok bool) {

	pos, ok := h.GetNodePos(stringKey)
	if !ok {
		return "", false
	}
	return h.ring[h.sortedKeys[pos]], true
}

func (h *HashRing) GetNodePos(stringKey string) (pos int, ok bool) {
	if len(h.ring) == 0 {
		fmt.Println("h is zero")
		return 0, false
	}

	key := h.GenKey(stringKey)

	nodes := h.sortedKeys
	pos = sort.Search(len(nodes), func(i int) bool { return nodes[i] > key })

	if pos == len(nodes) {
		// Wrap the search, should return first node
		return 0, true
	} else {
		return pos, true
	}
}

func writeToAFile(text string){

	f, err2 := os.OpenFile(LOG_FILE_NAME, os.O_APPEND, 0666)
	if err2!=nil{
		fmt.Printf(err2.Error())
	}
	defer f.Close()


    //n3, err := f.WriteString("writes\n")
	_, err := f.WriteString(text+"\n")
	if err!=nil{
		fmt.Printf(err.Error())
	}
	//fmt.Printf("wrote %d bytes\n", n3)
	f.Sync()
}





func (h *HashRing) GenKey(key string) HashKey {
	bKey := hashDigest(key)
	return hashVal(bKey[0:4])
}

func (h *HashRing) GetNodes(stringKey string, size int) (nodes []string, ok bool) {
	pos, ok := h.GetNodePos(stringKey)
	if !ok {
		return []string{}, false
	}

	if size > len(h.nodes) {
		return []string{}, false
	}

	returnedValues := make(map[string]bool, size)
	mergedSortedKeys := append(h.sortedKeys[pos:], h.sortedKeys[:pos]...)
	resultSlice := []string{}

	for _, key := range mergedSortedKeys {
		val := h.ring[key]
		if !returnedValues[val] {
			returnedValues[val] = true
			resultSlice = append(resultSlice, val)
		}
		if len(returnedValues) == size {
			break
		}
	}

	return resultSlice, len(resultSlice) == size
}

func (h *HashRing) AddNode(node string) *HashRing {
	return h.AddWeightedNode(node, 1)
}

func (h *HashRing) AddWeightedNode(node string, weight int) *HashRing {
	if weight <= 0 {
		return h
	}

	for _, eNode := range h.nodes {
		if eNode == node {
			return h
		}
	}

	nodes := make([]string, len(h.nodes), len(h.nodes)+1)
	copy(nodes, h.nodes)
	nodes = append(nodes, node)

	weights := make(map[string]int)
	for eNode, eWeight := range h.weights {
		weights[eNode] = eWeight
	}
	weights[node] = weight

	hashRing := &HashRing{
		ring:       make(map[HashKey]string),
		sortedKeys: make([]HashKey, 0),
		nodes:      nodes,
		weights:    weights,
	}
	hashRing.generateCircle()
	//Adding the node to our nodesList
	nodesList = append(nodesList, node)

	return hashRing
}

func (h *HashRing) RemoveNode(node string) *HashRing {
	nodes := make([]string, 0)
	for _, eNode := range h.nodes {
		if eNode != node {
			nodes = append(nodes, eNode)
		}
	}

	weights := make(map[string]int)
	for eNode, eWeight := range h.weights {
		if eNode != node {
			weights[eNode] = eWeight
		}
	}

	hashRing := &HashRing{
		ring:       make(map[HashKey]string),
		sortedKeys: make([]HashKey, 0),
		nodes:      nodes,
		weights:    weights,
	}
	hashRing.generateCircle()
	//delete from nodeList
	//get index
	var found bool
	var index int
	for i:= range nodesList{
		if nodesList[i]==node{
			index=i
			found=true
		}
	}
	if found {nodesList = append(nodesList[:index],nodesList[index+1:]...)}
	return hashRing
}

func hashVal(bKey []byte) HashKey {
	return ((HashKey(bKey[3]) << 24) |
		(HashKey(bKey[2]) << 16) |
		(HashKey(bKey[1]) << 8) |
		(HashKey(bKey[0])))
}

func hashDigest(key string) []byte {
	m := md5.New()
	m.Write([]byte(key))
	return m.Sum(nil)
}



//------
func putkey(x Response)error{
	var url string
   //ring := New(servers)
	server1,z := ring.GetNode(x.Key)
	if(z==true){
		url =server1+"/"+"keyvals" }
		fmt.Println(url)

   //making a map
		m:=make(map[string]string)

   //Converting 

		//fmt.Println("Body:",x.Key,x.Value)


		m[x.Key]=x.Value

		jsonString,_:=json.Marshal(m)

		req1, errReqC := http.NewRequest("POST", url, bytes.NewBuffer(jsonString))
		if errReqC!=nil{
			errMsg:="Request creation error"
			fmt.Println(errMsg)
         //errorCheck(errMsg,rw)
			return errors.New(errMsg)
		}

		client := &http.Client{}
		resp, errClient := client.Do(req1)
		if errClient != nil {
			//fmt.Println(resp)
			fmt.Println(errClient.Error())
			errMsg:="Request creation error.Check server side."
			fmt.Println(errMsg)
        //errorCheck(errMsg,rw)
			return errors.New(errMsg)
		}
		defer resp.Body.Close()

		
        //fmt.Println(resp)



 /*  // any status code 200..299 is "success", so fail on anything else
   if resp.StatusCode < 200 || resp.StatusCode >= 300 {
   	return errors.New(http.StatusText(resp.StatusCode))*/

   	return nil
   }


   func getkey(x string)(Response,bool,error,string){
   	
   	var url string
   	server1,z := ring.GetNode(x)
   	if(z==true){
   		url =server1+"/"+"keyvals"+"/"+x
   	}
   	fmt.Println(url)
   	resp, err := http.Get(url)
   	if err != nil || resp.StatusCode >= 400 {
   		return Response{}, false, err,""
   	}

   //fmt.Println("Response:",resp)

   	defer resp.Body.Close()

   	//Getting header info:
   	header:=resp.Header
   	//fmt.Println(resp.Header)
   	systemType:=header.Get("System-Type")
   	fmt.Println("System-type:",systemType)

      if !strings.EqualFold(systemType,"Cache"){
         systemType="MainDB"
      }

   	// trying Content-Type 
   	//contentType:=header.Get("Content-Type")

   	body, err := ioutil.ReadAll(resp.Body)
   	if err != nil {
       return Response{}, false, err,""
   	}
   //fmt.Println(string(body))


   //For now a string is being sent
   	valueString:=string(body)

   //fmt.Println("value:",u.Value)
   //fmt.Println("value:",valueString)
   	u:=Response{x,valueString}
   	return u, true, nil,systemType
   }


   func getkey_original(x string)(Response,bool,error){
   	fmt.Println("inside getkey")
   	var url string
   	server1,z := ring.GetNode(x)
   	if(z==true){
   		url =server1+"/"+"keyvals"+"/"+x
   	}
   	fmt.Println(url)
   	resp, err := http.Get(url)
   	if err != nil || resp.StatusCode >= 400 {
   		return Response{}, false, err
   	}

   //fmt.Println("Response:",resp)

   	defer resp.Body.Close()

   	//Getting header info:
   	header:=resp.Header
   	fmt.Println(resp.Header)
   	systemType:=header.Get("System-Type")
   	fmt.Println("System-type:",systemType)
   	// trying Content-Type 
   	contentType:=header.Get("Content-Type")
   	fmt.Println("Content-type:",contentType)

   	body, err := ioutil.ReadAll(resp.Body)
   	if err != nil {
       //return Response{}, false, err
   	}
   //fmt.Println(string(body))


   //For now a string is being sent
   	valueString:=string(body)

   //fmt.Println("value:",u.Value)
   //fmt.Println("value:",valueString)
   	u:=Response{x,valueString}
   	return u, true, nil
   }



   type Response struct{
   	Key string `json:"key"`
   	Value string `json:"value"`
   }

   type AllNodes struct
   {
   	Values []NodeUrl `json:"value"`
   }


   type NodeUrl struct
   {
   	Url string `json:"url"`
   }

   type GetRequest struct
   {
   	Key string `json:"key"`
   }



//Add an node into the system
   func addNodeReq(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
fmt.Println("inside add node")
      //Json stuff
         var errDecode error
            newNode:=NodeUrl{}

    //decode the sent json
            errDecode=json.NewDecoder(req.Body).Decode(&newNode)
            if errDecode!=nil{
                fmt.Println(errDecode.Error())
                msg:="Json sent was Empty/Incorrect .Error: "
                fmt.Println(msg)
            }


   	//fmt.Println("inside add node")
	//The input will be a string - node link
   //	nodeString:=p.ByName("node_ip")
         nodeString:=newNode.Url
   	fmt.Println("the obtained node ip address: ", nodeString)
   	httpnode:="http://"+ nodeString

   	ring = ring.AddNode(httpnode) 
   	rw.WriteHeader(http.StatusCreated)
   }



//Get all the nodes --->Not working
   func getAllNodes(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {


      rw.Header().Set("Content-Type", "application/json")

	//The input will be a string - node link
   	var urls []NodeUrl


	//New getAllNode code
   	len1 := len(nodesList)
   	for i:=0;i<len1;i++{
   		temp:=NodeUrl{nodesList[i]}
   		//fmt.Println(temp)
   		urls=append(urls,temp)
   	}

	//Create 
   	allNodes:=AllNodes{urls}

	//marshalling into a json
   	respJson, err := json.Marshal(allNodes)
   	if err!=nil{
   		fmt.Print("Error occcured in marshalling")
   	}

   	rw.Header().Set("Content-Type","application/json")
   	rw.WriteHeader(http.StatusOK)
   	fmt.Fprintf(rw, "%s", respJson)

   }


//Delete a node from the ring
   func deleteNodeReq(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	//The input will be a string - node link
   	nodeString:=p.ByName("node_ip")
   	fmt.Println("The obtained node to be deleted: ", nodeString)
   	httpnode:="http://"+ nodeString

   	ring = ring.RemoveNode(httpnode) 
	//Printing the nodes after the delete
   	fmt.Println("Printing the nodes after the delete")
   	for i := range ring.nodes {
   		fmt.Println(i)
   	}

   	rw.WriteHeader(http.StatusCreated)
   }



//Set a key value pair
   func setKeyValue(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	//The input will be a json (ince if wwe pass http etc. will be an error)
    //  fmt.Println("inside setkey")

   	jsonInp:=Response{}

    //decode the sent json
   	errDecode:=json.NewDecoder(req.Body).Decode(&jsonInp)
   	if errDecode!=nil{
   		fmt.Println(errDecode.Error())
   		rw.WriteHeader(http.StatusBadRequest)
                //msg:="Json sent was Empty/Incorrect .Error: "
                //errorCheck(msg,rw)
   	}


      //calling the put key
   	errorPut:=putkey(jsonInp)

   	if errorPut!=nil{
   		rw.WriteHeader(http.StatusNotFound)}else{
   			rw.WriteHeader(http.StatusOK)
   		}

   	}




///get a key value pair
   	func getKeyValue(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

            //Time of req
   		t1:=time.Now()
   		//fmt.Println(t1)

   		keyString:=p.ByName("key_id")
   		
         //fmt.Println("Key given:",keyString)

      //calling the put key
   		resp,_,errorGet,statusStr:=getkey(keyString)

      //Time of obtaining the file
   		duration := time.Since(t1)
   		//fmt.Println(duration.Seconds())
   		//durString := strconv.FormatFloat(duration.Seconds(), 'f', 6, 64)
         durString := strconv.FormatFloat(duration.Seconds()*1000, 'f', 6, 64)

   		//go writeToAFile(durString+","+statusStr)


   		if errorGet!=nil{
   			fmt.Println(errorGet.Error())
   			rw.WriteHeader(http.StatusNotFound)
   			return}

      		//sending the response:
      		//marshalling into a json

   			respJson, err4 := json.Marshal(resp)
   			if err4!=nil{
   				fmt.Print("Error occcured in marshalling")
   			}

            //Final Println:
            fmt.Println("Key given:",keyString, "Duration:", durString,"msecs", "Server Type:", statusStr)
   			rw.Header().Set("Content-Type","application/json")
   			rw.WriteHeader(http.StatusOK)
            //fmt.Println("Response:"+string(respJson)) 

   			fmt.Fprintf(rw, "%s", respJson)

   		}

   		type imagePost struct {
   			Value string `json:"value"`
   		}

   		func main(){

   			fmt.Println("=========================")

   			//create file on set up only if doesnt exist
   			//LOG_FILE_NAME
   			if _, err := os.Stat(LOG_FILE_NAME); os.IsNotExist(err) {
  // path/to/whatever does not exist
   				fmt.Println("log file does not exist,creating the file")
   				f, err := os.Create(LOG_FILE_NAME)
   				if err!=nil{
   					fmt.Printf(err.Error())
   				}
   				defer f.Close()
   			}



	//Initially creating a empty ring
   			memcacheServers := []string{}
	/* memcacheServers = []string{"http://localhost:3031",
                           "http://localhost:3031",
                           "http://localhost:3031"}*/

                           memcacheServers = []string{"http://54.175.28.88:3031",
                           "http://52.91.171.117:3031",
                           "http://54.84.66.145:3031"}
                           ring = New(memcacheServers)
                          
                           mux := httprouter.New()
	//node related rest end points
                           mux.POST("/nodes", addNodeReq)
                           mux.GET("/nodes", getAllNodes)
                           mux.DELETE("/nodes/:node_ip", deleteNodeReq)

	//key-vlaue related rest end points
                           mux.POST("/keys", setKeyValue)
                           mux.GET("/keys/:key_id", getKeyValue)

                          /* server := http.Server{
                           	Addr:        "0.0.0.0:3004",
                           	Handler: mux,
                           }
                           server.ListenAndServe()
*/
                           handler := cors.Default().Handler(mux)
                           http.ListenAndServe(":3004", handler)
                           
                       }




                       func  main_old() {
                       	fmt.Println("inside main")
	/*memcacheServers := []string{"http://localhost:3000",
                           "http://localhost:3001",
                           "http://localhost:3002"}*/

    /*memcacheServers := []string{"http://54.175.28.88:3030",
                           "http://52.91.171.117:3030",
                           "http://54.84.66.145:3030"}*/

                           memcacheServers := []string{"http://localhost:3004",
                           "http://localhost:3004",
                           "http://localhost:3004"}


                           ring = New(memcacheServers)

   //ring = ring.AddNode("http://localhost:3002")

                           for i := range ring.nodes {
                           	fmt.Println(i)
                           }


                           key := []Response{{Key:"1",Value:"a"},
                           {Key:"2",Value:"b"},
                           {Key:"3",Value:"c"},
                           {Key:"4",Value:"d"},
                           {Key:"5",Value:"e"},
                           {Key:"6",Value:"f"},
                           {Key:"7",Value:"g"},
                           {Key:"8",Value:"h"},
                           {Key:"9",Value:"i"},
                           {Key:"10",Value:"j"}}

   		/*putkey(key[0])
                           getkey(key[0].key)*/

                           putkey(key[0])
                           putkey(key[1])
                           putkey(key[2])
                           putkey(key[3])
                           putkey(key[4])
                           putkey(key[5])
                           putkey(key[6])
                           putkey(key[7])
                           putkey(key[8])
                           putkey(key[9])


                           var k []Response

   //Printing all caches values
                           for i,_:=range key{
                           	x,y,z,_:=getkey(key[i].Key)
                           	if(z==nil){
                           		if(y==true){
                           			k = append(k,x)
                           		}
                           	}
                           }

                           fmt.Println("===========")

   /*//removing a node
   ring = ring.RemoveNode("http://localhost:3000")

   for i,_:=range key{
       x,y,z:=getkey(key[i].key)
       if(z==nil){
           if(y==true){
               k = append(k,x)
           }
       }
   }

   fmt.Println("Removing a node:")

   getkey(key[1].key)

	

	  ring = ring.AddNode("http://localhost:3000") 
	  fmt.Println("Adding a node:")

	  for i,_:=range key{
       x,y,z:=getkey(key[i].key)
       if(z==nil){
           if(y==true){
               k = append(k,x)
           }
       }
       }*/

   }