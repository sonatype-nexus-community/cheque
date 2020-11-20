<p align="center">
    <img src="https://github.com/sonatype-nexus-community/cheque/blob/master/docs/images/Cheque-beaver.png" width="350"/>
</p>

<p align="center">
    <a href="https://circleci.com/gh/sonatype-nexus-community/cheque"><img src="https://circleci.com/gh/sonatype-nexus-community/cheque.svg?style=svg" alt="Circle CI Build Status"></img></a>
    <!--
    <a href="https://gitter.im/sonatype-nexus-community/cheque?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge"><img src="https://badges.gitter.im/sonatype-nexus-community/cheque.svg" alt="Gitter"></img></a>
    -->
</p>

# cheque

Like wearing a toque in the winter, ensuring your software is secure should be
second nature, eh.

`Cheque` helps you by finding all libraries used by your C/C++ projects, from A to Zed,
and retrieving known vulnerabilities from [OSS Index](https://ossindex.sonatype.org/).
This process saves you a significant amount of labour and time, which is much
better spent playing hockey, slamming back a two-four, drinking a double-double,
or pretty much anything else rather then poring through obtuse lists and
reading Makefiles.

And its completely free! Beauty, eh?

## Supported platforms

Currently only gcc on Linux is supported. There is marginal support for other
compilers (clang) and operating systems (osx) but it is far more rudimentary and
not to be trusted.

## Usage

```
Usage: cheque [options] <filename> ...

When you invoke cheque, it identifies static and dynamic library dependencies
and identifies known vulnerabilities using the OSS Index vulnerability database.

Cheque can be used as a wrapper around the compiler/linker by making symbolic
links to cheque with the compiler name, and ensuring they are in the front of
your PATH. Cheque will run, and also execute the compiler/linker appropriately.
This allows cheque to be embedded in most builds.

Option summary: (Many cheque options match those of the underlying compiler/linker)
  -L<dir>
    	Add the specified directory to the front of the library search path
  -l<library>
    	Specify the name of a DLL required for compiling/linking
  -Werror=cheque
    	Treat cheque warnings as errors
  -version
    	prints current cheque version
```

For example:

```
> cheque -lpng hello.c

[1/3] rpm/fedora/libglibc@2.17    No known vulnerabilities against package/version...
------------------------------------------------------------
[2/3] conan/bincrafters/libpng@1.2.49  [Vulnerable]    9 known vulnerabilities affecting installed version

[CVE-2016-3751] Unspecified vulnerability in libpng before 1.6.20, as used in Android 4.x before...
Unspecified vulnerability in libpng before 1.6.20, as used in Android 4.x before 4.4.4, 5.0.x before 5.0.2, 5.1.x before 5.1.1, and 6.x before 2016-07-01, allows attackers to gain privileges via a crafted application, as demonstrated by obtaining Signature or SignatureOrSystem access, aka internal bug 23265085.

ID: 79196806-d4cd-4730-8ca4-38692ad2b8b6
Details: https://ossindex.sonatype.org/vuln/79196806-d4cd-4730-8ca4-38692ad2b8b6

[CVE-2015-8126]  Improper Restriction of Operations within the Bounds of a Memory Buffer
Multiple buffer overflows in the (1) png_set_PLTE and (2) png_get_PLTE functions in libpng before 1.0.64, 1.1.x and 1.2.x before 1.2.54, 1.3.x and 1.4.x before 1.4.17, 1.5.x before 1.5.24, and 1.6.x before 1.6.19 allow remote attackers to cause a denial of service (application crash) or possibly have unspecified other impact via a small bit-depth value in an IHDR (aka image header) chunk in a PNG image.

ID: 3e2ddc24-dd11-47cf-a3e7-93710a0eab7f
Details: https://ossindex.sonatype.org/vuln/3e2ddc24-dd11-47cf-a3e7-93710a0eab7f
...
```

## Wrapping the compiler/linker

Cheque can be used as a wrapper around the compiler/linker by making symbolic
links to cheque with the compiler name, and ensuring they are in the front of
your PATH. Cheque will run, and also execute the compiler/linker appropriately.
This allows cheque to be embedded in most builds.

For example:

```
> pwd
/path/to/links

>  ls -l
total 0
lrwxrwxrwx 1 ec2-user ec2-user 27 Dec 30 01:48 cc -> /path/to/cheque
lrwxrwxrwx 1 ec2-user ec2-user 27 Oct 17 14:39 g++ -> /path/to/cheque
lrwxrwxrwx 1 ec2-user ec2-user 27 Oct 17 14:39 gcc -> /path/to/cheque
lrwxrwxrwx 1 ec2-user ec2-user 27 Oct 17 14:39 ld -> /path/to/cheque

> export PATH=`pwd`:$PATH

> cd /path/to/test

> gcc -lpng hello.c

[1/3] rpm/fedora/libglibc@2.17    No known vulnerabilities against package/version...
------------------------------------------------------------
[2/3] conan/bincrafters/libpng@1.2.49  [Vulnerable]    9 known vulnerabilities affecting installed version
...
```


## Configuration

There are two configuration files. They can be found in either:
* ~/.cheque/config
* .cheque-config in the current working directory.

Both configs offer the same options, however, the current working directory config
will override the global config.

See .cheque-config.sample for the format 

### IQ Server

If cheque is configured to use and submit reports to iq server, 
it requires the .iq-server-config file to be configured and placed 
in the following directory:

`~/.iqserver/.iq-server-config` 

See the sample file `.iq-server-config.sample`

### Oss Index

If there is interest to run against Oss Index with credentials,
it requires the .oss-index-config file to be configured and placed 
in the following directory:

`~/.ossindex/.oss-index-config` 

See the sample file `.oss-index-config.sample`

## Development

`cheque` is written using Golang 1.14, so it is best you start there.

Tests can be run like `go test ./... -v`

## Contributing

We care a lot about making the world a safer place, and that's why we created `cheque`. If you as well want to
speed up the pace of software development by working on this project, jump on in! Before you start work, create
a new issue, or comment on an existing issue, to let others know you are!

## The Fine Print

It is worth noting that this is **NOT SUPPORTED** by Sonatype, and is a contribution of ours
to the open source community (read: you!)

Remember:

* Use this contribution at the risk tolerance that you have
* Do NOT file Sonatype support tickets related to `cheque` support in regard to this project
* DO file issues here on GitHub, so that the community can pitch in

Phew, that was easier than I thought. Last but not least of all:

Have fun creating and using `cheque` and the [Sonatype OSS Index](https://ossindex.sonatype.org/), we are glad to have you here!
