## How the authentication works in Caffeine


If auth is enabled via env variable/param:

```sh
AUTH_ENABLED=true go run caffeine.go
```

Caffeine runs in authenticated mode.

In order for the authentication to work, the following steps need to be done:

* create private and public certificates:

```sh
openssl genrsa -out certs/auth-private.pem 2048

openssl rsa -in certs/auth-private.pem -outform PEM -pubout -out certs/public-cert.pem
```

The public pem **needs** to be stored in certs/public-cert.pem - store the private certificate securely.

* generate a valid JWT token:

go to https://jwt.io and paste your private and public keys (no worries, it works offline) and add a user id (jti) in the payload:

```json
{
  "name": "John Doe",
  "jti": "johnd",
  "iat": 1516239022
}
```

copy the encoded value generated.

Altenatively, you can write your own code to manage users and generate JWT tokens for them.

Now you're ready to interact with Caffeine. For example, to store a new value, the call should be something like:


```sh
curl -H "Authorization: Bearer YOUR_TOKEN" -X POST -d '{"name":"john"}' http://localhost:8000/ns/test/1
```

when queried, the response contains also the user_id that created the content:

```sh
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8000/ns/test/1

{
  "user_id": "johnd",
  "data": {
    "name": "john"
  }
}
```

The current implementation is for authentication only, doesn't support any form of authorization (more at https://auth0.com/docs/get-started/authentication-and-authorization)