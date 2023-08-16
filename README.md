# go-pop3-socks5

[![GoDoc](https://godoc.org/github.com/alfonmga/go-pop3-socks5?status.svg&style=flat-square)](http://godoc.org/github.com/alfonmga/go-pop3-socks5)

A simple Go POP3 client library for connecting and reading mails from POP3 servers. This is a full rewrite of [TheCreeper/go-pop3](https://github.com/TheCreeper/go-pop3) with bug fixes and new features.

## Fork changes

- [x] Add SOCKS5 proxy support.

## Install

`go get -u github.com/alfonmga/go-pop3-socks5`

## Example

```go
import (
	"fmt"
	"github.com/alfonmga/go-pop3-socks5"
)

func main() {
	// Initialize the client.
	p := pop3.New(pop3.Opt{
		Host: "pop.gmail.com",
		Port: 995,
		TLSEnabled: true,
		TLSSkipVerify:   true,
		Socks5ProxyAddr: "user:password@host:port",
	})

	// Create a new connection. POP3 connections are stateful and should end
	// with a Quit() once the opreations are done.
	c, err := p.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	defer c.Quit()

	// Authenticate.
	if err := c.Auth("myuser", "mypassword"); err != nil {
		log.Fatal(err)
	}

	// Print the total number of messages and their size.
	count, size, _ := c.Stat()
	fmt.Println("total messages=", count, "size=", size)

	// Pull the list of all message IDs and their sizes.
	msgs, _ := c.List(0)
	for _, m := range msgs {
		fmt.Println("id=", m.ID, "size=", m.Size)
	}

	// Pull all messages on the server. Message IDs go from 1 to N.
	for id := 1; id <= count; id++ {
		m, _ := c.Retr(id)

		fmt.Println(id, "=", m.Header.Get("subject"))

		// To read the multi-part e-mail bodies, see:
		// https://github.com/emersion/go-message/blob/master/example_test.go#L12
	}

	// Delete all the messages. Server only executes deletions after a successful Quit()
	for id := 1; id <= count; id++ {
		c.Dele(id)
	}
}
```

[![PkgGoDev](https://pkg.go.dev/badge/github.com/alfonmga/go-pop3-socks5)](https://pkg.go.dev/github.com/alfonmga/go-pop3-socks5)

### To-do: tests

Setup a Docker test environment that runs [InBucket](https://github.com/inbucket/inbucket) POP3 + SMTP server to run a dummy POP3 server and test all the commands in the lib.

Licensed under the MIT License.
