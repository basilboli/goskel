package service

import (
	"context"
	"fmt"
	"goskel/models"
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/fatih/color"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	S3_REGION = "eu-west-1"
	S3_BUCKET = "goskel"
)

type Service struct {
	Opts  *models.Opts
	mongo *mongo.Client
}

func NewService(opts *models.Opts) (*Service, error) {

	// Set Mongo options
	clientOptions := options.Client().ApplyURI(opts.MongoURI)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		return nil, err
	}
	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Printf("ERR. Cannot connect to Database. Check your connection details: %s\n", opts.MongoURI)
		return nil, err
	}
	color.Yellow("Successfully connected to MongoDB!")

	return &Service{mongo: client, Opts: opts}, nil
}

func (s *Service) GetJobs() ([]*models.Job, error) {
	collection := s.mongo.Database(s.Opts.DBName).Collection("jobs")

	// Pass these options to the Find method
	findOptions := options.Find()
	findOptions.SetLimit(10)

	// Here's an array in which you can store the decoded documents
	results := make([]*models.Job, 0)

	// Passing bson.D{{}} as the filter matches all documents in the collection
	cur, err := collection.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		return nil, err
	}

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var elem models.Job
		err := cur.Decode(&elem)
		if err != nil {
			return nil, err
		}

		results = append(results, &elem)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	// Close the cursor once finished
	err = cur.Close(context.TODO())
	if err != nil {
		log.Println("[WARN] Problem closing cursor")
	}
	return results, nil
}

func (s *Service) GetLog(configurationUuid string, fileName string) (*models.Job, error) {
	log.Printf("Finding log for configuration: %s (database %s) \n", configurationUuid, s.Opts.DBName)
	collection := s.mongo.Database(s.Opts.DBName).Collection("logs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var config *models.Job

	err := collection.FindOne(ctx, bson.D{{Key: "uuid", Value: configurationUuid}, {Key: "fileName", Value: fileName}}).Decode(&config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (s *Service) GetConfiguration(configurationUuid string) (*models.Job, error) {
	log.Printf("Finding configuration for customer: %s (database %s) \n", configurationUuid, s.Opts.DBName)
	collection := s.mongo.Database(s.Opts.DBName).Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var config *models.Job

	err := collection.FindOne(ctx, bson.D{{Key: "uuid", Value: configurationUuid}}).Decode(&config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// S3ExportJob exports data to Amazon S3
func (s *Service) S3ExportJob(configurationUuid string) (string, string, error) {

	// read configuration for the job
	conf, err := s.GetConfiguration(configurationUuid)
	if err != nil {
		return "", "", err
	}
	log.Printf("Found configuration: %#v\n", conf)

	// create sample data csv
	fileName := fmt.Sprintf("file-%s.csv", time.Now().Format("2006-01-02"))
	filePath := fmt.Sprintf("/tmp/%s", fileName)
	log.Printf("Creating file: %s\n", filePath)
	_, err = os.Create(filePath)
	if err != nil {
		return "", "", err
	}

	// calculate md5 hash
	checksum, err := md5hash(filePath)
	if err != nil {
		return "", "", err

	}

	// upload file to s3
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(S3_REGION),
		Credentials: credentials.NewStaticCredentials(
			s.Opts.AwsAccessKeyID,
			s.Opts.AwsSecretAccessKey,
			""),
	}))

	s3file, err := uploadToS3(sess, S3_BUCKET, fileName)
	if err != nil {
		sentry.CaptureException(err)
	}

	// write to log (timestamp, link to uploaded s3 file)
	logEntry := &models.LogEntry{
		Type:        "job",
		CreatedDate: time.Now(),
		Uuid:        conf.Uuid,
		FileName:    fileName,
		CheckSum:    checksum,
		S3File:      s3file,
	}

	// write log to database
	err = s.NewLogEntry(logEntry)
	if err != nil {
		sentry.CaptureException(err)
	}

	return fileName, checksum, nil
}

func (s *Service) NewLogEntry(logEntry *models.LogEntry) error {
	collection := s.mongo.Database(s.Opts.DBName).Collection("logs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, logEntry)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Cleanup() error {
	// cleanup data (used in testing)
	return nil
}
