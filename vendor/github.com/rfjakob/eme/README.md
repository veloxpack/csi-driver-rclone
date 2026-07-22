EME for Go [![CI](https://github.com/rfjakob/eme/actions/workflows/ci.yml/badge.svg)](https://github.com/rfjakob/eme/actions/workflows/ci.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/rfjakob/eme.svg)](https://pkg.go.dev/github.com/rfjakob/eme) ![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)
==========

**EME** (ECB-Mix-ECB or, clearer, **Encrypt-Mix-Encrypt**) is a wide-block
encryption mode developed by Halevi
and Rogaway in 2003 [[eme]](#eme).

EME uses multiple invocations of a block cipher to construct a new
cipher of bigger block size (in multiples of 16 bytes, up to 2048 bytes).

Key advantages:

* Deterministic encryption - no random nonce
* Zero overhead - zero ciphertext expansion
  * But this also means there is no authentication! The application
    must validate decrypted data.
* True wide-block encryption. Changing a single bit randomizes the
  whole ciphertext.
  * This is not the case for AES-XTS, for example.

Quoting from the original [[eme]](#eme) paper:

> We describe a block-cipher mode of operation, EME, that turns an n-bit block cipher into
> a tweakable enciphering scheme that acts on strings of mn bits, where m ∈ [1..n]. The mode is
> parallelizable, but as serial-efficient as the non-parallelizable mode CMC [6]. EME can be used
> to solve the disk-sector encryption problem. The algorithm entails two layers of ECB encryption
> and a “lightweight mixing” in between. We prove EME secure, in the reduction-based sense of
> modern cryptography.

Figure 2 from the [[eme]](#eme) paper shows an overview of the transformation:

[![Figure 2 from [eme]](paper-eme-fig2.png)](#)

This is an implementation of EME in Go, complete with test vectors from IEEE [[p1619-2]](#p1619-2)
and Halevi [[eme-32-testvec]](#eme-32-testvec).

It has no dependencies outside the standard library.

Is it patentend?
----------------

In 2007, the UC Davis has decided to abandon [[patabandon]](#patabandon)
the patent application [[patappl]](#patappl) for EME.

Related algorithms
------------------

**EME-32** is EME with the cipher set to AES and the length set to 512.
That is, EME-32 [[eme-32-pdf]](#eme-32-pdf) is a subset of EME.

**EME2**, also known as EME\* [[emestar]](#emestar), is an extended version of EME
that has built-in handling for data that is not a multiple of 16 bytes
long.  
EME2 has been selected for standardization in IEEE P1619.2 [[p1619.2]](#p1619.2).

How to use
----------

Ho to use the eme in you app: Use `eme.New`: https://pkg.go.dev/github.com/rfjakob/eme#New

How to run the self-tests:

	$ go test -v
	=== RUN   TestEnc512
	--- PASS: TestEnc512 (0.00s)
	=== RUN   TestEnc512x100
	--- PASS: TestEnc512x100 (0.00s)
	=== RUN   TestDec512
	--- PASS: TestDec512 (0.00s)
	=== RUN   TestDec512x100
	--- PASS: TestDec512x100 (0.00s)
	=== RUN   TestEnc16
	--- PASS: TestEnc16 (0.00s)
	=== RUN   TestEnc2048
	--- PASS: TestEnc2048 (0.00s)
	PASS
	ok  	github.com/rfjakob/eme	0.005s

How to run the benchmarks:

	$ ./benchmark.bash
	goos: linux
	goarch: amd64
	pkg: github.com/rfjakob/eme
	cpu: Intel(R) Core(TM) i5-3470 CPU @ 3.20GHz
	BenchmarkEnc512-4   	  142634	      8161 ns/op	  62.73 MB/s
	BenchmarkDec512-4   	  140930	      8245 ns/op	  62.09 MB/s
	PASS
	ok  	github.com/rfjakob/eme	2.508s


Other implementations
---------------------

Some other EME implementations I am aware of (PRs for more accepted):

* JavaScript (library): https://github.com/fyears/eme / https://www.npmjs.com/package/@fyears/eme
* C++ (part of cppcryptfs): https://github.com/bailey27/cppcryptfs/blob/master/libcppcryptfs/crypt/eme.cpp
* Rust (library): https://github.com/B4dM4n/eme-mode-rs / https://crates.io/crates/eme-mode
* Python (part of gocryptfs-inspect): https://github.com/maxpat78/gocryptfs-inspect/blob/master/aes256eme.py
* Java (library): https://github.com/alexey-lapin/eme-java

References
----------

#### [eme]
*A Parallelizable Enciphering Mode*  
Shai Halevi, Phillip Rogaway, 28 Jul 2003  
https://eprint.iacr.org/2003/147.pdf ([archive.org snapshot](https://web.archive.org/web/20210506160350/https://eprint.iacr.org/2003/147.pdf))

Note: This is the original EME paper. EME is specified for an arbitrary
number of block-cipher blocks. EME-32 is a concrete implementation of
EME with a fixed length of 32 AES blocks.

#### [eme-32-email]
*Re: EME-32-AES with editorial comments* (announcement email with [[eme-32-pdf]](#eme-32-pdf) attachments)  
Shai Halevi, 07 Jun 2005  
~~http://grouper.ieee.org/groups/1619/email/msg00310.html~~ (broken link as of July 2021, [archive.org snapshot](http://web.archive.org/web/20081227091850/http://grouper.ieee.org/groups/1619/email/msg00310.html)) (attachment: `EME-32-AES-Jun-05.pdf`=`pdf00020.pdf`),  
~~http://grouper.ieee.org/groups/1619/email/msg00309.html~~
(broken link as of July 2021, [archive.org snapshot](http://web.archive.org/web/20081228013334/http://grouper.ieee.org/groups/1619/email/msg00309.html))
(attachment: `EME-32-AES-Jun-05.doc`=`doc00011.doc`)

#### [eme-32-pdf]
*Draft Standard for Tweakable Wide-block Encryption*, `EME-32-AES-Jun-05.{doc,pdf}`
Shai Halevi, 02 June 2005  
`EME-32-AES-Jun-05.pdf` ... ~~http://grouper.ieee.org/groups/1619/email/pdf00020.pdf~~ (broken link as of July 2021, no archive.org snapshot available)  
`EME-32-AES-Jun-05.doc` ... ~~http://grouper.ieee.org/groups/1619/email/doc00011.doc~~ (broken link as of July 2021, [archive.org snapshot of external mirror](http://web.archive.org/web/20210701125726/https://samifar.in/code/crypto/eme-32-aes/doc00011.doc))

Note: This is the latest version of the EME-32 draft that I could find as of Dec 2015. It
includes test vectors and C source code.

#### [eme-32-testvec]
*Re: Test vectors for LRW and EME*  
Shai Halevi, 16 Nov 2004  
~~http://grouper.ieee.org/groups/1619/email/msg00218.html~~ (broken link as of July 2021, [archive.org snapshot](https://web.archive.org/web/20070305060551/http://grouper.ieee.org/groups/1619/email/msg00218.html))

#### [emestar]
*EME\*: extending EME to handle arbitrary-length messages with associated data*  
Shai Halevi, 27 May 2004  
https://eprint.iacr.org/2004/125.pdf ([archive.org snapshot](https://web.archive.org/web/20160826083914/http://eprint.iacr.org/2004/125.pdf))

#### [patabandon]
*Re: [P1619-2] Non-awareness patent statement made by UC Davis*  
Mat Ball, 26 Nov 2007  
~~http://grouper.ieee.org/groups/1619/email-2/msg00005.html~~ (broken link as of July 2021, [archive.org snapshot](https://web.archive.org/web/20110611145815/http://grouper.ieee.org/groups/1619/email-2/msg00005.html))

#### [patappl]
*Block cipher mode of operation for constructing a wide-blocksize block cipher from a conventional block cipher*  
US patent application US20040131182  
http://www.google.com/patents/US20040131182

#### [p1619-2]
*IEEE P1619.2™/D9 Draft Standard for Wide-Block Encryption for Shared Storage Media*  
IEEE, Dec 2008  
~~http://siswg.net/index2.php?option=com_docman&task=doc_view&gid=156&Itemid=41~~ (broken link as of July 2021, [archive.org snapshot](https://web.archive.org/web/20171018232831/http://siswg.net/index2.php?option=com_docman&task=doc_view&gid=156&Itemid=41))

Note: This is a draft version. The final version is not freely available
and must be [bought from IEEE](https://ieeexplore.ieee.org/document/5729263).

Package Changelog
-----------------

v1.2.0, 2026-01-25
* Convert if conditions in multByTwo() to constant-time operations
  ([#8](https://github.com/rfjakob/eme/issues/8))
* Update broken links in README, add links to other EME implementions
* Switch to `subtle.XORBytes` (45802ed4fb60f60)

v1.1.2, 2021-06-27
* Add `go.mod` file
* Switch from Travis CI to Github Actions
* No code changes

v1.1.1, 2020-04-13
* Update `go vet` call in `test.bash` to work on recent Go versions
* No code changes

v1.1, 2017-03-05
* Add eme.New() / \*EMECipher convenience wrapper
* Improve panic message and parameter wording

v1.0, 2015-12-08
* Stable release
