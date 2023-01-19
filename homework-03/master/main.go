package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/ssh"
	"log"
	master "main/master/src"
	"os"
)

func main() {
	key, err := os.ReadFile("//users//sergeidiagilev//.ssh//id_ed25519")
	if err != nil {
		log.Fatalf("os.ReadFile: %v", err)
	}

	var signer ssh.Signer
	signer, err = ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("ssh.ParsePrivateKey: %v", err)
	}

	server := &master.SSH{
		Ip:     "51.250.106.140",
		User:   "mdiagilev",
		Port:   22,
		Cert:   "postgres",
		Signer: signer,
	}

	err = server.Connect(master.CERT_PUBLIC_KEY_FILE)
	if err != nil {
		log.Fatalf("server.Connect: %v", err)
	}
	defer server.Close()

	sql.Register("postgres+ssh", &master.ViaSSHDialer{Client: server.Client})
	conn, err := sqlx.Open("postgres+ssh", "user=postgres dbname=mdiagilev sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to open: %w", err)
	}

	if err = conn.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	r := master.SetupRouter()
	r.PUT("/create", func(ctx *gin.Context) {
		if err = master.Put(ctx, conn); err != nil {
			log.Fatalf("master.Put: %v", err)
		}
	})

	go master.MasterReadFromReplicaIncrClick(conn, "mdiagilev-events-links")

	r.Run("0.0.0.0:8080")
}
