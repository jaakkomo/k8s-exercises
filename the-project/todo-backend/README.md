# todo-backend

## Startup

``` shell
PORT=<port> NATS_URL=<nats api> DB_USER=<username> DB_PASSWORD=<password> DB_HOST=<hostname> DB_PORT=<port> go run .
```

## Deployment

You probably want to deploy the whole stack, see `../README.md`. Here is deployment for `todo-backend` only:

Ensure namespace `project` exists:

``` shell
kubectl create namespace project
```

Deploy:

``` shell
kubectl apply -k .
```
