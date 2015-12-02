package main

import(
	"net/http"
	"fmt"
	"os"
	"strconv"
	"io/ioutil"
	"github.com/julienschmidt/httprouter"
	"time"
)

var LOG_FILE_NAME="data.csv"

func writeToAFile(text string){

	f, err2 := os.OpenFile(LOG_FILE_NAME, os.O_APPEND | os.O_RDWR, 0666)
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

func GetBulkData(rw http.ResponseWriter, req *http.Request, p httprouter.Params){
	f, err2 := os.OpenFile(LOG_FILE_NAME, os.O_APPEND, 0644)
	if err2!=nil{
		fmt.Printf(err2.Error())
	}
	defer f.Close()

    for i:= 15000; i < 25000; i++ {
    	url := "http://54.164.93.210:3004/keys/data" + strconv.Itoa(i) + ".txt"
    	startTime := time.Now()
    	fmt.Println(url)
    	response, err := http.Get(url)
    	if err != nil {
        	panic(err)
    	}

    	body, err := ioutil.ReadAll(response.Body)
	    if err != nil && body != nil{
	       //return Response{}, false, err
	    }

	    endTime := time.Now()


	    timeDifference := int(endTime.Sub(startTime)/time.Millisecond)

	    writeToAFile(strconv.Itoa(timeDifference))

	    time.Sleep(1000 * time.Millisecond)
    }
}

func main(){
	mux := httprouter.New()

	//create file on set up only if doesnt exist
   	//LOG_FILE_NAME
   	if _, err := os.Stat(LOG_FILE_NAME); os.IsNotExist(err) {	
  		// path/to/whatever does not exist
		fmt.Println("log file does not exist,creating the file")
		f, err := os.OpenFile(LOG_FILE_NAME, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
		if err!=nil{
			fmt.Printf(err.Error())
		}
		defer f.Close()
	}

	writeToAFile("Response")

	mux.GET("/keys", GetBulkData)

	server := http.Server{
		Addr:        "0.0.0.0:3006",
		Handler: mux,
	}

	server.ListenAndServe()
}
