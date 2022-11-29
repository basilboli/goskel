package service

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	jsonTimeLayout   = "2006-01-02T15:04:05Z07:00"
	exportTimeLayout = "2006-01-02 15:04:05"
)

// JSONTime is the time.Time with JSON marshal and unmarshal capability
type JSONTime struct {
	time.Time
}

// UnmarshalJSON will unmarshal using 2006-01-02T15:04:05+07:00 layout
func (t *JSONTime) UnmarshalJSON(b []byte) error {
	strInput := strings.Trim(string(b), `"`)
	parsed, err := time.Parse(jsonTimeLayout, strInput)
	if err != nil {
		return err
	}
	fmt.Printf("Parsing OK\n")
	t.Time = parsed
	return nil
}

// MarshalJSON will marshal using 2006-01-02T15:04:05+07:00 layout
func (t *JSONTime) MarshalJSON() ([]byte, error) {
	s := t.Format(jsonTimeLayout)
	return []byte(s), nil
}

func md5hash(fileName string) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func uploadToS3(s *session.Session, bucket string, filename string) (string, error) {

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(s)

	f, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file %q, %v", filename, err)
	}

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(S3_BUCKET),
		Key:    aws.String(uuid.New().String() + "-" + filename),
		Body:   f,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file, %v", err)
	}
	fmt.Printf("file uploaded to, %s\n", result.Location)
	return result.Location, nil
}
