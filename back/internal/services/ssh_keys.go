package services

import (
	"context"
	"fmt"
	"time"

	db "github.com/build-assistant/back/db/gen"
	"github.com/build-assistant/back/internal/api/handlers/ssh_keys"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

type SSHKeyService struct {
	queries *db.Queries
}

func NewSSHKeyService(queries *db.Queries) *SSHKeyService {
	return &SSHKeyService{queries: queries}
}

func (s *SSHKeyService) Create(ctx context.Context, name string, privateKey string) (*ssh_keys.SSHKey, error) {
	fingerprint, err := computeFingerprint(privateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	key, err := s.queries.CreateSSHKey(ctx, &db.CreateSSHKeyParams{
		Name:        name,
		PrivateKey:  privateKey,
		Fingerprint: fingerprint,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ssh key: %w", err)
	}

	return &ssh_keys.SSHKey{
		ID:          key.ID.String(),
		Name:        key.Name,
		Fingerprint: key.Fingerprint,
		CreatedAt:   key.CreatedAt.Time.Format(time.RFC3339),
	}, nil
}

func (s *SSHKeyService) List(ctx context.Context) ([]ssh_keys.SSHKey, error) {
	rows, err := s.queries.ListSSHKeys(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ssh_keys.SSHKey, len(rows))
	for i, r := range rows {
		result[i] = ssh_keys.SSHKey{
			ID:          r.ID.String(),
			Name:        r.Name,
			Fingerprint: r.Fingerprint,
			CreatedAt:   r.CreatedAt.Time.Format(time.RFC3339),
		}
	}

	return result, nil
}

func (s *SSHKeyService) Delete(ctx context.Context, id string) error {
	keyID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid ssh key ID: %w", err)
	}

	return s.queries.DeleteSSHKey(ctx, keyID)
}

func (s *SSHKeyService) GetByID(ctx context.Context, id string) (string, error) {
	keyID, err := uuid.Parse(id)
	if err != nil {
		return "", fmt.Errorf("invalid ssh key ID: %w", err)
	}

	key, err := s.queries.GetSSHKeyByID(ctx, keyID)
	if err != nil {
		return "", fmt.Errorf("ssh key not found: %w", err)
	}

	return key.PrivateKey, nil
}

func computeFingerprint(privateKeyPEM string) (string, error) {
	signer, err := ssh.ParsePrivateKey([]byte(privateKeyPEM))
	if err != nil {
		return "", err
	}
	return ssh.FingerprintSHA256(signer.PublicKey()), nil
}
