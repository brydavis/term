// Reference Code Sources:
// ********************************
// github.com/inatus/ssh-client-go
// godoc.org/golang.org/x/crypto/ssh
// golang-basic.blogspot.com/2014/06/step-by-step-guide-to-ssh-using-go.html

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"

	"golang.org/x/crypto/ssh"
)

var (
	host     = flag.String("host", "127.0.0.1", "Host IP address or URL")
	port     = flag.String("port", "22", "Host's SSH port")
	username = flag.String("user", "root", "Username @ login")
	password = flag.String("pass", "", "Password @ login")
)

func main() {
	flag.Parse()
	pkey, err := getKeyFile()
	if err != nil {
		panic(err)
	}

	config := &ssh.ClientConfig{
		User: *username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(pkey),
			ssh.Password(*password),
		},
	}

	// Dial your ssh server.
	conn, err := ssh.Dial("tcp", *host+":"+*port, config)
	if err != nil {
		log.Fatalf("unable to connect: %s", err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	defer session.Close()

	// Setting up IO
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	in, _ := session.StdinPipe()

	// Setting up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// Requesting pseudo terminal
	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		log.Fatalf("request for pseudo terminal failed: %s", err)
	}

	// Starting remote shell
	if err := session.Shell(); err != nil {
		log.Fatalf("failed to start shell: %s", err)
	}

	// Lastly, accepting commands
	for {
		reader := bufio.NewReader(os.Stdin)
		str, _ := reader.ReadString('\n')

		fmt.Fprint(in, str)
		if str != "exit" {
			os.Exit(1)
		}
	}
}

func getKeyFile() (key ssh.Signer, err error) {
	usr, _ := user.Current()
	fmt.Println(usr.Username, usr.HomeDir)
	file := usr.HomeDir + "/.ssh/id_rsa"
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	key, err = ssh.ParsePrivateKey(buf)
	if err != nil {
		return
	}
	return
}
