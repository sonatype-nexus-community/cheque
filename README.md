cheque
======

Like wearing a toque in the winter, ensuring your software is secure should be
second nature, eh.

`Cheque` helps you by finding all libraries used by your C/C++ projects, from A to Zed,
and retrieving known vulnerabilities from [OSS Index](https://ossindex.sonatype.org/).
This process saves you a significant amount of labour and time, which is much
better spent playing hockey, slamming back a two-four, drinking a double-double,
or pretty much anything else rather then poring through obtuse lists and
reading Makefiles.

And its completely free! Beauty, eh?

WARNING: This code is currently in development, and will almost certainly not
do what you want it to; it is pretty hosed. Sorry.



Limitations
===========

OSX
-------------
* otool does not give us "real" version numbers (as in version numbers used by the source)
* Not all libraries have reasonable version numbers on the file or in path
* OSX version numbers on library files often have no relation to the source version


Unix
--------------
* Inconsistent naming: Some libraries provide strange version variations,
  for example smashing major/minor revisions together. We can work around this
  issue (though perhaps not perfectly)

    libpng16.so.16.26.0  ==>  libpng16.so.1.6.26.0


Windows
---------------
* Doesn't work on Windows yet


Future improvements
===================

* Get it working properly on Linux
* On Linux, detect distribution and query the package manager for the file
  owner (we can better target vulnerabilities by specifically querying using
  the exact package's name)
* ...?

* Get it working on Windows
