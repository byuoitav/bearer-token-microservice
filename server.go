package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jessemillar/health"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/robfig/cron"
)

type token struct {
	Token string `json:"token"`
}

func main() {
	port := ":12000"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	router.GET("/health", echo.WrapHandler(http.HandlerFunc(health.Check)))

	job := cron.New()

	job.AddFunc("0 0 0 * * *", func() { // Generate a new token every night at midnight
		log.Printf("Running scheduled token generation")

		token, err := generateKey()
		if err != nil {
			log.Printf("ERROR: " + err.Error())
			return
		}

		log.Printf("Dunking token into S3 bucket")
		err = dunk(token)
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

// generateKey...generates...a new token...um...duh...
func generateKey() (token, error) {
	bytes := make([]byte, 1024)

	_, err := rand.Read(bytes)
	if err != nil {
		return token{}, err
	}

	return token{base64.URLEncoding.EncodeToString(bytes)}, nil
}

// dunk takes the new token and puts (dunks) it into the Amazon S3 bucket; three points, field goal, touchdown!
func dunk(payload token) error {

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	svc := s3.New(session.New(), &aws.Config{Region: aws.String("us-west-2")})
	response, err := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("elasticbeanstalk-us-west-2-194925301021"),
		Key:    aws.String("bearer-token.json"),
		Body:   bytes.NewReader(payloadBytes),
	})

	if err != nil {
		return err
	}

	log.Printf("%+v", response.GoString())
	/*
		frickinArguments := []string{"s3", "cp", filename, os.Getenv("AWS_BUCKET_ADDRESS")}
		err = exec.Command("aws", frickinArguments...).Run()
		if err != nil {
			return err
		}
	*/
	return nil
}
