package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/twitchtv/twirp"
	"rundoo.com/pkg/catalog"
	pb "rundoo.com/pkg/proto"
)

type Product struct {
	Sku      string
	Name     string
	Category string
}

// Run the implementation in a local server
func main() {

	address := os.Getenv("SVC_LISTEN_ADDR")
	if address == "" {
		address = ":"
	}
	port := os.Getenv("SVC_LISTEN_PORT")
	if port == "" {
		port = "8080"
	}

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

	twirpHandler := pb.NewCatalogServer(server,
		twirp.WithServerPathPrefix("/rundoo"))

	mux := http.NewServeMux()
	mux.Handle(twirpHandler.PathPrefix(), twirpHandler)
	log.Println("Listening on http://0.0.0.0:" + port + twirpHandler.PathPrefix())
	err := http.ListenAndServe(address+port, mux)
	if err != nil {
		log.Fatal(err)
		return
	}
}
