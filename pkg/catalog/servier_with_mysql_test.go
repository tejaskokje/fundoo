package catalog

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/twitchtv/twirp"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
	pb "rundoo.com/pkg/proto"
)

var (
	ErrExecFailed  = errors.New("error in exec statement")
	ErrQueryFailed = errors.New("error in query statement")
)

func TestCreate_ValidProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	server := &ServerWithMysql{
		openDBConnFunc: func() (*sql.DB, error) {
			return db, nil
		},
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO catalog (sku, name, category) VALUES")).
		WithArgs("1234", "Glossy White", "Paint").WillReturnResult(sqlmock.NewResult(1, 1))
	// Create a valid request
	req := &pb.CreateReq{
		Sku:      "1234",
		Name:     "Glossy White",
		Category: "Paint",
	}

	// Call the Create function
	resp, err := server.Create(context.Background(), req)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify the response is as expected
	if resp == nil {
		t.Error("Expected non-nil response, got nil")
	}
	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestCreate_InvalidProduct(t *testing.T) {
	// Create a server instance with a mock database
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	server := &ServerWithMysql{
		openDBConnFunc: func() (*sql.DB, error) {
			return db, nil
		},
	}
	defer db.Close()

	// Create an invalid request
	req := &pb.CreateReq{
		Sku:      "",
		Category: "electronics",
		Name:     "Test Product",
	}

	// Call the Create function
	resp, err := server.Create(context.Background(), req)
	if err != ErrSkuNameCategoryRequired {
		t.Errorf("Expected: %v, got %v", ErrSkuNameCategoryRequired, err)
	}

	// Verify the response is nil
	if resp != nil {
		t.Error("Expected nil response, got non-nil")
	}
}

func TestCreate_DatabaseFailure(t *testing.T) {
	// Create a server instance with a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	server := &ServerWithMysql{
		openDBConnFunc: func() (*sql.DB, error) {
			return db, nil
		},
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO catalog").WithArgs("1234", "Glossy White", "Paint").WillReturnError(ErrExecFailed)

	// Create a valid request
	req := &pb.CreateReq{
		Sku:      "1234",
		Name:     "Glossy White",
		Category: "Paint",
	}

	// Call the Create function
	resp, err := server.Create(context.Background(), req)
	var twerr twirp.Error
	if errors.As(err, &twerr) {
		if twerr.Code() != twirp.Internal {
			t.Errorf("Expected %v, got %v", twirp.Internal, err)
		}
	}

	// Verify the response is nil
	if resp != nil {
		t.Error("Expected nil response, got non-nil")
	}
}

func TestSearch_ValidQuery(t *testing.T) {
	// Create a server instance with a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	server := &ServerWithMysql{
		openDBConnFunc: func() (*sql.DB, error) {
			return db, nil
		},
	}

	rows := sqlmock.NewRows([]string{"sku", "name", "category"}).
		AddRow("1234", "Glossy White", "Paint").
		AddRow("5678", "Matte White", "Paint")

	mock.ExpectQuery("SELECT sku, name, category FROM catalog").WithArgs("Paint", "Paint", "Paint").WillReturnRows(rows)

	// Create a valid search request
	req := &pb.SearchReq{
		Query: "Paint",
	}

	// Call the Search function
	resp, err := server.Search(context.Background(), req)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify the response is as expected
	if resp == nil {
		t.Error("Expected non-nil response, got nil")
	}

	// Verify the result contains the expected products
	expectedProducts := []*pb.Product{
		{
			Name:     "Glossy White",
			Sku:      "1234",
			Category: "Paint",
		},
		{
			Name:     "Matte White",
			Sku:      "5678",
			Category: "Paint",
		},
	}

	if len(resp.Result) != len(expectedProducts) {
		t.Errorf("Expected %d products, got %d", len(expectedProducts), len(resp.Result))
	}
	for i, p := range resp.Result {
		expected := expectedProducts[i]
		if p.Name != expected.Name || p.Sku != expected.Sku || p.Category != expected.Category {
			t.Errorf("Expected product %d: %+v, got: %+v", i, expected, p)
		}
	}
}

func TestSearch_DatabaseQueryFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	server := &ServerWithMysql{
		openDBConnFunc: func() (*sql.DB, error) {
			return db, nil
		},
	}

	mock.ExpectQuery("SELECT sku, name, category FROM catalog").WithArgs("Paint", "Paint", "Paint").WillReturnError(ErrQueryFailed)

	// Create a valid search request
	req := &pb.SearchReq{
		Query: "Paint",
	}

	// Call the Search function
	resp, err := server.Search(context.Background(), req)
	var twerr twirp.Error
	if errors.As(err, &twerr) {
		if twerr.Code() != twirp.Internal {
			t.Errorf("Expected %v, got %v", twirp.Internal, err)
		}
	}

	// Verify the response is nil
	if resp != nil {
		t.Error("Expected nil response, got non-nil")
	}
}
