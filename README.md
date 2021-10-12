# HTTP to Kafka

Connector to push messages received from HTTP to a Kafka cluster. This service supports both HTTP and HTTPS.

Execute with the flag `-h` to list the configuration options.

The source code is distributed under the [Apache License Version 2.0](./LICENSE).

## Install the service locally

```
go get -v -t -d ./...
go build -v ./...
go install -v ./...
```

## Build the docker image

```
docker build . -t aerisconsulting/http-to-kafka && docker push aerisconsulting/http-to-kafka
```

## Use from docker

You can find an example in the file [docker-compose.yml](./docker-compose.yml).

To pull and see the configuration options, run the following:

```
> docker pull aerisconsulting/http-to-kafka && docker run -it --rm aerisconsulting/http-to-kafka -h

HTTP to Kafka is a lightweight service developed by AERIS-Consulting e.U., that acts as a connector between HTTP and Kafka.

It supports HTTP, HTTPS, sessions in memory and using Redis.
All the request to push data have to be identified using a cookie obtained from the login endpoint.

1. Post a request to the endpoint /login with a JSON payload as follows: {"username": "test", "password": "test"}
The response contains a session cookie named aeris-http-to-kafka-session, that has to be reused in further requests.
2. Post data with a request to the endpoint /data and any kind of payload. You can set the HTTP header "message-key" to specify the Kafka key to use.
3. Invalidate the sessions by executing a Delete request to the endpoint /logout.

Usage:
  http-to-kafka [flags]

Flags:
  -h, --help                          help for http-to-kafka
      --http                          enables the plain HTTP server (default true)
      --https                         enables the HTTPS server
      --kafka-bootstrap string        bootstrap for the Kafka client (default "localhost:9092")
      --kafka-configuration strings   general properties for the Kafka client, as key=value pairs
      --kafka-topic string            topic to produce the Kafka records to (default "http-request")
      --password string               password for the HTTP login (default "test")
      --plain-port int                port for plain HTTP (default 8080)
      --session-redis                 enables the HTTP session persistence in Redis
      --session-redis-auth string     auth secret for the Redis database for the HTTP session persistence
      --session-redis-database int    index for the Redis database for the HTTP session persistence
      --session-redis-uri string      URI to connect to Redis for the HTTP session persistence (default "localhost:6379")
      --session-secret string         secret for the for session store
      --ssl-cert string               certificate file for the server
      --ssl-key string                key file for the server certificate
      --ssl-port int                  port for HTTPS (default 8443)
      --username string               username for the HTTP login (default "test")

```

## How to use the HTTP server to sign in and push data

### Sign in

Post a request to the endpoint /login with a JSON payload as follows:

```
{"username": "test", "password": "test"}
```

The response contains a session cookie to be reused in further requests to push data.

### Push data

Post a request to the endpoint /data with any kind of payload. Do not forget to add the session cookie to pass through
the security filter.

If you want to set a key on the message published to Kafka, you can set the HTTP header "message-key".

### Close the session

Execute a Delete request to the endpoint /logout to invalidate the session.