## How the authentication works in Caffeine


If the right env variable is passed like:

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

The public pem **needs** to be stored in certs/public-cert.pem

* generate a valid JWT token:

go to https://jwt.io and paste your private and public keys (no worries, it works offline) and add a user id in the payload:

```json
{
  "name": "John Doe",
  "jti": "johnd",
  "iat": 1516239022
}
```

copy the encoded value generated.

Now you're ready to interact with Caffeine. For example, to store a new value, the call should be something like:


```sh
curl -H "Authorization: Bearer YOUR_TOKEN" -X POST -d '{"name":"john"}' http://localhost:8000/ns/test/1
```


