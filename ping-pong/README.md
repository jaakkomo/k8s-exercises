# ping-pong

## Startup

``` shell
PORT=<port> DB_USER=<username> DB_PASSWORD=<password> DB_HOST=<hostname> DB_PORT=<port> go run .
```

## Deployment

Ensure namespace `exercises` exists:

``` shell
kubectl create namespace exercises
```

Create shared resources:

``` shell
kubectl apply -k ../shared
```

Deploy:

``` shell
kubectl apply -k .
```
