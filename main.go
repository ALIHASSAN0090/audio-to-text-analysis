package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/transcribeservice"
	"github.com/google/uuid"
)

type Transcript struct {
	Results struct {
		Transcripts []struct {
			Transcript string `json:"transcript"`
		} `json:"transcripts"`
	} `json:"results"`
}

func transcribeAudio(sess *session.Session, fileURI string) (string, string) {
	transcribeClient := transcribeservice.New(sess)

	jobName := fmt.Sprintf("transcription-job-%s", uuid.New().String())

	_, err := transcribeClient.StartTranscriptionJob(&transcribeservice.StartTranscriptionJobInput{
		TranscriptionJobName: aws.String(jobName),
		Media: &transcribeservice.Media{
			MediaFileUri: aws.String(fileURI),
		},
		MediaFormat:      aws.String("wav"),
		IdentifyLanguage: aws.Bool(true),
	})
	if err != nil {
		log.Fatalf("Failed to start transcription job, %v", err)
	}

	for {
		result, err := transcribeClient.GetTranscriptionJob(&transcribeservice.GetTranscriptionJobInput{
			TranscriptionJobName: aws.String(jobName),
		})
		if err != nil {
			log.Fatalf("Failed to get transcription job, %v", err)
		}

		status := *result.TranscriptionJob.TranscriptionJobStatus
		if status == "COMPLETED" || status == "FAILED" {
			fmt.Printf("Transcription job %s completed with status: %s\n", jobName, status)
			if status == "COMPLETED" {
				return *result.TranscriptionJob.Transcript.TranscriptFileUri, *result.TranscriptionJob.LanguageCode
			}
			break
		}
		time.Sleep(5 * time.Second)
	}
	return "", ""
}

func downloadTranscript(transcriptURI string) string {
	resp, err := http.Get(transcriptURI)
	if err != nil {
		log.Fatalf("Failed to download transcript, %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read transcript body, %v", err)
	}

	var transcript Transcript
	if err := json.Unmarshal(body, &transcript); err != nil {
		log.Fatalf("Failed to unmarshal transcript, %v", err)
	}

	if len(transcript.Results.Transcripts) > 0 {
		return transcript.Results.Transcripts[0].Transcript
	}

	return ""
}

func main() {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		log.Fatal("AWS_REGION environment variable is not set")
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	bucketName := os.Getenv("S3_BUCKET")
	audioFileName := os.Getenv("AUDIO_FILE")

	if bucketName == "" || audioFileName == "" {
		log.Fatal("One or more required environment variables are not set")
	}

	fileURI := fmt.Sprintf("s3://%s/%s", bucketName, audioFileName)

	transcriptURI, detectedLanguage := transcribeAudio(sess, fileURI)
	fmt.Printf("Transcription completed. Detected language: %s\n", detectedLanguage)

	transcriptText := downloadTranscript(transcriptURI)
	fmt.Printf("Transcript: %s\n", transcriptText)
}
