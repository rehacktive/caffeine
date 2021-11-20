# caffeine - minimum viable backend
A very basic REST service for JSON data - enough for prototyping and MVPs!

Features:
- no need to set up a database, all data is managed automagically*
- REST paradigm CRUD for multiple entities/namespaces
- schema validation
- search using jq like syntax (see https://stedolan.github.io/jq/manual/)
- CORS enabled
- easy to deploy as container


Currently supports:
  - in memory database
  - postgres
  - filesystem storage

For a sample Vue app using caffeine see: https://gist.github.com/calogxro/6e601e07c2a937df4418d104fb717570

## How to

Simply start the server with:

```go 
go run caffeine.go
```
optional params are:

```sh
Usage of caffeine:
  -DB_TYPE="memory": db type to use, options: memory | postgres | fs
  -FS_ROOT="./data": path of the file storage root
  -IP_PORT=":8000": ip:port to expose
  -PG_HOST="0.0.0.0": postgres host (port is 5432)
  -PG_PASS="": postgres password
  -PG_USER="": postgres user
```

Store a new "user" with an ID and some json data:

```sh
> curl -X POST -d '{"name":"jack","age":25}'  http://localhost:8000/ns/users/1
{"name":"jack","age":25}
```

the value will be validated, but it could be anything (in JSON!)

retrieve later with:

```sh
> curl http://localhost:8000/ns/users/1
{"name":"jack","age":25}
```

## All operations

Insert/update
```sh
> curl -X POST -d '{"name":"jack","age":25}'  http://localhost:8000/ns/users/1
{"name":"jack","age":25}
```

Delete
```sh
> curl -X DELETE http://localhost:8000/ns/users/1
```

Get by ID
```sh
> curl http://localhost:8000/ns/users/1
{"name":"jack","age":25}
```

Get all values for a namespace
```sh
> curl http://localhost:8000/ns/users | jq 
[
  {
    "key": "2",
    "value": {
      "age": 25,
      "name": "john"
    }
  },
  {
    "key": "1",
    "value": {
      "age": 25,
      "name": "jack"
    }
  }
]
```

Get all namespaces
```sh
> curl http://localhost:8000/ns
["users"]
```

Delete a namespace
```sh
> curl -X DELETE http://localhost:8000/ns/users
{}
```

Search by property (jq syntax)
```sh
> curl http://localhost:8000/search/users?filter="select(.name==\"jack\")"  | jq
{
  "results": [
    {
      "key": "1",
      "value": {
        "age": 25,
        "name": "jack"
      }
    }
  ]
}
```

## Schema Validation

You can add a schema for a specific namespace, and only correct JSON data will be accepted

To add a schema for the namespace "user", use the one available in schema_sample/:

```sh
curl --data-binary @./schema_sample/user_schema.json http://localhost:8000/schema/user
```

Now only validated "users" will be accepted (see user.json and invalid_user.json under schema_sample/)


## Run as container

```sh
docker build -t caffeine .
```
and then run it:
```sh
docker run --publish 8000:8000 caffeine
```

## Run with Postgres

First run an instance of Postgres (for example with docker):

```sh
docker run -e POSTGRES_USER=caffeine -e POSTGRES_PASSWORD=mysecretpassword -p 5432:5432 -d postgres:latest
```

Then run caffeine with the right params to connect to the db:

```sh
DB_TYPE=postgres PG_HOST=0.0.0.0 PG_USER=caffeine PG_PASS=mysecretpassword go run caffeine.go
```

(params can be passed as ENV variables or as command-line ones)

A very quick to run both on docker with docker-compose:

```sh
docker-compose up -d
```