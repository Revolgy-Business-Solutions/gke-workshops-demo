```
# notice the "gcloud beta" as taints are not supported in not-beta version

# preemptible pool

gcloud beta container node-pools create preemptible-pool \
        --cluster testing --enable-autorepair --machine-type n1-highmem-2 \
        --num-nodes=1 --enable-autoscaling --max-nodes=3 --min-nodes=1 --zone=europe-west1-b \
        --preemptible --scopes=gke-default,storage-rw

# master pool
gcloud beta container node-pools create master-pool-1 \
        --node-labels=env=prod --node-taints=env=prod:NoSchedule \
        --cluster testing --enable-autorepair --machine-type n1-highmem-2 \
        --num-nodes=1 --enable-autoscaling --max-nodes=3 --min-nodes=1 --zone=europe-west1-b \
        --scopes=gke-default,storage-rw

# describe
gcloud container node-pools describe master-pool-1 --cluster testing --zone=europe-west1-b
```
