binrep
======

Simple static binary repository manager on Amazon S3 as backend storage.

# Overview

`binrep` provides a simple way to store and deploy your static binary files such as Go binary. binrep pushes the binary files into your S3 bucket, builds the directory layout like `go get` does on the bucket, and pulls the binary files from the bucket.

The deployment of (internel) tools written by Go takes a lot more works, especially in the environment with many servers, than that of no-dependent LL scripts such as shell script, Perl, Python, Ruby). Git is an informat approach to the deployment, but git is not for binary management. The next approach is a package manager such apt or yum, but it takes a lot of trouble to make packages. The other approach is just to use a HTTP file server including S3, but it needs uniform and accessible location of the binary files and version management. There, `binrep` resolves the problem of the binary management.

# Usage

## Prepareation

- Create S3 bucket for `binrep`.
- Install `binrep` binary, see https://github.com/yuuki/binrep/releases .

## Commands

```
binrep ls --endpoint s3://<bucket> github.com/yuuki/droot
binrep push --endpoint s3://<bucket> github.com/yuuki/droot ./droot
binrep pull --endpoint s3://<bucket> github.com/yuuki/droot /usr/local/bin
binrep sync --endpoint s3://<bucket> /usr/local/bin
```

# Directory layout on S3 bucket

```
s3://<bucket>/<host>/<user>/<project>/<date>/
                                         -- <bin>
                                         -- meta.yml
```

The example below.

```
s3://binrep-repository/
|-- github.com/
    -- yuuki/
        -- droot/
            -- 20171013080011/
                -- droot
                -- meta.yml
            -- 20171014102929/
                -- droot
                -- meta.yml
    -- prometheus/
        -- prometheus/
            -- 20171012081234/
                -- prometheus
                -- promtool
                -- meta.yml
|-- ghe.internal/
    -- opsteam/
        -- tools
            -- 20171010071022/
                -- ec2_bootstrap
                -- ec2_build_ami
                -- mysql_create_slave_snapshot

```

# License

[The MIT License](./LICENSE).

# Author

[y_uuki](https://github.com/yuuki)
