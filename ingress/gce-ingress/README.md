https://github.com/kubernetes/ingress-gce/tree/master/examples

```
gcloud compute addresses create kubernetes-ingress --global
```

```
watch kubectl get ep,svc,po -o wide --show-labels
```
