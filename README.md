sbrepo
======

static binary repository.

# Usage

```
sbrepo push --name droot --endpoint s3://<bucket> --version 0.8.0 ./droot
sbrepo pull --name droot --endpoint s3://<bucket> --version 0.8.0 --location /usr/local/bin
sbrepo sync --endpoint s3://<bucket> --location /usr/local/bin
```

install latest version.

```
sbrepo pull --name droot --endpoint s3://<bucket> /usr/local/bin
```

# Repository Structure

```
s3://<bucket>/<name>/<version>/<binname>
```
