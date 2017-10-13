sbrepo
======

static binary repository.

# Usage

```
sbrepo push --name droot --endpoint s3://<bucket> --version 0.8.0 ./droot
sbrepo pull --name droot --endpoint s3://<bucket> --version 0.8.0 /usr/local/bin/droot
sbrepo sync --endpoint s3://<bucket> /usr/local/bin
```

install latest version.

```
sbrepo pull --name droot --endpoint s3://<bucket> /usr/local/bin
```

# Repository Structure

```
s3://<bucket>/<name>/<version>/<binname>
```
