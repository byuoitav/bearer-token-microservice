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

	job.AddFunc("@hourly", func() {
		err := doTheThing()
		if err != nil {
			log.Println(err.Error())
		}
	})

	job.Start()

	log.Printf("Job scheduled")

	doTheThing() // Generate a new token each time we start the program

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	router.StartServer(&server)
}

func doTheThing() error {
	log.Printf("Generating token")

	token, err := generateToken()
	if err != nil {
		return err
	}

	log.Printf("Dunking token into S3 bucket")
	err = dunk(token)
	if err != nil {
		return err
	}

	log.Printf("GET DUNKED ON")

	return nil
}

// generateToken...generates...a new token...um...duh...
func generateToken() (token, error) {
	bytes := make([]byte, 512)

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
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("elasticbeanstalk-us-west-2-194925301021"),
		Key:    aws.String("bearer-token.json"),
		Body:   bytes.NewReader(payloadBytes),
	})

	if err != nil {
		return err
	}

	return nil
}
