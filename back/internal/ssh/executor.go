package ssh

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHKeyProvider interface {
	GetByID(ctx context.Context, id string) (string, error)
}

type SSHAction struct {
	Host     string
	Port     int
	Username string
	SSHKeyID string
	Command  string
}

type Executor struct {
	keyProvider SSHKeyProvider
}

type ExecutionResult struct {
	Stdout string
	Stderr string
	Error  error
}

func NewExecutor(keyProvider SSHKeyProvider) *Executor {
	return &Executor{keyProvider: keyProvider}
}

func (e *Executor) Execute(ctx context.Context, action *SSHAction) (*ExecutionResult, error) {
	if action.Port <= 0 {
		action.Port = 22
	}

	privateKey, err := e.keyProvider.GetByID(ctx, action.SSHKeyID)
	if err != nil {
		return &ExecutionResult{Error: fmt.Errorf("get ssh key: %w", err)}, err
	}

	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return &ExecutionResult{Error: fmt.Errorf("parse private key: %w", err)}, err
	}

	config := &ssh.ClientConfig{
		User: action.Username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	addr := net.JoinHostPort(action.Host, fmt.Sprintf("%d", action.Port))
	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		return &ExecutionResult{Error: fmt.Errorf("dial %s: %w", addr, err)}, err
	}
	defer conn.Close()

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return &ExecutionResult{Error: fmt.Errorf("ssh connect: %w", err)}, err
	}
	defer sshConn.Close()

	client := ssh.NewClient(sshConn, chans, reqs)
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return &ExecutionResult{Error: fmt.Errorf("new session: %w", err)}, err
	}
	defer session.Close()

	output, err := session.CombinedOutput(action.Command)
	stdoutStderr := string(output)
	if err != nil {
		return &ExecutionResult{Stdout: stdoutStderr, Stderr: "", Error: err}, err
	}
	return &ExecutionResult{Stdout: stdoutStderr}, nil
}
