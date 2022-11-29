package test

import (
	. "goskel/models"
	. "goskel/service"
	"log"
	"net/http/httptest"
	"os"
	"testing"
)

func DefaultOpts() Opts {
	mongoUri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("DB_NAME")
	adminToken := os.Getenv("ADMIN_TOKEN")

	if mongoUri == "" {
		mongoUri = "mongodb://localhost:27017"
	}

	if dbName == "" {
		dbName = "goskel-test"
	}

	if adminToken == "" {
		adminToken = "supersecretadmintoken"
	}

	return Opts{MongoURI: mongoUri, DBName: dbName, AdminToken: adminToken}

}

func setup(t *testing.T) (*Service, func()) {
	opts := DefaultOpts()

	log.Printf("Options: %#v\n", opts)

	srv, err := NewService(&opts)

	// Enable line numbers in logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if err != nil {
		t.Errorf("Service connection error: %s", err)
		return nil, func() {}
	}
	return srv, func() {
		if err := srv.Cleanup(); err != nil {
			t.Errorf("Service.Close: %s", err)
		}
	}
}

func debugResponse(w *httptest.ResponseRecorder) {
	log.Println("=========> Response:")
	log.Println("Status code:", w.Code)
	log.Println("Headers:")
	for k, v := range w.Header() {
		log.Printf("%s:%s\n", k, v)
	}
	log.Println("Body:", w.Body.String())
}
