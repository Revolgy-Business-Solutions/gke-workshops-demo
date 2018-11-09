# ingress

```
kubectl run hello-web --labels app=hello \
  --image=gcr.io/google-samples/hello-app:1.0 --port 8080 --expose

kubectl apply -f hello-allow-from-foo.yaml

kubectl run -l app=foo --image=alpine --restart=Never --rm -i -t test-1

# wget -qO- --timeout=2 http://hello-web:8080

```
Traffic from Pod app=foo to the app=hello Pods is enabled.

```
kubectl run -l app=other --image=alpine --restart=Never --rm -i -t test-1

# wget -qO- --timeout=2 http://hello-web:8080

```

download will time out.

# egress
```
kubectl apply -f foo-allow-to-hello.yaml

kubectl run hello-web-2 --labels app=hello-2 \
  --image=gcr.io/google-samples/hello-app:1.0 --port 8080 --expose

kubectl run -l app=foo --image=alpine --rm -i -t --restart=Never test-3

# wget -qO- --timeout=2 http://hello-web:8080

# wget -qO- --timeout=2 http://hello-web-2:8080

# wget -qO- --timeout=2 http://www.example.com
```


