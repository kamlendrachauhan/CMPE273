package main

import (
    "bytes"
    "flag"
    "image"
    _"image/color"
    _"image/draw"
    "image/jpeg"
    "log"
    "net/http"
    "strconv"
    "strings"
    "os"
    "fmt"
    "gopkg.in/redis.v3"
    //"github.com/julienschmidt/httprouter"
)

var root = flag.String("root", ".", "file system path")

func main() {
    http.HandleFunc("/black/", blackHandler)
    http.Handle("/", http.FileServer(http.Dir(*root)))
    http.HandleFunc("/black/tag",getFromMain_Database)
    log.Println("Listening on 8080")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}

func blackHandler(w http.ResponseWriter, r *http.Request) {

    key := "black"
    e := `"` + key + `"`
    w.Header().Set("Etag", e)
   // w.Header().Set("Cache-Control", "max-age=2592000") // 30 days

    if match := r.Header.Get("If-None-Match"); match != "" {
        if strings.Contains(match, e) {
            w.WriteHeader(http.StatusNotModified)
            return
        }
    }
/*
    m := image.NewRGBA(image.Rect(0, 0, 240, 240))
    black := color.RGBA{0, 0, 0, 255}
    draw.Draw(m, m.Bounds(), &image.Uniform{black}, image.ZP, draw.Src)
*/
    reader, err := os.Open("C:\\Users\\KD\\IdeaProjects\\GoPro\\FinalProject\\images.jpeg")
    if err != nil {
        log.Fatal(err)
    }
    defer reader.Close()
    //reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
    m, format, err := image.Decode(reader)
    fmt.Println(format)
    var img image.Image = m

    //code starts here
    buf := new(bytes.Buffer)
    err1 := jpeg.Encode(buf, m, nil)

    fmt.Println(err1)
    send_s3 := buf.Bytes()

    fmt.Println(string(send_s3))

    writeTODB(string(send_s3))
    //code ends here for
    writeImage(w, &img)
}

func writeTODB(imagedata string){
    client := redis.NewClient(&redis.Options{
        Addr:     "52.91.39.197:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })
    _, err := client.Ping().Result()
    if err != nil {
        panic(err)
    }
    err1 := client.Set("kd", imagedata, 0).Err()
    if err1 != nil {
        panic(err1)
    }
}
// writeImage encodes an image in jpeg format and writes it into ResponseWriter.
func writeImage(w http.ResponseWriter, img *image.Image) {

    buffer := new(bytes.Buffer)
    if err := jpeg.Encode(buffer, *img, nil); err != nil {
        log.Println("unable to encode image.")
    }

    w.Header().Set("Content-Type", "image/jpeg")
    w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
    if _, err := w.Write(buffer.Bytes()); err != nil {
        log.Println("unable to write image.")
    }
}

func getFromMain_Database(rw http.ResponseWriter, req *http.Request) {
   // missingKey :=  p.ByName("tagname")
missingKey := "kd"
    client := redis.NewClient(&redis.Options{
        Addr:     "52.91.39.197:6379",
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

    bytesImage := []byte(val)
    img, _, _ := image.Decode(bytes.NewReader(bytesImage))

    writeImage(rw, &img)

}
