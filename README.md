# k8s-exercises

## Creation of local cluster

``` shell
k3d cluster create --agents 2 -p 8080:80@loadbalancer --k3s-arg '--disable=traefik@server:0'
kubectl apply --server-side -f https://github.com/envoyproxy/gateway/releases/download/v1.8.2/install.yaml
kubectl -n envoy-gateway-system rollout status deployment/envoy-gateway --timeout=180s

kubectl create namespace argo-rollouts
kubectl apply -n argo-rollouts -f https://github.com/argoproj/argo-rollouts/releases/latest/download/install.yaml

helm repo add nats https://nats-io.github.io/k8s/helm/charts
helm repo update

kubectl create namespace argocd
kubectl apply -n argocd --server-side=true --force-conflicts \
  -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

istioctl install \
  --set profile=ambient \
  --set values.global.platform=k3d \
  --set values.cni.cniBinDir=/var/lib/rancher/k3s/data/cni
```

## Chapter 2

- [1.1.](https://github.com/jaakkomo/k8s-exercises/tree/1.1/log-output)
- [1.2.](https://github.com/jaakkomo/k8s-exercises/tree/1.2/the-project)
- [1.3.](https://github.com/jaakkomo/k8s-exercises/tree/1.3/log-output)
- [1.4.](https://github.com/jaakkomo/k8s-exercises/tree/1.4/the-project)
- [1.5.](https://github.com/jaakkomo/k8s-exercises/tree/1.5/the-project)
- [1.6.](https://github.com/jaakkomo/k8s-exercises/tree/1.6/the-project)
- [1.7.](https://github.com/jaakkomo/k8s-exercises/tree/1.7/log-output)
- [1.8.](https://github.com/jaakkomo/k8s-exercises/tree/1.8/the-project)
- [1.9.](https://github.com/jaakkomo/k8s-exercises/tree/1.9/ping-pong)
- [1.10.](https://github.com/jaakkomo/k8s-exercises/tree/1.10/log-output)
- [1.11.](https://github.com/jaakkomo/k8s-exercises/tree/1.11)
- [1.12.](https://github.com/jaakkomo/k8s-exercises/tree/1.12/the-project)
- [1.13.](https://github.com/jaakkomo/k8s-exercises/tree/1.13/the-project)

## Chapter 3

- [2.1.](https://github.com/jaakkomo/k8s-exercises/tree/2.1)
- [2.2.](https://github.com/jaakkomo/k8s-exercises/tree/2.2/the-project)
- [2.3.](https://github.com/jaakkomo/k8s-exercises/tree/2.3)
- [2.4.](https://github.com/jaakkomo/k8s-exercises/tree/2.4/the-project)
- [2.5.](https://github.com/jaakkomo/k8s-exercises/tree/2.5/log-output)
- [2.6.](https://github.com/jaakkomo/k8s-exercises/tree/2.6/the-project/todo-app)
- [2.7.](https://github.com/jaakkomo/k8s-exercises/tree/2.7/ping-pong)
- [2.8.](https://github.com/jaakkomo/k8s-exercises/tree/2.8/the-project)
- [2.9.](https://github.com/jaakkomo/k8s-exercises/tree/2.9/the-project/todo-backend)
- [2.10.](https://github.com/jaakkomo/k8s-exercises/tree/2.10)

## Chapter 4

- [3.1.](https://github.com/jaakkomo/k8s-exercises/tree/3.1/ping-pong)
- [3.2.](https://github.com/jaakkomo/k8s-exercises/tree/3.2)
- [3.3.](https://github.com/jaakkomo/k8s-exercises/tree/3.3)
- [3.4.](https://github.com/jaakkomo/k8s-exercises/tree/3.4/ping-pong)
- [3.5.](https://github.com/jaakkomo/k8s-exercises/tree/3.5/the-project)
- [3.6.](https://github.com/jaakkomo/k8s-exercises/tree/3.6)
- [3.7.](https://github.com/jaakkomo/k8s-exercises/tree/3.7)
- [3.8.](https://github.com/jaakkomo/k8s-exercises/tree/3.8/.github/workflows)
- [3.9.](https://github.com/jaakkomo/k8s-exercises/tree/3.9/the-project)
- [3.10.](https://github.com/jaakkomo/k8s-exercises/tree/3.10/the-project/backup)
- [3.11.](https://github.com/jaakkomo/k8s-exercises/tree/3.11/the-project)
- [3.12.](https://github.com/jaakkomo/k8s-exercises/tree/3.12/the-project)

## Chapter 5

- [4.1.](https://github.com/jaakkomo/k8s-exercises/tree/4.1)
- [4.2.](https://github.com/jaakkomo/k8s-exercises/tree/4.2/the-project)
- [4.3.](https://github.com/jaakkomo/k8s-exercises/tree/4.3/monitoring)
- [4.4.](https://github.com/jaakkomo/k8s-exercises/tree/4.4/ping-pong)
- [4.5.](https://github.com/jaakkomo/k8s-exercises/tree/4.5/the-project)
- [4.6.](https://github.com/jaakkomo/k8s-exercises/tree/4.6/the-project)
- [4.7.](https://github.com/jaakkomo/k8s-exercises/tree/4.7/log-output)
- [4.8.](https://github.com/jaakkomo/k8s-exercises/tree/4.8/the-project)
- [4.9.](https://github.com/jaakkomo/k8s-exercises/tree/4.9/the-project)
- 4.10.
  - [Manifests](https://github.com/jaakkomo/k8s-project-manifests/tree/4.10)
  - [Workflows](https://github.com/jaakkomo/k8s-exercises/tree/4.10/.github/workflows)

## Chapter 6

- [5.1.](https://github.com/jaakkomo/k8s-exercises/tree/5.1/dummy-site)
- [5.2.](https://github.com/jaakkomo/k8s-exercises/tree/5.2/istio-samples)
