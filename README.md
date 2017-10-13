sbrepo
======

static binary repository.

# Usage

```
sbrepo ls --endpoint s3://<bucket> github.com/yuuki/droot
sbrepo push --endpoint s3://<bucket> github.com/yuuki/droot ./droot
sbrepo pull --endpoint s3://<bucket> github.com/yuuki/droot /usr/local/bin/droot
sbrepo sync --endpoint s3://<bucket> /usr/local/bin
```

# Repository Structure

```
s3://<bucket>/<host>/<user>/<project>/<date>/<bin>_<os>_<arch>
s3://<bucket>/<host>/<user>/<project>/<date>/<bin>_VERSION (Option)
s3://<bucket>/<host>/<user>/<project>/<date>/<bin>_CHECKSUM (Option)
```
