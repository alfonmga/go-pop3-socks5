package pop3

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/smtp"
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
)

const MSG = `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do.`

// n represents number of messages to add to the testuser's inbox
func add_messages(n int) error {
	to := []string{"recipient@example.net"}
	msgs := make([][]byte, 0)
	for i := 0; i < 5; i++ {
		to := "To: recipient@example.net\r\n"
		subject := fmt.Sprintf("Subject: Subject %d\r\n", i)
		mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
		body := fmt.Sprintf("Message %d.\r\n"+MSG+"\r\n", i)
		msg := []byte(to + subject + mime + body)
		msgs = append(msgs, msg)
	}

	for _, msg := range msgs {
		err := smtp.SendMail("localhost:2500", nil, "sender@example.org", to, msg)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func getConnection() (*Conn, error) {
	p := New(Opt{
		Host:       "localhost",
		Port:       1100,
		TLSEnabled: false,
	})

	c, err := p.NewConn()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func readAndCompareMessageBody(m *message.Entity, msg string) error {
	mr := mail.NewReader(m)
	if mr != nil {
		// This is a multipart message
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}
			b, err := io.ReadAll(p.Body)
			if err != nil {
				return err
			}
			if !strings.EqualFold(string(b), msg) {
				return fmt.Errorf("expected message body:\n%sreceived:\n%s", msg, string(b))
			}
		}
		return nil
	} else {
		t, _, _ := m.Header.ContentType()
		log.Println("This is a non-multipart message with type", t)
		return nil
	}
}

func TestAll(t *testing.T) {

	c, err := getConnection()
	if err != nil {
		t.Fatal("error establishing connection to pop3 server ", err)
	}

	err = add_messages(5)
	if err != nil {
		t.Fatal("unable to send messages to the mail server", err)
	}

	// testing Auth
	if err := c.Auth("recipient", "password"); err != nil {
		t.Fatal(err)
	}

	// testing Stat
	count, size, err := c.Stat()
	if err != nil {
		t.Fatal("error using Stat", err)
	}
	log.Printf("count: %d, size: %d\n", count, size)

	// testing Uidl
	msgIds, err := c.Uidl(0)
	if err != nil {
		t.Fatal("error using Uidl(0)", err)
	}

	if len(msgIds) != count {
		t.Fatalf("Uidl returned: %d number of messages, but actually there are %d messages\n", len(msgIds), 5)
	}

	msgId, err := c.Uidl(msgIds[0].ID)
	if err != nil {
		t.Fatal("error using Uidl for positive message ID", err)
	}
	if len(msgId) != 1 {
		t.Fatalf("Uidl returns a list of (message ID, message UID) pairs. If the optional msgID is > 0, then only that particular message is listed but it returned %d pair\n", len(msgId))
	}

	// testing List
	msgs, err := c.List(0)
	if err != nil {
		t.Fatal("error using List(0)", err)
	}
	if len(msgs) != 5 {
		t.Fatalf("List(0) returned incorrect number of messages got: %d actual: %d\n", len(msgs), 5)
	}
	msgId, err = c.List(msgs[1].ID)
	if err != nil {
		t.Fatal("error using List for positive message ID", err)
	}
	if len(msgId) != 1 {
		t.Fatalf("List returns a list of (message ID, message UID) pairs. If the optional msgID is > 0, then only that particular message is listed but it returned %d pair\n", len(msgId))
	}

	// testing Retr
	m, err := c.Retr(msgs[0].ID)
	if err != nil {
		t.Fatal("error using Retr", err)
	}
	if m.Header.Get("subject") != "Subject 0" {
		t.Fatalf("Retr returned wrong subject returned: %s, expected: Subject 0 ", m.Header.Get("subject"))
	}
	err = readAndCompareMessageBody(m, "Message 0.\r\n"+MSG+"\r\n")
	if err != nil {
		t.Fatal(err)
	}

	// testing RetrRaw
	mb, err := c.RetrRaw(msgs[0].ID)
	if err != nil {
		t.Fatal("error using RetrRaw", err)
	}
	b := mb.Bytes()
	if !bytes.Contains(b, []byte("Message 0.\r\n"+MSG+"\r\n")) {
		t.Fatalf("expected message body:\n%s, received:\n%s", "Message 0.\r\n"+MSG+"\r\n", string(b))
	}

	// testing Top
	m, err = c.Top(msgs[0].ID, 1)
	if err != nil {
		t.Fatal("error using Top", err)
	}
	err = readAndCompareMessageBody(m, "Message 0.\r\n")
	if err != nil {
		t.Fatal(err)
	}

	// testing Noop
	err = c.Noop()
	if err != nil {
		t.Fatal("error in using Noop", err)
	}

	// testing Dele
	err = c.Dele([]int{1, 2}...)
	if err != nil {
		t.Fatal("error using Dele", err)
	}
	msgs, _ = c.List(0)
	if len(msgs) != 3 {
		t.Fatalf("after deleting 2 messages number of messages in inbox should be 3 but got %d", len(msgs))
	}
	// testing Rset, list
	err = c.Rset()
	if err != nil {
		t.Fatal("error using Rset", err)
	}
	msgs, _ = c.List(0)
	if len(msgs) != 5 {
		t.Fatalf("after Rseting number of messages in inbox should be 5 but got %d", len(msgs))
	}

	// testing Quit
	err = c.Quit()
	if err != nil {
		t.Fatal("error using Quit method", err)
	}
}

func TestSocks5(t *testing.T) {
	t.Skip("Add your own SOCKS5 proxy to test this")

	p := New(Opt{
		Host:            "outlook.office365.com",
		Port:            995,
		Socks5ProxyAddr: "", // <- Add here
		TLSSkipVerify:   true,
		TLSEnabled:      true,
		DialTimeout:     1 * time.Second,
	})
	c, err := p.NewConn()
	if err != nil {
		t.Fatalf("error establishing connection to pop3 server %s", err)
	}
	defer c.Quit()

	// POP3 authentication
	if err := c.Auth("", ""); err != nil {
		t.Fatalf("authentication failed: %s", err)
	}
	count, _, err := c.Stat()
	if err != nil {
		t.Fatalf("failed to get mailbox status: %s", err)
	}
	t.Logf("mailbox contains %d messages", count)
}
