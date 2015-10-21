package main

import (
"fmt"
"encoding/json"
"net/http"
"bytes"
"strconv"
)


func main(){

    //No of masters configured
    const noOfMasters =2

    //Ip address of masters
    var ipAddress [noOfMasters]string
    ipAddress[0]="52.91.104.10:3030"
    ipAddress[1]="54.149.18.16:3030"
    //ipAddress[1]="52.91.104.10:3030"


    //Rest end point urls
    var postURLS [noOfMasters]string
    for in:=0;in<len(ipAddress);in++{
        postURLS[in]="http://"+ipAddress[in]+"/keyvals"
    }


    //Dynamically creating array of requests[Each request is a map of key value pairs]
    var reqData [4]map[string]string 

    for j:=0;j<4;j++{
        outerIndex:=strconv.Itoa(j)
        reqData[j]=make(map[string]string)
        //Populating each map entry with 2 key,value pairs
        for j1:=0;j1<2;j1++{
            reqData[j]["map"+outerIndex+strconv.Itoa(j1)]="value"+outerIndex+strconv.Itoa(j1)
        }
    }

//For printing and checking successful map creation
    fmt.Println(reqData[0]["map00"])


//Sharding the requests equally to the 2 masters
    for k:=0;k<len(reqData);k++{
        fmt.Println("")
        fmt.Println("---Request No. :",strconv.Itoa(k),"---")
        if(k%2==0){
                //Forwarding request to master 1
            fmt.Println("Forwarding request to IP: ",postURLS[0])
            //Printing request
            fmt.Println("The request is: ")
            fmt.Println(reqData[k])
            jsonString,err:=json.Marshal(reqData[k])
            resp, err := http.Post(postURLS[0], "application/json", bytes.NewBuffer(jsonString))
            if err != nil {
                panic(err)
            }
            //Printing response
            fmt.Println("The response is: ")
            fmt.Println(resp)
        }else{
                //Forwarding request to master 2
            fmt.Println("Forwarding request to IP: ",postURLS[1])
            //Printing request
            fmt.Println("The reqest is: ")
            fmt.Println(reqData[k])
            jsonString,err:=json.Marshal(reqData[k])
            resp, err := http.Post(postURLS[1], "application/json", bytes.NewBuffer(jsonString))
            if err != nil {
                panic(err)
            }
            //Printing response
            fmt.Println("The response is: ")
            fmt.Println(resp)
        }

    }

}
