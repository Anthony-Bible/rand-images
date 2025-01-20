package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"html/template"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

var (
	lastTwoImages []string
	mutex         sync.Mutex
	s3Client      *s3.S3
	bucketName    = "rand-images"
	tmpl          *template.Template
)

func init() {
	lastTwoImages = make([]string, 2)
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator
	var err error
	tmpl, err = template.New("").ParseFiles("/templates/image.html")
	if err != nil {
		fmt.Printf("Error parsing templates: %v", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}
	result, err := s3Client.ListObjectsV2(input)
	if err != nil {
		http.Error(w, "Unable to list the images in the bucket", http.StatusInternalServerError)
		return
	}

	if len(result.Contents) == 0 {
		http.Error(w, "No images found", http.StatusNotFound)
		return
	}
	var randomObject *s3.Object
	mutex.Lock()
	for {
		randomObject = result.Contents[rand.Intn(len(result.Contents))]
		if *randomObject.Key != lastTwoImages[0] && *randomObject.Key != lastTwoImages[1] {
			break
		}
	}

	// Update lastTwoImages to remember the last two served images
	lastTwoImages[0], lastTwoImages[1] = lastTwoImages[1], *randomObject.Key
	fmt.Printf("using image %s\n", *randomObject.Key)
	imageURL := fmt.Sprintf("https://photos.anthony.bible/file/%s/%s", bucketName, url.QueryEscape(*randomObject.Key))
	mutex.Unlock()
	//http.Redirect(w, r, urlStr, http.StatusTemporaryRedirect)
	redirectParam := r.URL.Query().Get("redirect")
	if redirectParam == "true" {
		http.Redirect(w, r, imageURL, http.StatusSeeOther)
		return
	}
	data := struct {
		ImageURL string
	}{
		ImageURL: imageURL,
	}

	err = tmpl.ExecuteTemplate(w, "image", data)
	if err != nil {
		fmt.Printf("Error rendering template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func main() {
	// Connect to backblaze
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(os.Getenv("IMAGE_ID"), os.Getenv("IMAGE_KEY"), ""),
		Endpoint:         aws.String("https://s3.us-west-001.backblazeb2.com"),
		Region:           aws.String("us-west-001"),
		S3ForcePathStyle: aws.Bool(true),
	}
	newSession := session.New(s3Config)
	s3Client = s3.New(newSession)
	http.HandleFunc("/", handler)

	http.ListenAndServe(":8080", nil)
}
