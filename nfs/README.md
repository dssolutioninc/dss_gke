# Build up NFS and use it in GKE
Hashimoto Du at DevSamurai

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

## 2. Build up NFS in GKE
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
# build a NFS server
kubectl apply -f nfs-container.deployment.yaml

# for access from outside of cluster, deploy a service. Skip this step if do not need access from outsite
kubectl apply -f nfs-service.deployment.yaml
```

## 3. Define PersistentVolumeClaim in GKE
```
# make PersistentVolume and PersistentVolumeClaim. Multiple nodes access concoaccessModes（ReadWriteMany）
kubectl apply -f nfs-volume.yaml
```

## 4. Make and deploy some jobs using NFS for saving data
```
# build dummy-job image on Container Registry
gcloud builds submit --config cloudbuild.dummyjob.yaml

# deploy job
kubectl apply -f deployment/dummy_job_01.deployment.yaml
kubectl apply -f deployment/dummy_job_02.deployment.yaml
```