smb2
====

[![Build Status](https://github.com/cloudsoda/go-smb2/actions/workflows/go.yml/badge.svg)](https://github.com/cloudsoda/go-smb2/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/cloudsoda/go-smb2.svg)](https://pkg.go.dev/github.com/cloudsoda/go-smb2)

Description
-----------

An SMB2/3 client implementation. This is a fork of the project [github.com/hirochachacha/go-smb2](https://github.com/hirochachacha/go-smb2). Any releases will be pre-1.0.0 for some time as features and bug fixes are implemented.

Installation
------------

`go get github.com/cloudsoda/go-smb2`

Documentation
-------------

https://pkg.go.dev/github.com/cloudsoda/go-smb2

Examples
--------

### List share names ###

```go
package main

import (
	"fmt"
	"net"

	"github.com/cloudsoda/go-smb2"
)

func main() {
	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     "USERNAME",
			Password: "PASSWORD",
		},
	}

	s, err := d.Dial(context.Background(), "SERVERNAME:445")
	if err != nil {
		panic(err)
	}
	defer s.Logoff()

	names, err := s.ListSharenames()
	if err != nil {
		panic(err)
	}

	for _, name := range names {
		fmt.Println(name)
	}
}
```

### File manipulation ###

```go
package main

import (
	"io"
	"net"

	"github.com/cloudsoda/go-smb2"
)

func main() {
	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     "USERNAME",
			Password: "PASSWORD",
		},
	}

	s, err := d.Dial(context.Background(), "SERVERNAME:445")
	if err != nil {
		panic(err)
	}
	defer s.Logoff()

	fs, err := s.Mount("SHARENAME")
	if err != nil {
		panic(err)
	}
	defer fs.Umount()

	f, err := fs.Create("hello.txt")
	if err != nil {
		panic(err)
	}
	defer fs.Remove("hello.txt")
	defer f.Close()

	_, err = f.Write([]byte("Hello world!"))
	if err != nil {
		panic(err)
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		panic(err)
	}

	bs, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bs))
}
```

### Check error types ###

```go
package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/cloudsoda/go-smb2"
)

func main() {
	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     "USERNAME",
			Password: "PASSWORD",
		},
	}

	s, err := d.Dial(context.Background(), "SERVERNAME:445")
	if err != nil {
		panic(err)
	}
	defer s.Logoff()

	fs, err := s.Mount("SHARENAME")
	if err != nil {
		panic(err)
	}
	defer fs.Umount()

	_, err = fs.Open("notExist.txt")

	fmt.Println(os.IsNotExist(err)) // true
	fmt.Println(os.IsExist(err))    // false

	fs.WriteFile("hello2.txt", []byte("test"), 0444)
	err = fs.WriteFile("hello2.txt", []byte("test2"), 0444)
	fmt.Println(os.IsPermission(err)) // true

	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	_, err = fs.WithContext(ctx).Open("hello.txt")

	fmt.Println(os.IsTimeout(err)) // true
}
```

### Glob and WalkDir through FS interface ###

```go
package main

import (
	"fmt"
	"net"
	iofs "io/fs"

	"github.com/cloudsoda/go-smb2"
)

func main() {
	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     "USERNAME",
			Password: "PASSWORD",
		},
	}

	s, err := d.Dial(context.Background(), "SERVERNAME:445")
	if err != nil {
		panic(err)
	}
	defer s.Logoff()

	fs, err := s.Mount("SHARENAME")
	if err != nil {
		panic(err)
	}
	defer fs.Umount()

	matches, err := iofs.Glob(fs.DirFS("."), "*")
	if err != nil {
		panic(err)
	}
	for _, match := range matches {
		fmt.Println(match)
	}

	err = iofs.WalkDir(fs.DirFS("."), ".", func(path string, d iofs.DirEntry, err error) error {
		fmt.Println(path, d, err)

		return nil
	})
	if err != nil {
		panic(err)
	}
}
```

### Authenticate with Kerberos

> [!NOTE]
> See [gokrb5 documentation](https://github.com/jcmturner/gokrb5/blob/master/v8/USAGE.md) for more details on how to initialize a Kerberos client.

```go
package main

import (
	"context"
	"fmt"

	"github.com/cloudsoda/go-smb2"
	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
)

func main() {
	cfg, err := config.Load("/etc/krb5.conf")
	if err != nil {
		panic(err)
	}

	cl := client.NewWithPassword("USERNAME", "REALM", "PASSWORD", cfg)

	d := &smb2.Dialer{
		Initiator: &smb2.Krb5Initiator{
			Client:    cl,
			TargetSPN: "cifs/SERVERNAME",
		},
	}

	s, err := d.Dial(context.Background(), "SERVERNAME:445")
	if err != nil {
		panic(err)
	}
	defer s.Logoff()

	names, err := s.ListSharenames()
	if err != nil {
		panic(err)
	}

	for _, name := range names {
		fmt.Println(name)
	}
}
```
