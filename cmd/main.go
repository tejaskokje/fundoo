package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/cors"
	"github.com/twitchtv/twirp"
	"rundoo.com/pkg/catalog"
	pb "rundoo.com/pkg/proto"
)

func main() {

	address := os.Getenv("SVC_LISTEN_ADDR")
	if address == "" {
		address = ":"
	}
	port := os.Getenv("SVC_LISTEN_PORT")
	if port == "" {
		port = "8080"
	}

	// get database configuration from the environment
	// variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbLocation := os.Getenv("DB_LOCATION")
	dbName := os.Getenv("DB_NAME")

	server := catalog.NewServerWithMysql(
		catalog.WithDBName(dbName),
		catalog.WithDBLocation(dbLocation),
		catalog.WithDBUserName(dbUser),
		catalog.WithDBPassword(dbPassword),
	)

	// create a new server
	twirpServer := pb.NewCatalogServer(server,
		twirp.WithServerPathPrefix("/rundoo"))

	// make a CORS wrapper:
	corsWrapper := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // not safe, ok for POC
		AllowedMethods: []string{"POST"},
		AllowedHeaders: []string{"Content-Type"},
	})

	twirpHandler := corsWrapper.Handler(twirpServer)

	mux := http.NewServeMux()
	mux.Handle(twirpServer.PathPrefix(), twirpHandler)
	log.Println("Listening on http://" + address + ":" + port + twirpServer.PathPrefix())
	err := http.ListenAndServe(address+port, mux)
	if err != nil {
		log.Fatal(err)
		return
	}
}
