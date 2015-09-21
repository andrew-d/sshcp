package main

import (
	"fmt"
	"io"
	"os"

	"github.com/andrew-d/sshcp"
	"golang.org/x/crypto/ssh"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Fprintf(os.Stderr, "Usage: %s <host> <username> <password> <remote name>\n",
			os.Args[0])
		os.Exit(1)
	}

	conn, err := sshcp.NewConn(
		os.Args[1],
		os.Args[2],
		os.Args[4],
		ssh.Password(os.Args[3]),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open connection: %s", err)
		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error copying data: %s", err)
		return
	}

	fmt.Fprintln(os.Stderr, "done")
}
