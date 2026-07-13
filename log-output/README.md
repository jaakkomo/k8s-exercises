# log-output

## Startup

The program can be started in `writer` or `reader` mode. Writer mode writes to `LOG_FILE` while reader mode reads from `LOG_FILE`, `FILE`, `MESSAGE` and `PINGS_API` and serves them at `PORT`.

``` shell
ROLE=<"writer"/"reader"> LOG_FILE=<file> FILE=<file> MESSAGE=<text> PINGS_API=<link> GREETER_API=<link> PORT=<port> go run .
```

## Deployment

Ensure namespace `exercises` exists:

``` shell
kubectl create namespace exercises
```

Enable Istio on the namespace:

``` shell
kubectl label namespace exercises istio.io/dataplane-mode=ambient
```

Create an Istio Waypoint:

``` shell
istioctl waypoint apply -n exercises --enroll-namespace
```

Deploy:

``` shell
kubectl apply -k .
```

Go then deploy `../ping-pong` and `../greeter`.
