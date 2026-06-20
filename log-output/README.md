# log-output

## Startup

The program can be started in `writer` or `reader` mode. Writer mode writes to `FILE` while reader mode reads from `FILE` and serves it at `PORT`.

``` shell
ROLE=<"writer"/"reader"> FILE=<file> PORT=<port> go run .
```

## Deployment

``` shell
kubectl apply -f manifests
kubectl apply -f ../shared
```
