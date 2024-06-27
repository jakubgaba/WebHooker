package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/teris-io/shortid"
	gossh "golang.org/x/crypto/ssh"
)

func main() {
	sshPort := ":2222"

	respCh := make(chan string)

	go func() {
		time.Sleep(time.Second * 3)
		id, _ := shortid.Generate()
		respCh <- "http://webhooker.com/" + id
	}()

	handler := &SHHHandler{
		respCh: respCh,
	}

	server := &ssh.Server{
		Addr:    sshPort,
		Handler: handler.handleSSHSession,
		ServerConfigCallback: func(ctx ssh.Context) *gossh.ServerConfig {
			cfg := &gossh.ServerConfig{
				ServerVersion: "SSH-2.0-sendit",
			}
			cfg.Ciphers = []string{"chacha20-poly1305@openssh.com"}
			return cfg
		},
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			return true
		},
	}

	b, err := os.ReadFile("keys/privatekey")
	if err != nil {
		log.Fatal(err)
	}

	privateKey, err := gossh.ParsePrivateKey(b)
	if err != nil {
		log.Fatal(err)
	}

	server.AddHostKey(privateKey)

	log.Printf("Starting SSH server on port %s...", sshPort)
	log.Fatal(server.ListenAndServe())
}

type SHHHandler struct {
	respCh chan string
}

func (h *SHHHandler) handleSSHSession(session ssh.Session) {
	forwardURL := session.RawCommand()
	_ = forwardURL
	resp := <-h.respCh
	fmt.Println("Received: ", resp)
	session.Write([]byte(resp + "\n"))
}
