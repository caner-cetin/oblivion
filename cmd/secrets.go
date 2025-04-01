package cmd

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/1password/onepassword-sdk-go"
)

// todo: comment
func (a *AppCtx) resolveSecrets(keys []string) (map[string]string, map[string]onepassword.Response[onepassword.ResolvedReference, onepassword.ResolveReferenceError], error) {
	var prefixedKeys = make([]string, 0, len(keys))
	for _, key := range keys {
		prefixedKeys = append(prefixedKeys, a.Vault.Prefix+strings.TrimSpace(key))
	}
	secretsResponse, err := a.Vault.Client.Secrets().ResolveAll(a.Context, prefixedKeys)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve secrets: %w", err)
	}
	for _, v := range secretsResponse.IndividualResponses {
		secret := v.Content.Secret
		secret = strings.TrimSpace(secret)
		secret = strings.TrimFunc(secret, func(r rune) bool { return unicode.IsControl(r) })
		v.Content.Secret = secret
	}
	var pfKeyMap = make(map[string]string)
	for i, k := range prefixedKeys {
		pfKeyMap[keys[i]] = k
	}
	return pfKeyMap, secretsResponse.IndividualResponses, nil
}
