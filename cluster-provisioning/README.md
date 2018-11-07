# gcloud-sdk

## cluster
```
zone="europe-west1-b"
cluster="cluster-1"

gcloud container clusters create $cluster --zone $zone --username "admin" \
--cluster-version "1.11.2-gke.9" --machine-type "n1-standard-2" --image-type "COS" --disk-size "100" \
--scopes "https://www.googleapis.com/auth/compute","https://www.googleapis.com/auth/devstorage.read_only",\
"https://www.googleapis.com/auth/logging.write","https://www.googleapis.com/auth/monitoring",\
"https://www.googleapis.com/auth/servicecontrol","https://www.googleapis.com/auth/service.management.readonly",\
"https://www.googleapis.com/auth/trace.append" \
--num-nodes "4" --network "default" --enable-cloud-logging --enable-cloud-monitoring --enable-ip-alias --enable-autorepair --no-issue-client-certificate --metadata disable-legacy-endpoints=true --async
```

## nodepools
```
# notice the "gcloud beta" as taints are not supported in not-beta version

# preemptible pool

gcloud beta container node-pools create preemptible-pool \
        --cluster $cluster --enable-autorepair --machine-type n1-highmem-2 \
        --num-nodes=1 --enable-autoscaling --max-nodes=3 --min-nodes=1 --zone=$zone \
        --preemptible --scopes=gke-default,storage-rw  --metadata disable-legacy-endpoints=true

# master pool
gcloud beta container node-pools create master-pool-1 \
        --node-labels=env=prod --node-taints=env=prod:NoSchedule \
        --cluster $cluster --enable-autorepair --machine-type n1-highmem-2 \
        --num-nodes=1 --enable-autoscaling --max-nodes=3 --min-nodes=1 --zone=$zone \
        --scopes=gke-default,storage-rw --metadata disable-legacy-endpoints=true

# describe
gcloud container node-pools describe master-pool-1 --cluster $cluster --zone=$zone
```
