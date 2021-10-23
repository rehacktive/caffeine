# caffeine - minimum viable backend

A very basic REST service for JSON data - enough for prototyping and MVPs!

Features:
- no need to set up a database, all data in memory
- REST paradigm CRUD for multiple entities/namespaces
- search using jq like syntax (see https://stedolan.github.io/jq/manual/)
- CORS enabled
- easy to deploy as container

## How to

Simply start the server with:

```go 
go run caffeine.go
```
optionally provide -ip_port param, default is `:8000`

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

## Run as container

```sh
> docker build -t caffeine .
```
and then run it:
```sh
> docker run --publish 8000:8000 caffeine
```
