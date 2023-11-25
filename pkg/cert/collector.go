package cert

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"

	gmlog "github.com/onuryurdupak/gomod/v2/log"
)

type certificateContextManager interface {
	ExtractSessionID(ctx context.Context) string
}

type collector struct {
	certificateContextManager certificateContextManager
}

func NewCertificateCollector(certificateContextManager certificateContextManager) *collector {
	return &collector{
		certificateContextManager: certificateContextManager,
	}
}

func (c *collector) AppendCertsFromDir(ctx context.Context, certPool *x509.CertPool, dir string) error {
	return c.appendCertsFromDirRecursive(ctx, certPool, dir)
}

func (c *collector) appendCertsFromDirRecursive(ctx context.Context, certPool *x509.CertPool, dir string) error {
	sessionID := c.certificateContextManager.ExtractSessionID(ctx)
	logger := gmlog.NewLogger("appendCertsFromDirRecursive", sessionID)

	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed reading directory: %w", err)
	}

	for _, file := range files {
		filePath := filepath.Join(dir, file.Name())

		if file.IsDir() {
			if err := c.appendCertsFromDirRecursive(ctx, certPool, filePath); err != nil {
				return err
			}
		} else {
			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed reading file %s: %w", file.Name(), err)
			}

			if ok := certPool.AppendCertsFromPEM(content); !ok {
				return fmt.Errorf("failed to parse client ca cert from file: %s", filePath)
			}
			logger.Infof("Successfully loaded certificate: %s", filePath)
		}
	}
	return nil
}
