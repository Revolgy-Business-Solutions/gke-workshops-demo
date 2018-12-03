# Example deployment of Go app with CloudSQL (mysql) to GKE

## CLI tools initiation

In this course we will use following cli tools:

- gcloud
- kubectl
- docker

### gcloud

Download: `https://cloud.google.com/sdk/install`

Init configuration: `gcloud init`

Gcloud init will guide you through init configuration including authenticating with your gmail account which we will use to create our GCP project.

### kubectl

Kubernetes cli can be installed as a gcloud component. In can be installed as a standalone app but considering we will be working on GCP infrastructure it is not advisable.

```
gcloud components install kubectl
```

### Docker

instruction: [docs.docker.com/install](https://docs.docker.com/install)

Choose your operating system and follow instructions

### Tool Check

Check that all tools needed are installed correctly by running this commands

```
gcloud version
kubectl version
docker version
```

You should see outcome similar to this:

```
>> gcloud version

Google Cloud SDK 210.0.0
alpha 2018.07.16
app-engine-go 
app-engine-python 1.9.73
beta 2018.07.16
bq 2.0.34
core 2018.07.27
gsutil 4.33
kubectl 

>> kubectl version

Client Version: version.Info{Major:"1", Minor:"10", GitVersion:"v1.10.7", GitCommit:"0c38c362511b20a098d7cd855f1314dad92c2780", GitTreeState:"clean", BuildDate:"2018-08-20T10:09:03Z", GoVersion:"go1.9.3", Compiler:"gc", Platform:"linux/amd64"}
Server Version: version.Info{Major:"1", Minor:"10+", GitVersion:"v1.10.7-gke.6", GitCommit:"06898a4d0f2b96f82b43d9e59fa2825bd3d616a2", GitTreeState:"clean", BuildDate:"2018-10-02T17:32:01Z", GoVersion:"go1.9.3b4", Compiler:"gc", Platform:"linux/amd64"}

>> docker version
Client:
 Version:           18.06.0-ce
 API version:       1.38
 Go version:        go1.10.3
 Git commit:        0ffa8257ec
 Built:             Sun Aug 19 09:47:50 2018
 OS/Arch:           linux/amd64
 Experimental:      false

Server:
 Engine:
  Version:          18.06.0-ce
  API version:      1.38 (minimum version 1.12)
  Go version:       go1.10.3
  Git commit:       0ffa8257ec
  Built:            Sun Aug 19 09:47:31 2018
  OS/Arch:          linux/amd64
  Experimental:     false
```

Note you may have only Client version listed after fresh install and thats correct as well as different version or gcloud components listed. The idea here is that those commands works.

## GCP Project preparation

In this workshop we will be using Google Cloud Platform environment. It offers managed kubernetes environment, Google Kubernetes Engine as well as other services that will be used within this workshop such as managed MySQL database, container image repository and alternatively CI/CD tools.

### Creating Cluster

In GCP console: https://console.cloud.google.com/kubernetes/ we will create new Kubernetes cluster 

Setting will be guided in workshop, here are the main setting changes:

- Location type: Zonal
- Zone: europe-west1-c
- Master version: 1.10.6-gke.6

- NodePool:
  - size: 2
  - Boot disk size: 40Gb
  - Enable preemptible nodes
  - Disable auto-upgrade

- Scopes:
  - CloudSQL - enabled

## Docker image preparation

First, we need docker image from our application which we can deploy to kubernetes. In root of our project we create `Dockerfile` file:

The easiest way to build our image is just simply take golang base image prepare specificaly for building Go containers.


```
FROM golang:onbuild

EXPOSE 8080
```

This approach is strait forward, but with lots of issues and security threats. As we can see from docker image list, base image golang:onbuild has over 600Mb itself and after adding our code, it got only slighly bigger.

This means that most of the final image size consist of build dependencies and other stuff not needed for actuall program running.

One approach to reduce the size of final image is to use smaller base images. Almost all programming language base image will offer images in alpine or similar versions. In comparision with onbuild, we have to build and run application ourself as well as install dependencies in dockerfile.

```
FROM golang:alpine

WORKDIR /app

ADD . /app

RUN apk add --no-cache git \
 && go get github.com/gorilla/mux \
 && go get github.com/go-sql-driver/mysql

RUN cd /app && go build -o goapp

EXPOSE 8080

CMD ["./goapp"]
``` 

After building this dockerfile we will see that we reduced it to around half size. But its still pretty big considering that our application has just few megabytes in size.

One trick that we can use with compiled languages like Golang, is that we dont actually need all the building dependencies like in the case of interpreted languages.

What we can do is build the application in one container, and then copy only the executable to the other container.

In previous versions of docker this had to be done by using helper Dockerfile and build application in-place in repository.

In newer version we can use a builder pattern. Thats mean that we can specify multiple FROM lines in one Dockerfile where the final image will be build by the last one. The power in this is that we can copy files between these stages.

```
FROM golang:onbuild as builder

WORKDIR /go/src/app

ADD . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

#---

FROM alpine

WORKDIR /root/

COPY --from=builder /go/src/app/main .

CMD ["./main"]
```

You have to tag builded image as `eu.gcr.io/_GCP_PROJECT_ID/SOME_FOLDER/APP_NAME:BUILD_NUMBER`

In my case docker build will look like this:

```
docker build -t eu.gcr.io/xxx/k8s-demo/app:1
```

Next we can push this image to Container Registry on our Google Cloud Project

```
gcloud docker -- push eu.gcr.io/xxx/k8s-demo/app:1
```

## Kubernetes configuration

### Accessing cluster

```
gcloud container clusters list
```

```
gcloud container clusters get-credentials k8s-demo --zone europe-west1-c
```

### Create namespace

```
kubectl create namespace demo
```

```
kubectl get namespaces / kubectl get ns
```

```
kubectl config current-context 
```

```
kubectl config set-context $(kubectl config current-context) --namespace demo
```

### Deployment configuration

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demoapp
  labels:
    app: demoapp
    version: 1.0.0
spec:
  replicas: 1
  selector:
    matchLabels:
      app: demoapp
  template:
    metadata:
      labels:
        app: demoapp
    spec:
      containers:
        - name: demoapp
          image: _IMAGE_NAME
          ports:
            - containerPort: 3000
```

### Deploying application

```
kubectl create -f app-deployment.yaml
```

### manual scaling of deployment

```
kubectl scale deployment demoapp --replicas 2
```

### Accessing microservices

We have our application deployed, but the only way we can access our pods is via their private IP. Thats not very convenient since we would have to address each replica separately and IP would change with every restart of application.

Kubernetes has feature called services for this. Kubernetes service could be described as loadbalancing reverse proxy above all pods from deployment. 

```
apiVersion: v1
kind: Service
metadata:
  name: demoapp
spec:
  selector:
    app: demoapp
  ports:
  - name: http-demoapp
    protocol: TCP
    port: 80
    targetPort: 3000
  type: LoadBalancer
```

With type: `LoadBalancer` we get assigned external IP address to our service. We can now directly access our service with this IP address

Since it's not the best practice to use these service types and have multiple services accessible directly from the internet, we will edit service in place with kubectl and change service type to NodePort.

```
kubectl edit svc demoapp
```

Next we will make service service accessible with better solution.

#### Kubernetes Ingress

Acording to best practices, system should have allways have only one entrypoint to the cluster from the internet

```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: basic-ingress
spec:
 rules:
 - host: XXXXX.nip.io
   http:
     paths:
     - path: /
       backend:
         serviceName: demoapp
         servicePort: 80
```

### Deployment health cycle - Health and Readiness probess

Deployment itself keeps desired replicas deployed but it can only see if container is running or not which is based on if core continer process is running or not. What it can't do though is recognize where application is healthy, working properly, isn't overloaded etc. For this purpose we have two probes that we can utilize.

```
    readinessProbe:
      tcpSocket:
        port: 3000
      initialDelaySeconds: 10
      periodSeconds: 10

    livenessProbe:
      tcpSocket:
        port: 3000
      initialDelaySeconds: 15
      periodSeconds: 20
```

We can redeploy our application with:

```
kubectl apply -f app.yaml
```

Kuard demo:

```
k apply -f kubernetes/kuard.yaml
```

For testing or debugging purposes, we dont have to creating loadbalancers and ingresses for every deployment we run. Kubectl has ability to forward ports directly from pods to local network.

```
kubectl port-forward kuard 8080:8080 &
```

### Deployment rollout strategy

We can specify parameters of upgrading deployment to new image version.
Default upgrade strategy is rolling update with 25% max unavailable and 25% max surge

```
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxSurge: 1
    maxUnavailable: 0
```

### kubectl set image deployment 

We don't have to apply whole deployment configuration when we need to upgrade deployment image

```
kubectl set image deployment.v1.apps/demoapp demoapp=IMAGE_TAG
```

### Deployment resource management

We can limit the amount of resources that each container can utilize on node as well as how much is guaranteed for container to have. 

```
resources: 
  requests: 
    cpu: 200m
    memory: 128Mi 
  limits: 
    cpu: 300m 
    memory: 256Mi
```

### Deployment autoscaling

We can turn on deployment autoscaling manually with kubectl cli.

```
kubectl autoscale deployment demoapp --min 1 --max 3 --cpu-percent 60
```

HPA can be also deploy from file which is usualy preffered. Here we will try to get configuration setup of already running resource. `-o yaml` parameter returns yaml representation of deployed resource

```
kubectl get hpa demoapp -o yaml
```

We can delete internal kubernetes generated field from given yaml config and we will end up with usable configuration.

```
apiVersion: autoscaling/v1
kind: HorizontalPodAutoscaler
metadata:
  name: demoapp
spec:
  maxReplicas: 3
  minReplicas: 1
  scaleTargetRef:
    apiVersion: extensions/v1beta1
    kind: Deployment
    name: demoapp
  targetCPUUtilizationPercentage: 60
```

Now we can try generate some load on our application

### Sidecar containers

In some situations, multiple containers per pod can be deployed together. It should serve as helper container only, we shouldn't deploy multiple core processes to one pod.

In this scenario we will deploy CloudSQL proxy which will allow us secure connection to CLoudSQL instance.

### Init Containers

As pod can have multiple containers, it can have one or more init containers. Init containers allways runs before application containers and they have to complete successfuly for application containers to start. In case of multiple init containers, they run one after the other, not simultaneously.

The advantage of using init containers is that whatever init container is used for dont have to be integrated into application so in case of a restart of app container, there is no excess operations because init containers dont run again.

example usage:

- download or process configs i.e. remotely, from config maps or secrets to local storage
- prepare some environment variables for main containers
- wait for services that app is dependent on to become ready

Here we have example of init container that downloads index page of kubernetes website and saves it to local storage.

```
initContainers:
- name: install
  image: busybox
  command:
  - wget
  - "-O"
  - "/work-dir/kube.html"
  - http://kubernetes.io
  volumeMounts:
  - name: workdir
    mountPath: "/work-dir"
```

We mounted volume into the container, but we didn't specify where or how the volume is created.

```
volumes:
- name: workdir
  emptyDir: {}
```

Next we mount that volume to our application container so we can access downloaded index.

```
volumeMounts:
- name: workdir
  mountPath: /usr/src/app/assets
```


### Creating Cloud SQL user for authentication

First we will create SQL user with gcloud cli. We need to know name of our SQL instance to do so. We can either find it in GCP UI console, or with following command

```
gcloud sql instances list
```

Next we can create db user: 

```
gcloud sql users create [DBUSER] --host=% --instance=[INSTANCE_NAME] --password=[PASSWORD]
```

#### Creating secrets

Kubernetes environment supports secrets where you can add critical information and credentials and securely inject it directly into containers. That way these information dont have to be included nowhere in CI/CD pipeline and thus creating security risks.

We need two secrets to deploy sql proxy as a sidecar. One for authenticating sql proxy itself on GCP security layer and second for user authentication on database layer.

```
kubectl create secret generic cloudsql-instance-credentials \
    --from-file=credentials.json=[PROXY_KEY_FILE_PATH]
```

```
kubectl create secret generic cloudsql-db-credentials \
    --from-literal=username=[DBUSER] --from-literal=password=[PASSWORD]
```

#### Cloudsql proxy deployment configuration

with credentials prepared, we can add sidecar yaml configuration to our application deployment.

```
- name: cloudsql-proxy
  image: gcr.io/cloudsql-docker/gce-proxy:1.11
  command: ["/cloud_sql_proxy",
            "-instances=<INSTANCE_CONNECTION_NAME>=tcp:3306",
            "-credential_file=/secrets/cloudsql/credentials.json"]
  securityContext:
    runAsUser: 2  # non-root user
    allowPrivilegeEscalation: false
  volumeMounts:
    - name: cloudsql-instance-credentials
      mountPath: /secrets/cloudsql
      readOnly: true
```

```
volumes:
  - name: cloudsql-instance-credentials
    secret:
      secretName: cloudsql-instance-credentials
```

#### Adding secrets as container environment variables

```
env:
- name: DB_USER
  valueFrom:
    secretKeyRef:
      name: cloudsql-db-credentials
      key: username
- name: DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: cloudsql-db-credentials
      key: password
```

### Assigning pods to nodes

#### Node affinities

Node affinity is used for restricting on what nodes can specified pod run

DEPRECATED
```
nodeSelector:
  cloud.google.com/gke-nodepool: dev-pool
```

In newer version of kubernetes affinities needs to be used for handling pod assigning.

```
affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: env
            operator: In
            values:
            - dev
      preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 1
        preference:
          matchExpressions:
          - key: node-type
            operator: In
            values:
            - preemptible
```

#### Pod affinity and anti-affinity

We can also configure with pod affinity and anti-affinity that we want that pod to run or not run on nodes that allready runs pods with specific labels.

```
affinity:
  podAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchExpressions:
        - key: app
          operator: In
          values:
          - kuard
      topologyKey: "kubernetes.io/hostname"
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchExpressions:
        - key: app
          operator: In
          values:
          - demoapp
      topologyKey: "kubernetes.io/hostname"
```

#### Taints and tolerations

```
kubectl taint nodes node1 env=dev:NoSchedule
```

```
tolerations:
- key: "env"
  operator: "Equal"
  value: "dev"
  effect: "NoSchedule"
```
