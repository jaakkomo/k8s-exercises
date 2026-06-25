# ping-pong

## Startup

``` shell
PORT=<port> DB_USER=<username> DB_PASSWORD=<password> DB_HOST=<hostname> DB_PORT=<port> go run .
```

## Deployment

Create namespace `exercises`:
``` shell
kubectl apply -f ../shared/ns-exercises.yaml
```

Create secret for Postgres database:

``` shell
kubectl create secret generic ping-pong-postgres \
  --namespace exercises \
  --from-literal=username=postgres \
  --from-literal=password=postgres
```

Deploy app:

``` shell
kubectl apply -f ../shared
kubectl apply -f manifests
```
