Refer
https://medium.com/platformer-blog/nfs-persistent-volumes-with-kubernetes-a-case-study-ce1ed6e2c266


https://github.com/mappedinn/kubernetes-nfs-volume-on-gke


### Deploy Workloads


```
# create persistent disk
gcloud compute disks create --size=10GB --zone=asia-northeast1-b dataprocessing-gce-nfs-disk


# create vm attaches the disk
gcloud compute instances create dataprocessing-vm --zone=asia-northeast1-b
       
gcloud compute instances create dataprocessing-vm --zone=asia-northeast1-b --machine-type=f1-micro  --disk=name=dataprocessing-gce-nfs-disk


gcloud compute ssh --zone=asia-northeast1-b dataprocessing-vm


duluong@dataprocessing-vm:~$ sudo lsblk
NAME   MAJ:MIN RM SIZE RO TYPE MOUNTPOINT
sda      8:0    0  10G  0 disk 
`-sda1   8:1    0  10G  0 part /
sdb      8:16   0  10G  0 disk 


sudo mkfs.ext4 -m 0 -F -E lazy_itable_init=0,lazy_journal_init=0,discard /dev/[DEVICE_ID]


sudo mkdir -p /mnt/disks/sdb

sudo mount -o discard,defaults /dev/sdb /mnt/disks/sdb

# Configure read and write permissions on the device. For this example, grant write access to the device for all users.
sudo chmod a+w /mnt/disks/sdb

```



```
# create persistent volume
kubectl apply -f pvc-demo.yaml
```

accessModes:
    - ReadWriteOnce
    - ReadOnlyMany
    - ReadWriteMany
