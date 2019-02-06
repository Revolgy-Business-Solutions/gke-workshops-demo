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

In this workshop we will be using Google Cloud Platform environment. It offers managed kubernetes environment, Google Kubernetes Engine as well as other services that will be used within this workshop such as managed MySQL database and container.

For that you have to create Google Cloud Project [https://cloud.google.com/resource-manager/docs/creating-managing-projects](cloud.google.com/resource-manager/docs/creating-managing-projects) and you have to enable billing for this project as you can't create any resources without billing configured.

Google Cloud Platform have a free tear for new accounts/projects that consist of $300 (or 12 months, whichever run out first). That is plenty for this whole workshop and other testing.

### Creating Cluster

In GCP console: https://console.cloud.google.com/kubernetes/ we will create new Kubernetes cluster 

Setting will be guided in workshop, here are the main setting changes:

- Location type: Zonal
- Zone: europe-west1-c
- Master version: 1.10.6-gke.6 (or current default value)

- NodePool:
  - size: 3
  - Boot disk size: 40Gb
  - Enable preemptible nodes
  - Disable auto-upgrade

## Docker image preparation

First, we need docker image from our application which we can deploy to kubernetes. In root of our project we create `Dockerfile` file.

### Docker best practices - Image size

The easiest way to build our image is just simply take golang base image prepare specificaly for building Go containers.


```
FROM golang:onbuild

EXPOSE 8080
```

This approach is strait forward, but with lots of issues and security threats. As we can see from docker image list, base image golang:onbuild has over 600Mb itself and after adding our code, it got only slighly bigger.

That means that most of the final image size consist of build dependencies and other stuff not needed for actuall program running.

One approach to reduce the size of final image is to use smaller base images. For almost all programming languages there will be images in alpine or similar versions available. In comparision with onbuild, we have to build and run application ourself as well as install dependencies in dockerfile.

```
FROM golang:alpine

WORKDIR /app

ADD . /app

RUN apk add --no-cache git \
 && go get github.com/gorilla/mux \
 && go get github.com/go-sql-driver/mysql

RUN cd /app && go build -o goapp

EXPOSE 8000

CMD ["./goapp"]
``` 

After building this dockerfile we will see that we reduced it to around half size. But its still pretty big considering that our application has just few megabytes in size.

One trick that we can use with compiled languages like Golang, is that we dont actually need either building, neither running dependencies like in the case of interpreted languages.

What we can do is build the application in one container, and then copy only the executable file to the final container.

In previous versions of docker this had to be done by using another, helper, Dockerfile and build application in-place in repository.

In newer version we can use the builder pattern. Thats mean that we can specify multiple FROM lines in one Dockerfile where the final image will be build by the last one. The power in this is that we can copy files between these stages.

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

As you will see from docker images, this final images will have a little over 10Mb in size. Thats quite a difference over the previsous 700Mb. This means lower pulling times on VM instances as well as smaller atack surface for possible vulnerabilities.

Even this container isn't the smallest possible. As we have a executable file, we dont even need all the alpine dependencies and use SCRATCH base image, which basically means a blank base image where we copy only our executable files. Of course in this case you have to take care about things that you application possibly needs, for example for receiving ssl traffic you need to install ca-certificates and so on. 

With this last method you will get the final image to around 6Mb which is 1/100th of the onbuild variation.

### Uploading image to Google Container Repository (GCR)

As we will be running kubernetes environmnet on GCP, we want to pull all images from associated projects GCR. That provides faster image pull speeds and better security.

In order to upload docker image into GCR you have to tag builded image as `eu.gcr.io/GCP_PROJECT_ID/SOME_FOLDER/APP_NAME:BUILD_NUMBER`

In my case docker build will look like this:

```
export PROJECT=$(gcloud info --format='value(config.project)')
docker build -t eu.gcr.io/$PROJECT/k8s-demo/app:1 . 
```

Second step is (if not done already) is to set gcloud as Docker credential helper [https://cloud.google.com/container-registry/docs/advanced-authentication](https://cloud.google.com/container-registry/docs/advanced-authentication) `gcloud auth configure-docker`. After that, we can use typical `docker push IMAGE_TAG` and docker engine will automatically (based on the host prefix "eu.gcr.io" or just "gcr.io") push the image into the GCR of a projects specified in the image tag (provided you have enough priviledges with you account to do that)

## Kubernetes access configuration

### Accessing cluster

```
gcloud container clusters list
```

```
gcloud container clusters get-credentials k8s-demo --zone europe-west1-c
```

## Operating Kubernetes

### Create namespace

```
kubectl create namespace demo
```

```
kubectl get namespaces
kubectl get ns
```

```
kubectl config current-context 
```

```
kubectl config set-context $(kubectl config current-context) --namespace demo
```

```
alias ksc='function _ksc(){ kubectl config set-context `kubectl config current-context` --namespace=$1; };_ksc'
ksc demo
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
  replicas: 3
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
          image: kotrev/go-demo-app:1.0.0
          ports:
            - containerPort: 8000
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
    targetPort: 8000
  type: LoadBalancer
```

With type: `LoadBalancer` we get assigned external IP address to our service. We can now directly access our service with this IP address

Since it's usually not the best practice to use these service types and have multiple services accessible directly from the internet, we will edit service in place with kubectl and change service type to NodePort.

```
kubectl edit svc demoapp
```

Next we will make service service accessible with better solution.

#### Kubernetes Ingress

Acording to best practices, system should have usually only one entrypoint to the cluster from the internet (or one per environment etc.)

Such gateway is implemented in Kubernetes as Ingress resource.

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
              port: 8000
            initialDelaySeconds: 10
            periodSeconds: 10
          livenessProbe:
            tcpSocket:
              port: 8000
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
kubectl set image deployment/demoapp demoapp=kotrev/go-demo-app:2.0.0
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
kubectl autoscale deployment demoapp --min 3 --max 10 --cpu-percent 60
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


```
IP=$(kubectl get svc demoapp -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
siege -c 100 -d '0.5' http://$IP/
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
          - weight: 100
            podAffinityTerm:
              labelSelector:
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

### Volumes

There is a lot of different type of volumes that can by mounted into the container. You can find full list of them in the kubernetes documentation.

We will deploy simple nginx server with mounted Google Persistend Disk. First we need to create the disk. On GCP go to compute -> disks and create new disk with the name acording to the name in /volume/nginx-pd-volume.yaml in the repo. Or named it as you like and we will update the configuration.

Now deploy nginx-pd-volume.yaml with the optional change of the PDName (the name of the GCP volume).

When you will query nginx from the create service loadbalacer. You will see either no answer or an error. Thats because we overwritten the /usr/share/nginx/html folder with the volume. We will generate our own persistent html file by executing into the container

```
kubectl exec -it nginx bash
```

and creating some test index.html file

```
echo "Hello from the volume!"
```

If we query our nginx server again, on root, we will get our generated html site. Now try to delete the pod and recreate it again and see it the data are trully persistent.

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
    - name: test-volume
      mountPath: "/work-dir"
```

## CLoudSQL sidecar/secrets demo

### Create CloudSQL database

CloudSQL is MySQL (not only) database managed by Google. In the "Storage" section in GCP console is "SQL". Here you can "Create Instance". We want MySQL, second generation database. For our purpose, lowest configurations is sufficient. Here is basic overview of non-default CloudSQL configuration (other than descriptive name and root password)

region and Zone - Select same region and zone as your new cluster
machine type and storage - if not set with "MySQL development" option, select "db-f1-micro" ad machine type and 10 GB storage capacity

Confirm configuration and wait a few minutes for database to create.

#### CloudSQL setup

Users and access will be configure later in the workshop. Now we will just prepare database, tables and some example data. In database details in GCP console - SQL we have an option to `Connect using Cloud Shell`. After initiation of virtual console we are presented with prearanged gcloud command that will connect us directly to the datase. We will need to enter root password for this connection.

Once we are in the database. we will create our test database, table and insert some test data.

```
CREATE DATABASE bookstore CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
use bookstore;
```

```
CREATE TABLE IF NOT EXISTS books (
    id INT AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    author VARCHAR(255),
    pages INT,
    PRIMARY KEY (id)
)  ENGINE=INNODB;
```

```
INSERT INTO books (name, author, pages) VALUES ("testname", "testauthor", 123);
```

#### Allow CloudSQL Admin API

Later on we will use special proxy that will allow us to securely connect to the database. For this to work we need to enable `CloudSQL Admin API` on GCP. In the `APIs & Services` we have option to `Enable apis and services`. If we search for `SQL` in the opened API Library we will see our `Cloud SQL Admin API` which we need to select and then `ENABLE`.

#### create service account

For authentication to the database we will need service account. You can create one in "IAM & admin" section and "Service accounts" subsection. Here we will `Create service account`. We name it i.e. sqlproxy and on the permissions step we will select `Cloud SQL Admin`. On the last step we will generate json config file which we save for later.


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

#### Defining environment variables for connecting to DB
```
          env:
          - name: "MYSQL_DATABASE"
            value: "True"
          - name: "MYSQL_USER"
            value: "user"
          - name: "MYSQL_PASS"
            value: "usertest"
```
As we can see, we are specifying user credentials directly in the container configuration. This is, of course, bad practice since those credentials could be eaily accessed either from container description or stored configuration in repo. We will change it to use kubernetes secrets.

#### Adding secrets as container environment variables

```
          env:
          - name: "MYSQL_DATABASE"
            value: "True"
          - name: MYSQL_USER
            valueFrom:
              secretKeyRef:
                name: cloudsql-db-credentials
                key: username
          - name: MYSQL_PASS
            valueFrom:
              secretKeyRef:
                name: cloudsql-db-credentials
                key: password
```
