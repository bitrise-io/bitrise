package rde

import (
	"context"
	"fmt"
)

// SSHCredentials is the credential bundle a session exposes for SSH. The JSON
// tags define the stable shape used by `rde session ssh --output json`. Host,
// Port and User are the backend's ssh_address decomposed into discrete fields
// so a caller building its own connection (scp, an SSH_ASKPASS helper, a
// tunnel) never has to parse the address; Command is the ready-to-run ssh
// command line without the password.
type SSHCredentials struct {
	Address  string `json:"address" yaml:"address"`
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	Command  string `json:"command" yaml:"command"`
}

// GetSessionSSH fetches the session and returns its SSH credentials, erroring
// clearly when the session is not running or its SSH endpoint is not open yet.
func (s *Service) GetSessionSSH(ctx context.Context, workspaceID, sessionID string) (SSHCredentials, error) {
	if s.client == nil {
		return SSHCredentials{}, errClient()
	}
	sess, err := s.GetSession(ctx, workspaceID, sessionID)
	if err != nil {
		return SSHCredentials{}, err
	}
	return SSHCredentialsFromSession(sess)
}

// SSHCredentialsFromSession assembles the bundle from an already loaded
// Session, running the same reachability pre-flight `rde session exec` uses.
func SSHCredentialsFromSession(sess Session) (SSHCredentials, error) {
	target, err := sshTargetForSession(sess)
	if err != nil {
		return SSHCredentials{}, err
	}
	return SSHCredentials{
		Address:  sess.SSHAddress,
		Host:     target.Host,
		Port:     target.Port,
		User:     target.User,
		Password: target.Password,
		Command:  fmt.Sprintf("ssh -o StrictHostKeyChecking=no -p %d %s@%s", target.Port, target.User, target.Host),
	}, nil
}
