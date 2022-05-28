# Create and Use Persistent Disk on GKE
Hashimoto Du at DSS

Instruction for build and run as below


## 1. Prepare a persistent disk

```
# create persistent disk
gcloud compute disks create --size=10GB --zone=asia-northeast1-b sample-gce-nfs-disk


# create vm attaches the disk
gcloud compute instances create diskformat-vm --zone=asia-northeast1-b --machine-type=f1-micro  --disk-name=sample-gce-nfs-disk

# ssh to vm
gcloud compute ssh --zone=asia-northeast1-b diskformat-vm

sudo lsblk
> NAME   MAJ:MIN RM SIZE RO TYPE MOUNTPOINT
> sda      8:0    0  10G  0 disk 
> `-sda1   8:1    0  10G  0 part /
> sdb      8:16   0  10G  0 disk 

# format disk
sudo mkfs.ext4 -m 0 -F -E lazy_itable_init=0,lazy_journal_init=0,discard /dev/sdb
```


## 2. Create PersistentVolume and PersistentVolumeClaim in GKE
```
# create a cluster in GKE
gcloud container clusters create ds-gke-small-cluster \
	--project ds-project \
	--zone asia-northeast1-b \
	--machine-type n1-standard-1 \
	--num-nodes 1 \
	--enable-stackdriver-kubernetes

# get credentials to use to access cluster
gcloud container clusters get-credentials --zone asia-northeast1-b ds-gke-small-cluster
```

```
# create persistent volume
kubectl apply -f pvc-demo.yaml
```

## 3. Deploy Postgres using Persistent Volume saving data
```
＃　deploy postgres on gke
kubectl apply -f postgres.deployment.yaml
```