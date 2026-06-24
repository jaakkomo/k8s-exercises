# log-output

## Startup

The program can be started in `writer` or `reader` mode. Writer mode writes to `LOG_FILE` while reader mode reads from `LOG_FILE`, `FILE`, `MESSAGE` and `PINGS_API` and serves them at `PORT`.

``` shell
ROLE=<"writer"/"reader"> LOG_FILE=<file> FILE=<file> MESSAGE=<text> PINGS_API=<link> PORT=<port> go run .
```

## Deployment

``` shell
kubectl apply -f ../shared/ns-exercises.yaml
kubectl apply -f ../shared
kubectl apply -f manifests
```
