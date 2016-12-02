package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/jessemillar/health"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/robfig/cron"
)

type key struct {
	Key string `json:"key"`
}

func main() {
	port := ":12000"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	router.GET("/health", echo.WrapHandler(http.HandlerFunc(health.Check)))

	job := cron.New()

	// job.AddFunc("0 0 0 * * *", func() { // Generate a new key every night at midnight
	job.AddFunc("* * * * * *", func() { // Generate a new key every night at midnight
		log.Printf("Running scheduled key generation")

		key, err := generateKey()
		if err != nil {
			log.Printf("ERROR: " + err.Error())
			return
		}

		log.Printf("Dunking key into S3 bucket")
		err = dunk(key)
		if err != nil {
			log.Printf("ERROR: " + err.Error())
			return
		}

		log.Printf("GET DUNKED ON")
	})

	job.Start()

	log.Printf("Job scheduled")

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	router.StartServer(&server)
}

// generateKey...generates...a new key...um...duh...
func generateKey() (key, error) {
	bytes := make([]byte, 1024)

	_, err := rand.Read(bytes)
	if err != nil {
		return key{}, err
	}

	return key{base64.URLEncoding.EncodeToString(bytes)}, nil
}

// dunk takes the new key and puts (dunks) it into the Amazon S3 bucket; three points, field goal, touchdown!
func dunk(payload key) error {
	filename := "bearer-token.json"

	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, bytes, 0644)
	if err != nil {
		return err
	}

	frickinArguments := []string{"s3", "cp", filename, os.Getenv("AWS_BUCKET_ADDRESS")}
	err = exec.Command("aws", frickinArguments...).Run()
	if err != nil {
		return err
	}

	return nil
}
