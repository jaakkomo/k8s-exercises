# todo-backend

## Startup

``` shell
PORT=<port> DB_USER=<username> DB_PASSWORD=<password> DB_HOST=<hostname> DB_PORT=<port> go run .
```

## Deployment

Create namespace `project`:

``` shell
kubectl apply -f ../../shared/ns-project.yaml
```

Create secret for Postgres database:

``` shell
kubectl create secret generic todo-backend-postgres \
  --namespace project \
  --from-literal=username=postgres \
  --from-literal=password=postgres
```

Deploy app:

``` shell
kubectl apply -f ../../shared
kubectl apply -f manifests
```
