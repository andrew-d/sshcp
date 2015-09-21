package sshcp

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/ssh"
)

// Conn represents an open SSH connection writing to a given file.  It
// implements the io.WriteCloser interface, and should be closed once copying
// is complete.
type Conn struct {
	client  *ssh.Client
	session *ssh.Session
	writer  io.WriteCloser
}

// NewConn creates a new connection to the remote host, writing to a file named
// the given name.
func NewConn(host, username, remoteName string, authMethods ...ssh.AuthMethod) (*Conn, error) {
	if !strings.Contains(host, ":") {
		host = host + ":22"
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: authMethods,
	}
	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, fmt.Errorf("sshcp: failed to dial: %s", err)
	}

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("sshcp: failed to create session: %s", err)
	}

	// Get a writer for the program's stdin.
	writer, err := session.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("sshcp: failed to create stdin pipe: %s", err)
	}

	// Run the 'cat' which actually copies.
	if err := session.Start("cat > " + remoteName); err != nil {
		return nil, fmt.Errorf("sshcp: failed to run: %s", err)
	}

	ret := &Conn{
		client:  client,
		session: session,
		writer:  writer,
	}
	return ret, nil
}

func (c *Conn) Close() error {
	var errs []error

	// Close the writer to force 'cat' to finish.
	if err := c.writer.Close(); err != nil {
		errs = append(errs, err)
	}

	// Wait for the process to finish.
	if err := c.session.Wait(); err != nil {
		errs = append(errs, err)
	}

	// Tell the session to close.
	if err := c.session.Close(); err != nil {
		errs = append(errs, err)
	}

	// Shut down the connection.
	c.client.Close()

	// Return the first error
	if len(errs) != 0 {
		return errs[0]
	}

	return nil
}

func (c *Conn) Write(p []byte) (n int, err error) {
	return c.writer.Write(p)
}

var _ io.WriteCloser = &Conn{}
