# ping-pong

## Startup

``` shell
PORT=<port> go run .
```

## Deployment

``` shell
kubectl apply -f ../shared/ns-exercises.yaml
kubectl apply -f ../shared
kubectl apply -f manifests
```
