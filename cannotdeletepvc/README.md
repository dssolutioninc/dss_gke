# PersistentVolumeClaim (pvc)削除できず、Terminating ステータスのままとの問題

### 問題の内容
NFS を使って PersistentVolumeClaim を作成しました。
NFS 修正があって、IP が変わりました。
この状態で、PersistentVolumeClaim を削除して、再作成するつもりでしたが、
削除できず、Terminating ステータスのままとなっていました。

```sh
# PersistentVolumeClaim リストする
kubectl get persistentVolumeClaims
> Output:
> NAME              CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS     CLAIM                            STORAGECLASS   REASON   AGE
> nfs-data-volume   50Gi       RWX            Retain           Released   default/nfs-data-volume                           25d

# 対象の PersistentVolumeClaim を削除します
kubectl delete persistentVolumeClaims nfs-data-volume
# ずっと処理が完了できず、Ctrl+C でキャンセルした

# PersistentVolumeClaim を再確認して、Terminating ステータスのままとなっています
kubectl get persistentVolumeClaims
> Output:
> NAME              STATUS        VOLUME            CAPACITY   ACCESS MODES   STORAGECLASS   AGE
> nfs-data-volume   Terminating   nfs-data-volume   50Gi       RWX                           25d
```

### 解決方法
原因はPersistentVolumeClaimがProtected状態になってしまい、削除できなくなります。

kubectl patch コマンドでProtected状態をクリアしたら、解決できます。

```sh
# Protected状態の確認
kubectl describe pvc nfs-data-volume | grep Finalizers
> Output:
> Finalizers:    [kubernetes.io/pvc-protection]

# Protected状態をクリアする
kubectl patch pvc nfs-data-volume -p '{"metadata":{"finalizers": []}}' --type=merge

# 削除したい PersistentVolumeClaim を確認
kubectl get persistentVolumeClaims
> Output: なし（nfs-data-volumeが削除された）
```

<br> 
記事のご覧、どうもありがとうございます！
DevSamurai 橋本