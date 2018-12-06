https://github.com/kubernetes/ingress-gce/tree/master/examples

```
kubectl apply -f ../../helm/rbac.yaml
helm init --service-account=tiller --history-max=10 --debug --upgrade
```
```
gcloud compute addresses create kubernetes-ingress --global
IP=$(gcloud compute addresses describe kubernetes-ingress --global --format='value(address)')
sed "s/__IP__/$IP/" gce-tls-ingress.yaml
```

```
watch kubectl get ep,svc,po -o wide --show-labels
```
