package analytics

import (
	"context"
	"fmt"
	"strings"

	"github.com/plexusone/omnivault"
	envprovider "github.com/plexusone/omnivault/providers/env"
	fileprovider "github.com/plexusone/omnivault/providers/file"
	"github.com/plexusone/omnivault/vault"
)

// NewDefaultResolver builds the server's OmniVault resolver with the schemes
// supported out of the box:
//
//   - env://VAR_NAME reads an environment variable.
//   - file:///absolute/path reads a secret file (rooted at the filesystem
//     root, so references carry absolute paths).
//
// Additional schemes (keyring://, aws-sm://, sql://, ...) can be registered
// on the returned resolver without further changes to callers.
func NewDefaultResolver() (*omnivault.Resolver, error) {
	resolver := omnivault.NewResolver()
	resolver.Register("env", envprovider.New())
	fileVault, err := fileprovider.New(fileprovider.Config{Directory: "/"})
	if err != nil {
		return nil, fmt.Errorf("creating file secret provider: %w", err)
	}
	resolver.Register("file", fileVault)
	return resolver, nil
}

// ResolveDSN resolves a dsnRef secret reference to a DSN. The returned value
// must stay in memory only: callers dial with it and discard it, and must not
// log, persist, or serialize it.
//
// A reference whose scheme has no registered vault provider is rejected,
// which also catches URL-style raw DSNs such as mysql://user:pass@host/db.
func ResolveDSN(ctx context.Context, resolver *omnivault.Resolver, dsnRef string) (string, error) {
	if resolver == nil {
		return "", fmt.Errorf("secret resolver not configured")
	}
	if !IsSecretRefStrict(dsnRef) {
		return "", fmt.Errorf("dsnRef must be a secret reference such as env://VAR_NAME")
	}
	scheme := vault.SecretRef(dsnRef).Scheme()
	if _, ok := resolver.Get(scheme); !ok {
		return "", fmt.Errorf("no secret provider registered for scheme %q (available: %s)",
			scheme, strings.Join(resolver.Schemes(), ", "))
	}
	dsn, err := resolver.Resolve(ctx, dsnRef)
	if err != nil {
		// The error from a vault provider names the reference path, never the
		// secret value, so it is safe to wrap.
		return "", fmt.Errorf("resolving %s: %w", dsnRef, err)
	}
	if strings.TrimSpace(dsn) == "" {
		return "", fmt.Errorf("secret reference %s resolved to an empty value", dsnRef)
	}
	return dsn, nil
}
