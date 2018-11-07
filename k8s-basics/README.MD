# Example deployment of nodejs app with mysql to GKE

## CLI tools initiation

In this course we will use following cli tools:

- gcloud
- kubectl
- Docker

### gcloud

Download: `https://cloud.google.com/sdk/install`

Init configuration: `gcloud init`

Gcloud init will guide you through init configuration including authenticating with your gmail account which we will use to create our GCP project.

### Kubectl

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

-NodePool:
  - size: 2
  - Boot disk size: 40Gb
  - Enable preemptible nodes
  - Disable auto-upgrade

- Scopes:
  - CloudSQL - enabled

### Prepare CI/CD (Optional, may not part of kubernetes workshop)

In this workshop we use GCP integrated solution for CI/CD. For CI part, there is feature called "Source Repositories" and for the CD part there is "Cloud Build". It's convenient for our use case since we don't need to employ more 3rd party services than necessary.

Many other CI/CD solution can be used for deploying into GKE (i.e. github / gitlab / bitbucket with jenkins / gilab-ci / shipable etc.) but they are not covered in this workshop.

#### Google Cloud Source Repositories

Cloud Source Repository serve as a private git repository. We create repository in Google UI console. In Tools section of menu there is "Source repositories". Here we select tab "Repositories" where we can create new repository. After creation we are presented with commands to either clone repository or connect existing one.

#### Cloud Build

In GCP console, we select Cloud Build from left menu. We will immediately see Build history, which is probably going to be empty because we didnt run our pipeline yet.

When you go to Build triggers tab, you can create your first deployment trigger with "Add Trigger" button.

As of now, trigger creation have 3 steps, i will list here values we want to fill into each step

1 - Source
- Cloud Source Repository

2 - Repository
- Select your created Source Repository

3 - Trigger settings
- Name - Anything descriptive (optional)
- Trigger type: (default) Branch
- Branch: .*
- Included and ignored files filter: (default) empty
- Build configuration: cloudbuild.yaml
- cloudbuild.yaml location: (default) /cloudbuild.yaml
- substitution variables:
  - _CLOUDSDK_CONTAINER_CLUSTER: your cluster name 
  - _CLOUDSDK_COMPUTE_ZONE: your cluster zone

### Create CloudSQL database

CloudSQL is MySQL (not only) database managed by Google. In the "Storage" section in GCP console is "SQL". Here you can "Create Instance". We want MySQL, second generation database. For our purpose, lowest configurations is sufficient. Here is basic overview of non-default CloudSQL configuration (other than descriptive name and root password)

region and Zone - Select same region and zone as your new cluster
machine type and storage - if not set with "MySQL development" option, select "db-f1-micro" ad machine type and 10 GB storage capacity

Confirm configuration and wait a few minutes for database to create.

#### CloudSQL setup

Users and access will be configure later in the workshop. Now we will just prepare database, tables and some example data. In database details in GCP console - SQL we have an option to `Connect using Cloud Shell`. After initiation of virtual console we are presented with prearanged gcloud command that will connect us directly to the datase. We will need to enter root password for this connection.

Once we are in the database. we will create our test database, table and insert some test data.

```
CREATE DATABASE demoapp CHARACTER SET utf8 COLLATE utf8_general_ci;
```

```
CREATE TABLE `users` (
  `id`       int(11)     unsigned NOT NULL AUTO_INCREMENT,
  `name`     varchar(30) DEFAULT '',
  `email`    varchar(50) DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
```

```
INSERT INTO users (name, email) 
     VALUES ('Earl Grey', 'earl@grey.com'), 
            ('Lapsang Souchong',  'lapsang@souchong.com');
```

#### Allow CloudSQL Admin API

Later on we will use special proxy that will allow us to securely connect to the database. For this to work we need to enable `CloudSQL Admin API` on GCP. In the `APIs & Services` we have option to `Enable apis and services`. If we search for `SQL` in the opened API Library we will see our `Cloud SQL Admin API` which we need to select and then `ENABLE`.

#### create service account

For authentication to the database we will need service account. You can create one in "IAM & admin" section and "Service accounts" subsection. Here we will `Create service account`. We name him i.e. sqlproxy and on the permissions step we will select `Cloud SQL Admin`. On the last step we will generate json config file which we save for later.

## Docker image preparation

First, we need docker image from our application which we can deploy to kubernetes. In root of our project we create `Dockerfile` file:

```
FROM node:8

WORKDIR /usr/src/app

ADD package.json ./

RUN npm install

ADD . .

EXPOSE 3000

CMD ["npm", "start"]
```

We will build and deploy image manually so we can experience whats behind most of the CI/CD pipelines.

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
    targetPort: 8000
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
  backend:
    serviceName: demoapp
    servicePort: 80
```

#### Update firewall rules

For ingress to work correctly, we need to create a firewall rule that will allow traffic from Google subnets to our pods. The reason for that is that from these subnets are comming health check request that mark Loadbalancers backend on GCP layer healthy or not. If health check do not reach our services, ingress will not route traffic to them.

Firewall rule configurationi (besides default values):

- Targets: All instances in the network

- Source IP ranges:
  - 35.191.0.0/16
  - 130.211.0.0/22

- ports: tcp: 30000-32767

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
