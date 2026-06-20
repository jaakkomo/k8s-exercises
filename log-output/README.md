# log-output

## Startup

The program can be started in `writer` or `reader` mode. Writer mode writes to `LOG_FILE` while reader mode reads from `LOG_FILE` and `PONG_FILE` and serves them at `PORT`.

``` shell
ROLE=<"writer"/"reader"> LOG_FILE=<file> PONG_FILE=<file> PORT=<port> go run .
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
