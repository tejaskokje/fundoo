# Rundoo Catalog Service

This repo contains the code for Rundoo Catalog Service. The catalog maintains records for `SKU`, `Name` and `Category` of a given product. 
This service is implemented using `Go` as the backend service and simple `HTML`/`JavaScript` at the frontend.

The service listens on the TCP port 8080 by default on all the IP addresses on the server. You can set `SVC_LISTEN_ADDR`  and `SVC_LISTEN_PORT` to change the listening IP addresses and TCP port.

## Endpoints

This service currently provides two endpoints

- *_/rundoo/product.Catalog/Create_* Endpoint: The `Create` endpoint allows creation of a record for a given product into the catalog service. The endpoint uses `POST` method to get inputs for this record. Below is an example of how you can invoke this endpoint using `curl` utility.

```
curl --request POST --url http://rundoo.kokje.me:8080/rundoo/product.Catalog/Create --header 'Content-Type: application/json' --data '{
  "sku": "foo",
  "name": "bar",
  "category": "baz"
}'
```

**If the `SKU` already exist, this endpoint will return an error.**

- *_/rundoo/product.Catalog/Search_*  Endpoint: The `Search` endpoint allows for searching the catalog using any arbitrary search term. It searches across all attributes of a given product (`SKU`, `Name` & `Category`). This endpoint usese `POST` method to get the search term.
Below is an example of how you can invoke this endpoint using `curl` utility.

```curl --request POST --url http://0.0.0.0:8080/rundoo/product.Catalog/Search --header 'Content-Type: application/json' --data '{
    "query": "foo"
}'
```

**This endpoint will return an error if no products are found for a given search term. The search is not case sensitive.**

## Database

The design for this service allows for using any database of your choice. You just have to implement the `Catalog` interface for a given database.
The current implementation uses `MySQL` as the backend database. 

Database configuration should be provided to the service using environment variable. Below are the environment variables that are needed

- `DB_USER`: MySQL database user name.
- `DB_PASSWORD`: MySQL database password for the `DB_USER`.
- `DB_LOCATION`: This is the hostname or IP address of the server that hosts `MySQL`. Optional port number can be appended after a `:` if using non default MySQL port. For example, `DB_LOCATION` can have value `rundoo.kokje.me:3306`.
- `DB_NAME`: This is the database name that will be used by the Catalog service.

Below are the SQL commands used to create the database. It is also highly recommended to create indexes on various columns for efficient searching.

```
CREATE TABLE catalog ( SKU varchar(12) NOT NULL, Name varchar(255) NOT NULL, Category varchar(255) NOT NULL PRIMARY KEY (SKU));
CREATE UNIQUE INDEX idx_sku on catalog (sku);
CREATE UNIQUE INDEX idx_name on catalog (name);
CREATE INDEX idx_category on catalog (category);
```
## Frontend

The frontend is implemented using `HTML` and `JavaScript`. This can be hosted on any webserver. The default backend service is https://rundoo.kokje.me, but you can change it using `API Server` text box.

## Building The Service

You can build the service using `Go` toolchain. Use the following command the build

```go build -o catalog cmd/main.go```

This will create a binary called `catalog` in the current working directory. To run the service, use the following command

```DB_USER=<db_user> DB_PASSWORD=<db_password> DB_LOCATION=localhost DB_NAME=rundoo ./catalog```