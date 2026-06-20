# todo-app

## Startup

``` shell
PORT=<port> PICTURE=<file> PICTURE_API=<link> GIN_MODE=<"release"/"debug"> go run .
```

## Deployment

Make sure `/tmp/todo-app` exists on `k3d-k3s-default-agent-1`:

``` shell
docker exec k3d-k3s-default-agent-1 mkdir /tmp/todo-app
```

``` shell
kubectl apply -f ../shared
kubectl apply -f manifests
```
