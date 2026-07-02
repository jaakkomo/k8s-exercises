# todo-app

## Startup

``` shell
PORT=<port> PICTURE=<file> PICTURE_API=<link> CACHE_INTERVAL=<duration> TODOS_API=<link> GIN_MODE=<"release"/"debug"> go run .
```

## Deployment

You probably want to deploy the whole stack, see `../README.md`. Here is deployment for `todo-app` only:

Ensure namespace `project` exists:

``` shell
kubectl create namespace project
```

Deploy:

``` shell
kubectl apply -k .
```
