# wikipedia

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
