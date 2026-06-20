# ping-pong

## Startup

``` shell
PORT=<port> FILE=<file> go run .
```

## Deployment

Make sure `/tmp/kube` exists on `k3d-k3s-default-agent-0`:

``` shell
docker exec k3d-k3s-default-agent-0 mkdir /tmp/kube
```

``` shell
kubectl apply -f manifests
kubectl apply -f ../shared
```
