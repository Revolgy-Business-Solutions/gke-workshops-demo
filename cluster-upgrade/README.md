

```
kubectl cordon gke-NODE_NAME
kubectl drain gke-NODE_NAME --ignore-daemonsets --delete-local-data --v=4
kubectl delete node gke-NODE_NAME
```
