package main

import (
	"errors"
	"fmt"
	goskel "goskel/http"
	"goskel/models"
	"goskel/service"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/getsentry/sentry-go"
	"github.com/jasonlvhit/gocron"
	"github.com/jessevdk/go-flags"
	"rsc.io/quote"
)

var (
	BuildTime  string
	CommitHash string
)

func main() {

	color.Cyan(quote.Go())
	var opts models.Opts
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(1)
	}

	opts.BuildTime = BuildTime
	opts.CommitHash = CommitHash

	log.Printf("Options: %#v\n", opts)

	// Enable line numbers in logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// get from env
	sentryDSN := os.Getenv("SENTRY_DSN")

	// init sentry
	sentry.Init(sentry.ClientOptions{
		Dsn:         sentryDSN,
		Environment: opts.Mode,
	})

	// Since sentry emits events in the background we need to make sure
	// they are sent before we shut down
	sentry.Flush(time.Second * 5)

	// initialize http
	httpServer := goskel.NewServer()
	srv, err := service.NewService(&opts)

	if err != nil {
		log.Fatal(err)
	}

	httpServer.Service = srv

	go scheduleJobs(err, srv)

	log.Fatal(http.ListenAndServe(":8080", httpServer))
}

// task is the function to be executed by cron job
func task(srv *service.Service, configurationUuid string) {

	fmt.Printf("Running export for configuration: %s\n", configurationUuid)
	filename, checksum, err := srv.S3ExportJob(configurationUuid)
	if err != nil {
		message := fmt.Sprintf("Problem running export for configuration %s: %s", err, configurationUuid)
		color.Red(message)
		sentry.CaptureException(errors.New(message))
	}
	color.Yellow("Exported file: %s, checksum: %s\n", filename, checksum)
}

// scheduleJobs schedules the cron jobs
func scheduleJobs(err error, srv *service.Service) {
	// read jobs
	log.Println("Lookup jobs from db...")
	cs, err := srv.GetJobs()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Found jobs in total:", len(cs))
	for _, c := range cs {
		if c.ScheduleTime == "" { // skip jobs without scheduleTime
			continue
		}
		color.Yellow("Scheduling export for configuration %s at %s\n", c.Uuid, c.ScheduleTime)
		gocron.Every(1).Day().At(c.ScheduleTime).Do(task, srv, c.Uuid)
	}

	// start cron job
	<-gocron.Start()
}
