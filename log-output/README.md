# log-output

## Startup

The program can be started in `writer` or `reader` mode. Writer mode writes to `LOG_FILE` while reader mode reads from `LOG_FILE` and `PINGS_API` and serves them at `PORT`.

``` shell
ROLE=<"writer"/"reader"> LOG_FILE=<file> PINGS_API=<link> PORT=<port> go run .
```

## Deployment

``` shell
kubectl apply -f ../shared/ns-exercises.yaml
kubectl apply -f ../shared
kubectl apply -f manifests
```
