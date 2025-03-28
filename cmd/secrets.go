package cmd

import (
	"fmt"
	"strings"
	"unicode"
)

// returns secrets in order of original references
// e.g:
//
//		secrets, err := a.resolveSecrets(
//			[]string{
//				"/Postgres/Replicator/username",
//				"/Postgres/Replicator/password"
//			},
//	)
//
// secrets[0] is replicator username and so on.
func (a *AppCtx) resolveSecrets(keys []string) ([]string, error) {
	var prefixedKeys = make([]string, 0, len(keys))
	for _, key := range keys {
		prefixedKeys = append(prefixedKeys, a.Vault.Prefix+strings.TrimSpace(key))
	}
	secretsResponse, err := a.Vault.Client.Secrets().ResolveAll(a.Context, prefixedKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve secrets: %w", err)
	}
	var secrets = make([]string, len(prefixedKeys))
	for _, key := range prefixedKeys {
		secret := secretsResponse.IndividualResponses[key]
		if secret.Error != nil {
			return nil, fmt.Errorf("error: %s", secret.Error.Type)
		}
		cleanedSecret := strings.TrimSpace(secret.Content.Secret)
		cleanedSecret = strings.TrimFunc(cleanedSecret, func(r rune) bool { return unicode.IsControl(r) })
		secrets = append(secrets, cleanedSecret)
	}
	return secrets, nil
}
