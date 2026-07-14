# knative exercise log

## Deploying a Knative Service

Create cluster:

``` shell
k3d cluster create --port 8082:30080@agent:0 -p 8081:80@loadbalancer --agents 2 --k3s-arg "--disable=traefik@server:0" --image rancher/k3s:v1.34.1-k3s1
```

Install knative:

``` shell
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.22.1/serving-crds.yaml
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.22.1/serving-core.yaml
kubectl apply -f https://github.com/knative-extensions/net-kourier/releases/download/knative-v1.22.1/kourier.yaml
kubectl patch configmap/config-network \
--namespace knative-serving \
--type merge \
--patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'
```

Wait until pods are running:

``` shell
kubectl get pods -n knative-serving
```

Configure DNS:

``` shell
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.22.1/serving-default-domain.yaml
```

Install Hello app:

``` shell
kubectl apply -f hello.yaml
```

Get hostname for Hello app:

``` shell
kubectl get ksvc
```
```
NAME    URL                                        LATESTCREATED   LATESTREADY   READY   REASON
hello   http://hello.default.172.18.0.2.sslip.io   hello-00001     hello-00001   True
```

Test Hello app:

``` shell
curl -H "Host: hello.default.172.18.0.2.sslip.io" http://localhost:8081
```
```
Hello World!
```

## Autoscaling

After curling the app, the pods are created:

``` shell
kubectl get pod -l serving.knative.dev/service=hello
```
```
NAME                                      READY   STATUS    RESTARTS   AGE
hello-00001-deployment-77549fcbf7-fzw68   2/2     Running   0          4s
```

After waiting a while, the pods are going to terminate due to no traffic:

``` shell
kubectl get pod -l serving.knative.dev/service=hello
```
```
NAME                                      READY   STATUS        RESTARTS   AGE
hello-00001-deployment-77549fcbf7-fzw68   2/2     Terminating   0          76s
```

After waiting another while:

``` shell
kubectl get pod -l serving.knative.dev/service=hello
```
```
No resources found in default namespace.
```

Curling the app again, the pods are created:

``` shell
curl -H "Host: hello.default.172.18.0.2.sslip.io" http://localhost:8081
kubectl get pod -l serving.knative.dev/service=hello
```
```
Hello World!
NAME                                      READY   STATUS    RESTARTS   AGE
hello-00001-deployment-77549fcbf7-zhkn8   2/2     Running   0          4s
```

## Traffic splitting

Deploy an update version of the hello service:

``` shell
kubectl apply -f hello-knative.yaml
```

Curl the same URL and see the changed hello message:

``` shell
curl -H "Host: hello.default.172.18.0.2.sslip.io" http://localhost:8081
```
```
Hello Knative!
```

View Revisions:

``` shell
kubectl get revisions
```
```
NAME          CONFIG NAME   GENERATION   READY   REASON   ACTUAL REPLICAS   DESIRED REPLICAS
hello-00001   hello         1            True             0                 0
hello-00002   hello         2            True             0                 0
```

Observe how all the traffic goes to the new revision:

``` shell
for i in {1..10}; do
  curl -s -H "Host: hello.default.172.18.0.2.sslip.io" http://localhost:8081
  printf '\n'
done | grep -v '^$' | sort | uniq -c
```
```
     10 Hello Knative!
```

Add a traffic split to the hello service:

``` shell
kubectl apply -f hello-split.yaml
```

Observe how the traffic is split between the revisions:

``` shell
for i in {1..10}; do
  curl -s -H "Host: hello.default.172.18.0.2.sslip.io" http://localhost:8081
  printf '\n'
done | grep -v '^$' | sort | uniq -c
```
```
      5 Hello Knative!
      5 Hello World!
```
