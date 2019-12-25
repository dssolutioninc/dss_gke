# CentOS7で libgeotiff パッケージインストールのドッカーイメージビルドエラー

### 問題の内容
CentOS7を使って写真処理のパッケージをDockerイメージでビルドします。

```sh: Dockerfile
FROM centos:7
RUN yum update -y && yum install -y \
                    curl \
                    libgeotiff \
                    which && \
    yum clean all
```

ビルドすると、「No package libgeotiff available」エラーが出ています。

### 解決方法
原因は　libgeotiff パッケージは　epel レポジトリにあり、CentOS7イメージのデフォルトはこのレポジトリが登録されていない。

解決方法は　epel レポジトリを登録するか、libgeotiff パッケージのURLの直接でインストールする方法となります。
レポジトリを登録する場合、yum-config-manager コマンドが必要となり、yum-utils パッケージをインストルしないといけない。
Dockerイメージのサイズを最低化するため、パッケージのURLを直接でインストールする方法とします。

パッケージのバージョンを選定して、パッケージダウンロードURLに修正します。
libgeotiff パッケージは libproj パッケージを利用するため、同時にインストールする必要です。
最終の Dockerfileはこれとなります。

```sh: Dockerfile
FROM centos:7
RUN yum update -y && yum install -y \
                    curl \
                    http://download-ib01.fedoraproject.org/pub/epel/7/x86_64/Packages/p/proj-4.8.0-4.el7.x86_64.rpm \
                    http://download-ib01.fedoraproject.org/pub/epel/7/x86_64/Packages/l/libgeotiff-1.2.5-14.el7.x86_64.rpm \
                    which && \
    yum clean all
```

<br> 
記事のご覧、どうもありがとうございます！
DevSamurai 橋本