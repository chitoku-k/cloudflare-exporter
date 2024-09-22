package config

import (
	"fmt"
	"os"
	"strings"
)

type Environment struct {
	Port       string
	TLSCert    string
	TLSKey     string
	Cloudflare Cloudflare
}

type Cloudflare struct {
	// API Tokens require an API Token.
	APIToken string

	// API Keys require an API key and an email address.
	APIKey    string
	APIEmail  string
	AccountID string
}

func Get() (Environment, error) {
	var missing []string
	var env Environment

	for k, v := range map[string]*string{
		"TLS_CERT":      &env.TLSCert,
		"TLS_KEY":       &env.TLSKey,
		"CF_API_TOKEN":  &env.Cloudflare.APIToken,
		"CF_API_KEY":    &env.Cloudflare.APIKey,
		"CF_API_EMAIL":  &env.Cloudflare.APIEmail,
		"CF_ACCOUNT_ID": &env.Cloudflare.AccountID,
	} {
		*v = os.Getenv(k)
	}

	for k, v := range map[string]*string{
		"PORT": &env.Port,
	} {
		*v = os.Getenv(k)

		if *v == "" {
			missing = append(missing, k)
		}
	}

	if len(missing) > 0 {
		return env, fmt.Errorf("missing env(s): %s", strings.Join(missing, ", "))
	}

	if env.Cloudflare.APIToken == "" && env.Cloudflare.APIKey == "" && env.Cloudflare.APIEmail == "" {
		return env, fmt.Errorf("Cloudflare requires either CF_API_TOKEN, or CF_API_KEY and CF_API_EMAIL")
	}
	if env.Cloudflare.APIToken == "" && (env.Cloudflare.APIKey == "" || env.Cloudflare.APIEmail == "") {
		return env, fmt.Errorf("Cloudflare requires both CF_API_KEY and CF_API_EMAIL are set")
	}
	if env.Cloudflare.APIToken != "" && (env.Cloudflare.APIKey != "" || env.Cloudflare.APIEmail != "") {
		return env, fmt.Errorf("Cloudflare API token cannot be used with API key and email address")
	}

	return env, nil
}
