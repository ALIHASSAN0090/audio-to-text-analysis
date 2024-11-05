package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		log.Fatal("AWS_REGION environment variable is not set")
	}

	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if accessKeyID == "" || secretAccessKey == "" {
		log.Fatal("AWS_ACCESS_KEY_ID or AWS_SECRET_ACCESS_KEY environment variable is not set")
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
		Credentials: credentials.NewStaticCredentials(
			accessKeyID,     // Access key ID
			secretAccessKey, // Secret access key
			""),             // Token (optional)
	})
	if err != nil {
		log.Fatalf("Failed to create AWS session: %v", err)
	}

	bucketName := os.Getenv("S3_BUCKET")
	audioFileName := os.Getenv("AUDIO_FILE")

	if bucketName == "" || audioFileName == "" {
		log.Fatal("One or more required environment variables are not set")
	}

	svc := s3.New(sess)
	_, err = svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(audioFileName),
	})
	if err != nil {
		log.Fatalf("Failed to find file: %v", err)
	}

	fmt.Printf("File %s is accessible in the bucket %s\n", audioFileName, bucketName)
}
