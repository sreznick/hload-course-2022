package utils

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
	"time"
)

const (
	CERT_PASSWORD        = 1
	CERT_PUBLIC_KEY_FILE = 2
	DEFAULT_TIMEOUT      = 3 // second
)

type SSH struct {
	Ip      string
	User    string
	Cert    string //password or key file path
	Port    int
	session *ssh.Session
	client  *ssh.Client
}

func (sshClient *SSH) readPublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func (sshClient *SSH) Connect(mode int) error {

	var ssh_config *ssh.ClientConfig
	var auth []ssh.AuthMethod
	if mode == CERT_PASSWORD {
		auth = []ssh.AuthMethod{ssh.Password(sshClient.Cert)}
	} else if mode == CERT_PUBLIC_KEY_FILE {
		auth = []ssh.AuthMethod{sshClient.readPublicKeyFile(sshClient.Cert)}
	} else {
		log.Println("Does not support mode: ", mode)
		return errors.New("Does not support mode")
	}

	ssh_config = &ssh.ClientConfig{
		User: sshClient.User,
		Auth: auth,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: time.Second * DEFAULT_TIMEOUT,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshClient.Ip, sshClient.Port), ssh_config)
	if err != nil {
		fmt.Println(err)
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		fmt.Println(err)
		client.Close()
		return err
	}

	sshClient.session = session
	sshClient.client = client

	return nil
}

func (sshClient *SSH) RunCmd(cmd string) {
	out, err := sshClient.session.CombinedOutput(cmd)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(out))
}

func (sshClient *SSH) Close() {
	sshClient.session.Close()
	sshClient.client.Close()
}
