package catalog

import (
	"context"
	"database/sql"
	"fmt"

	pb "rundoo.com/pkg/proto"
)

type Options struct {
	dbName     string
	dbLocation string
	dbUserName string
	dbPassword string
}

type OptionsFunc func(option *Options)

func WithDBName(dbName string) OptionsFunc {
	return func(option *Options) {
		option.dbName = dbName
	}
}

func WithDBLocation(dbLocation string) OptionsFunc {
	return func(option *Options) {
		option.dbLocation = dbLocation
	}
}

func WithDBUserName(dbUserName string) OptionsFunc {
	return func(option *Options) {
		option.dbUserName = dbUserName
	}
}

func WithDBPassword(dbPassword string) OptionsFunc {
	return func(option *Options) {
		option.dbPassword = dbPassword
	}
}

type ServerWithMysql struct {
	openDBConnFunc func() (*sql.DB, error)
}

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

func (s *ServerWithMysql) Create(ctx context.Context, req *pb.CreateReq) (*pb.CreateResp, error) {
	if req.Sku == "" || req.Category == "" || req.Name == "" {
		return nil, ErrInvalidProduct
	}

	sqlStatement := `INSERT INTO catalog (sku, name, category) VALUES (?, ?, ?);`

	db, err := s.openDBConnFunc()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	_, err = db.Exec(sqlStatement, req.Sku, req.Name, req.Category)
	if err != nil {
		return nil, err
	}
	return &pb.CreateResp{}, nil
}

func (s *ServerWithMysql) Search(context context.Context, req *pb.SearchReq) (*pb.SearchResp, error) {

	db, err := s.openDBConnFunc()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	sqlStatement := "SELECT sku, name, category FROM catalog WHERE sku=? OR name=? OR category=?;"
	rows, err := db.Query(sqlStatement, req.Query, req.Query, req.Query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var products []*pb.Product
	var sku, name, category string
	for rows.Next() {
		err = rows.Scan(&sku, &name, &category)
		if err != nil {
			return nil, err
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
		return nil, err
	}

	return &pb.SearchResp{
		Result: products,
	}, nil
}
