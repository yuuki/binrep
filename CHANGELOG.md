# Changelog

## [v0.6.0](https://github.com/yuuki/binrep/compare/v0.5.1...v0.6.0) (2017-11-06)

* S3 unit testing [#10](https://github.com/yuuki/binrep/pull/10) ([yuuki](https://github.com/yuuki))
* Preserve file permission when pushing [#9](https://github.com/yuuki/binrep/pull/9) ([yuuki](https://github.com/yuuki))
* Remove sync command because it is too complicated [#8](https://github.com/yuuki/binrep/pull/8) ([yuuki](https://github.com/yuuki))
*  Skip pushing the binaries if each checksum of them is the same with one of the binaries on remote storage. [#7](https://github.com/yuuki/binrep/pull/7) ([yuuki](https://github.com/yuuki))
* Implement pull --max-bandwidth option [#6](https://github.com/yuuki/binrep/pull/6) ([yuuki](https://github.com/yuuki))
* Avoid S3 multipart download to efficiently calculate checksum  [#5](https://github.com/yuuki/binrep/pull/5) ([yuuki](https://github.com/yuuki))

## [v0.5.1](https://github.com/yuuki/binrep/compare/v0.5.0...v0.5.1) (2017-10-21)


## [v0.5.0](https://github.com/yuuki/binrep/compare/v0.4.0...v0.5.0) (2017-10-21)

* list subcommand [#3](https://github.com/yuuki/binrep/pull/3) ([yuuki](https://github.com/yuuki))

## [v0.4.0](https://github.com/yuuki/binrep/compare/v0.3.0...v0.4.0) (2017-10-20)

* Implement sync subcommand [#2](https://github.com/yuuki/binrep/pull/2) ([yuuki](https://github.com/yuuki))
* Implement cleanup some old directories when pushing [#1](https://github.com/yuuki/binrep/pull/1) ([yuuki](https://github.com/yuuki))
