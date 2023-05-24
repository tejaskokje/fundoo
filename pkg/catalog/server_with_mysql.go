package catalog

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	pb "fundoo.com/pkg/proto"
	my "github.com/go-mysql/errors"
	"github.com/twitchtv/twirp"
)

// Options struct takes various options for creating
// DB connection.
type Options struct {
	dbName     string
	dbLocation string
	dbUserName string
	dbPassword string
}

type OptionsFunc func(option *Options)

// WithDBName sets the database name.
func WithDBName(dbName string) OptionsFunc {
	return func(option *Options) {
		option.dbName = dbName
	}
}

// WithDBLocation sets the database location. It
// is a combination of hostname:port. Port can be
// omitted if using default 3306.
func WithDBLocation(dbLocation string) OptionsFunc {
	return func(option *Options) {
		option.dbLocation = dbLocation
	}
}

// WithDBUserName sets the database user name.
func WithDBUserName(dbUserName string) OptionsFunc {
	return func(option *Options) {
		option.dbUserName = dbUserName
	}
}

// WithDBPassword sets the database password.
func WithDBPassword(dbPassword string) OptionsFunc {
	return func(option *Options) {
		option.dbPassword = dbPassword
	}
}

// ServerWithMySql implements Catalog interface.
type ServerWithMysql struct {
	openDBConnFunc func() (*sql.DB, error)
}

// NewServerWithMysql returns a server implementing Catalog
// interface using MySql as a backend.
func NewServerWithMysql(options ...OptionsFunc) *ServerWithMysql {
	var opts Options
	for _, apply := range options {
		apply(&opts)
	}

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s", opts.dbUserName,
		opts.dbPassword, opts.dbLocation, opts.dbName)

	return &ServerWithMysql{
		openDBConnFunc: func() (*sql.DB, error) {
			return sql.Open("mysql", dataSourceName)

		},
	}
}

// Create function inserts a record of (sku, name, category) into the MySql database
func (s *ServerWithMysql) Create(ctx context.Context, req *pb.CreateReq) (*pb.CreateResp, error) {

	// return error if any of sku, category or name is empty. All three fields are
	// required.
	if req.GetSku() == "" || req.GetCategory() == "" || req.GetName() == "" {
		return nil, ErrSkuNameCategoryRequired
	}

	log.Printf("create req received with sku: %s name: %s category: %s", req.Sku, req.Name, req.Category)
	sqlStatement := `INSERT INTO catalog (sku, name, category) VALUES (?, ?, ?);`

	// open DB connection
	db, err := s.openDBConnFunc()
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	defer db.Close()

	// execute the query
	_, err = db.Exec(sqlStatement, req.Sku, req.Name, req.Category)
	if err != nil {
		if ok, myerr := my.Error(err); ok { // MySQL error
			if myerr == my.ErrDupeKey {
				return nil, ErrProductAlreadyExists
			}

		}
		return nil, twirp.InternalErrorWith(err)
	}

	return &pb.CreateResp{}, nil
}

// Search function searches across the MySQL database using a string in the request field.
// It searches across sku, name and category.
func (s *ServerWithMysql) Search(context context.Context, req *pb.SearchReq) (*pb.SearchResp, error) {

	if req.GetQuery() == "" {
		return nil, ErrSearchQueryRequired
	}
	log.Printf("search req received for: %s", req.GetQuery())

	// open a database connection
	db, err := s.openDBConnFunc()
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	defer db.Close()
	query := req.GetQuery()

	sqlStatement := "SELECT sku, name, category FROM catalog WHERE sku=? OR name=? OR category=?;"

	// Query the database
	rows, err := db.Query(sqlStatement, query, query, query)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	defer rows.Close()

	// scan the results and create a product array
	var products []*pb.Product
	var sku, name, category string
	for rows.Next() {
		err = rows.Scan(&sku, &name, &category)
		if err != nil {
			return nil, twirp.InternalErrorWith(err)
		}
		products = append(products, &pb.Product{
			Name:     name,
			Sku:      sku,
			Category: category,
		})
	}

	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	// if no product was found, return an error
	if len(products) == 0 {
		return nil, ErrNoResultFound
	}

	return &pb.SearchResp{
		Result: products,
	}, nil
}
