# gke-workshops-demo

# debug a failed pod with labeling it so that it's removed from a prod replicaset
```
kubectl label pods kuard-prod-75b665f5f6-c52wh track=failed  --overwrite
```
# kubectl cheat sheet
## run
```
kubectl run -l app=other --image=alpine --restart=Never --rm -i -t test-1

kubectl run hello-web --labels app=hello \
  --image=gcr.io/google-samples/hello-app:1.0 --requests=cpu=200m --port 8080 --expose
  
kubectl run -it --rm --image=mysql:5.6 --restart=Never -n prod mysql-client \
  -- sh -c 'apt-get update && apt-get install -y dnsutils && nslookup sqlproxy && mysql -h sqlproxy -uuser -ppassword'
  
kubectl run -it --rm --image=mysql:5.6 --restart=Never mysql-client -- mysql -h 1.2.3.4 -uuser -ppassword --execute 'show databases;'  
```
## expose
```
kubectl expose deployment nginx --type=NodePort --name nginx --port 80

kubectl expose deployment iis --type="LoadBalancer" --name=iis --port=80 --target-port=80
```
## get / delete
```
kubectl get pods -l=app=kuard  
kubectl delete pod -l app=kuard

kubectl delete -f filewithmoreobjectsdefined.yaml

kubectl delete pod $(kubectl get pods | grep 0/1 | awk '{print $1}')

kubectl delete pv --all
kubectl delete pvc --all

watch kubectl get ep,svc,po,ing -o wide --show-labels

kubectl get all --all-namespaces -o yaml --export
```
## apply
```
kubectl apply --recursive -f clusterrolebindings.rbac.authorization.k8s.io/

kubectl create clusterrolebinding cluster-admin-binding \
  --clusterrole cluster-admin --user $(gcloud config get-value account)
  
for FOLDER in *; do echo "doing $FOLDER" ; kubectl apply --recursive --validate=false -v=4 -f ./$FOLDER; done
```
## edit
```
kubectl edit cm config-autoscaler -n knative-serving
kubectl edit sts redis-cache-dev-master
```
## logs
```
kubectl logs -f redis-cache-dev-master-0 -c volume-permissions
```
## config
```
kubectl config view -o template --template='{{ index . "current-context" }}'
```
## rollout
```
kubectl rollout status statefulset web
```
## drain
```
kubectl drain gke-cluster-name-default-pool-426bc9ef-ngv3 --ignore-daemonsets --delete-local-data --v=4
```
## useful aliases
```
alias util='kubectl get nodes --no-headers | awk '\''{print $1}'\'' | xargs -I {} sh -c '\''echo {} ; kubectl describe node {} | grep Allocated -A 5 | grep -ve Event -ve Allocated -ve percent -ve -- ; echo '\'''
util
```
