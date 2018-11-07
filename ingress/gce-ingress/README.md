https://github.com/kubernetes/ingress-gce/tree/master/examples

```
kubectl apply -f ../../helm/rbac.yaml
helm init --service-account=tiller --history-max=10 --debug --upgrade
```
```
gcloud compute addresses create kubernetes-ingress --global
```

```
watch kubectl get ep,svc,po -o wide --show-labels
```
