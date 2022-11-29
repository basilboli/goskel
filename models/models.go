package models

import "time"

// Opts represents options passed via command line or env vars
type Opts struct {
	MongoURI   string `long:"conn_uri" env:"MONGO_URI" description:"mongo connection URI" default:"mongodb://localhost:27017"`
	DBName     string `long:"db_name" env:"DB_NAME" description:"mongo database name"`
	Mode       string `long:"mode" env:"MODE" description:"mode" default:"local"`
	BuildTime  string `long:"build_time" env:"BUILD_TIME" description:"app build time" default:""`
	CommitHash string `long:"commit_hash" env:"COMMIT_HASH" description:"app commit hash" default:""`
	// admin token
	AdminToken string `long:"admin_token" env:"ADMIN_TOKEN" description:"admin token"`
	//aws info
	AwsAccessKeyID     string `long:"aws_access_key_id" env:"AWS_ACCESS_KEY_ID" description:"aws access key id"`
	AwsSecretAccessKey string `long:"aws_secret_access_key" env:"AWS_SECRET_ACCESS_KEY" description:"aws secret access key"`
}

type Job struct {
	Uuid         string `json:"uuid,omitempty" bson:"uuid"`
	Addr         string
	ScheduleTime string `bson:"scheduleTime"`
	User         string
	KeyPath      string `bson:"keyPath"`
	Path         string
}

type LogEntry struct {
	CreatedDate time.Time `json:"createdDate" bson:"createdDate"`
	Type        string    `json:"type" bson:"type"`
	Uuid        string    `json:"uuid,omitempty" bson:"uuid"`
	FileName    string    `json:"fileName" bson:"fileName"`
	CheckSum    string    `json:"md5"  bson:"md5"`
	S3File      string    `json:"s3File"  bson:"s3File"`
}
