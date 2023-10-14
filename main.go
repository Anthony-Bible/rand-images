package main

import (
    "fmt"
    "io/ioutil"
    "math/rand"
    "net/http"
    "os"
    "sync"
    "time"
)

var (
    lastTwoImages []string
    mutex         sync.Mutex
)

func init() {
    lastTwoImages = make([]string, 2)
    rand.Seed(time.Now().UnixNano()) // Seed the random number generator
}

func handler(w http.ResponseWriter, r *http.Request) {
    files, err := ioutil.ReadDir("./images")
    if err != nil {
        http.Error(w, "Unable to read the images directory", http.StatusInternalServerError)
        return
    }

    if len(files) == 0 {
        http.Error(w, "No images found", http.StatusNotFound)
        return
    }

    var randomFile os.FileInfo
    mutex.Lock()
    for {
        randomFile = files[rand.Intn(len(files))]
        if randomFile.Name() != lastTwoImages[0] && randomFile.Name() != lastTwoImages[1] {
            break
        }
    }

    // Update lastTwoImages to remember the last two served images
    lastTwoImages[0], lastTwoImages[1] = lastTwoImages[1], randomFile.Name()
	fmt.Printf("using image %s", randomFile.Name())
    mutex.Unlock()

    http.ServeFile(w, r, fmt.Sprintf("./images/%s", randomFile.Name()))
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}

