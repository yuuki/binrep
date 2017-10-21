Binrep
======

[![Build Status](https://travis-ci.org/yuuki/binrep.png?branch=master)][travis]
[![Go Report Card](https://goreportcard.com/badge/github.com/yuuki/droot)][goreportcard]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]

[travis]: https://travis-ci.org/yuuki/binrep
[goreportcard]: (https://goreportcard.com/report/github.com/yuuki/binrep)
[license]: https://github.com/yuuki/binrep/blob/master/LICENSE

Binrep is the static binary repository manager that uses Amazon S3 as file storage.

# Overview

`binrep` provides a simple way to store and deploy static binary files such as Go binaries, tarballs, images and web assets(css/javascript). `binrep` just pushes static binary files into a S3 bucket with directory layouts like `go get` does, and pulls the binaries from the bucket.

The deployment of (private) tools written by Go takes a lot more works, especially in the environment that has many servers, than that of LL scripts in a single file such as shell script, Perl, Python, Ruby). Git is an informat approach to the deployment, but git is not for binary management. The next approach is a package manager such apt or yum, but it takes a lot of trouble to make packages. The other approach is just to use a HTTP file server including S3, but it needs uniform and accessible location of binary files and the version management. There, `binrep` resolves the problems of static binary management.

## Features

The following features will be supported.

- Amazon S3 as storage backend
- directory layout `<host>/<user>/<project>` like `go get` and [ghq](https://github.com/motemen/ghq) do
- version management like [Capistrano](http://capistranorb.com/)
- synchronization of all latest binary files from S3 to a local filesystem

The following features will **not** be supported.

- dependency management, that is supported by popular package managers such as Apt, Yum and Rubygems

# Usage

## Prepareation

- Create S3 bucket for `binrep`.
- Install `binrep` binary, see https://github.com/yuuki/binrep/releases .

### Getting the latest version

```sh
$ curl -fsSL https://raw.githubusercontent.com/yuuki/binrep/master/scripts/install_latest_binary | bash /dev/stdin $GOOS $GOARCH | tar --exclude 'README.md' --exclude 'LICENSE' -xzf - -C /usr/local/bin/
```

- GOOS: 'linux' or 'darwin'
- GOARCH: '386' or 'amd64'

## Commands

### list

```sh
$ binrep list --endpoint s3://binrep-bucket
github.com/fujiwara/stretcher/20171013135903/
github.com/fujiwara/stretcher/20171014110009/
github.com/motemen/ghq/20171013140424/
github.com/yuuki/droot/20171017152626/
github.com/yuuki/droot/20171018125535/
github.com/yuuki/droot/20171019204009/
...
```

### show

```sh
$ binrep show --endpoint s3://binrep-bucket github.com/yuuki/droot
NAME                    TIMESTAMP       BINNARY1
github.com/yuuki/droot  20171019204009  droot//2e6ccc3
```

### push

```sh
$ binrep push --endpoint s3://binrep-bucket github.com/yuuki/droot ./droot
--> Uploading [./droot] to s3://binrep-bucket/github.com/yuuki/droot/20171020152356
Uploaded to s3://binrep-bucket/github.com/yuuki/droot/20171020152356
--> Cleaning up the old releases
Cleaned up 20171017152626
```

`push` supports to push multiple binary files.

### pull

```sh
$ binrep pull --endpoint s3://binrep-bucket github.com/yuuki/droot /usr/local/bin
--> Downloading s3://binrep-bucket/github.com/yuuki/binrep/20171019204009 to /usr/local/bin
```

### sync

```sh
$ binrep sync --endpoint s3://binrep-bucket --concurrency 4 --max-bandwidth '5 MB' /opt/binrep/
Set max bandwidth total: 10 MB/sec, per-release: 2.5 MB/sec
--> Downloading to /opt/binrep/github.com/fujiwara/stretcher/20171014110009/
--> Downloading to /opt/binrep/github.com/motemen/ghq/20171013140424/
--> Downloading to /opt/binrep/github.com/yuuki/droot/20171019204009/
...
```

`sync` skips the download if there are already the same timestamp release on local filesystem.

# Directory layout on S3 bucket

```
s3://<bucket>/<host>/<user>/<project>/<timestamp>/
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

# Terms

- `release`: `<host>/<user>/<project>/<timestamp>/`

# How to release

binrep uses the following tools for the artifact release.

- [goreleaser](https://goreleaser.com/)
- [gobump](https://github.com/motemen/gobump)
- [ghch](https://github.com/Songmu/ghch)

```sh
make release
```

# License

[The MIT License](./LICENSE).

# Author

[y_uuki](https://github.com/yuuki)
