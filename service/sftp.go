package service

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
)

type foosftp struct {
	addr    string
	user    string
	keyFile string
	config  *ssh.ClientConfig
}

// NewSftp creates new sftp client
func NewSftp(addr string, user string, keyFile string) (*foosftp, error) {

	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("uunable to read private key: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sftp := &foosftp{
		addr:    addr,
		user:    user,
		keyFile: keyFile,
		config:  config,
	}

	fmt.Printf("Sftp created: %#v\n", sftp)

	return sftp, nil

}

// Copy copies local file to remote via sftp
// if ok returns md5 checksum of the sent file
func (s *foosftp) Copy(path string, filename string) (string, error) {

	// connect
	log.Println("Dialing sftp:", s.addr)
	client, err := ssh.Dial("tcp", s.addr, s.config)
	if err != nil {
		log.Println(err)
		return "", fmt.Errorf("unable to connect: %v", err)
	}
	defer client.Close()

	// create new SFTP client
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		log.Println(err)
		return "", fmt.Errorf("unable to create sftp client: %v", err)
	}
	defer sftpClient.Close()

	// open src file
	srcFile, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		return "", fmt.Errorf("unable to open file: %s, error: %v", filename, err)
	}
	defer srcFile.Close()

	// create destination file
	dstFileName := path + "/" + srcFile.Name()
	log.Println("Writing filename:", dstFileName, "addr:", s.addr, "user:", s.user, "ssh key:", s.keyFile)
	dstFile, err := sftpClient.Create(dstFileName)
	if err != nil {
		log.Println(err)
		return "", fmt.Errorf("unable to create remote file: %s, error: %v", dstFileName, err)
	}
	defer dstFile.Close()

	// copy source file to destination file
	written, err := io.Copy(dstFile, srcFile)
	if err != nil {
		log.Println(err)
		return "", fmt.Errorf("unable to copy source file to destination file: %s, error: %v", dstFileName, err)
	}
	log.Printf("%d bytes copied\n", written)

	// check it's there
	fi, err := sftpClient.Lstat(dstFile.Name())
	if err != nil {
		log.Println(err)
		return "", fmt.Errorf("unable to read destination file: %s, error: %v", dstFileName, err)
	}
	log.Println(fi)

	// calculate md5 sum of the file
	h, err := md5hash(filename)
	if err != nil {
		log.Println(err)
		return "", fmt.Errorf("unable to calculate md5 sum: %v", err)
	}

	return h, nil
}

// Lstat lists remote file
// if ok returns FileInfo
func (s *foosftp) Lstat(path string, fileName string) (os.FileInfo, error) {

	// connect
	log.Println("Dialing addr:", s.addr)
	client, err := ssh.Dial("tcp", s.addr, s.config)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to connect: %v", err)
	}
	defer client.Close()

	// create new SFTP client
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to create sftp client: %v", err)
	}
	defer sftpClient.Close()

	// create destination file
	dstFileName := path + "/" + fileName

	// check it's there
	fi, err := sftpClient.Lstat(dstFileName)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to read destination file: %s, error: %v", dstFileName, err)
	}
	log.Println(fi)

	return fi, nil
}
