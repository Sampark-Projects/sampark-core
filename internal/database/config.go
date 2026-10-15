package database

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

// DefaultPostgresSSLMode is the sslmode used when the caller leaves
// PostgresEnv.SSLMode empty. "disable" preserves the behavior the
// split-variable configuration path has always had; operators reaching a
// remote or managed PostgreSQL should set POSTGRES_SSLMODE=require or
// stricter. The connection-string path (POSTGRES_CONNECTION_STRING) carries
// its own sslmode in the URL and never passes through here.
const DefaultPostgresSSLMode = "disable"

// PostgresEnv holds the individual POSTGRES_* parameters that compose a
// connection string. Callers (internal/config) resolve defaults from env
// before invoking BuildPostgresConnString.
// last review: ser, 210426
type PostgresEnv struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	// SSLMode is the libpq sslmode. Empty means DefaultPostgresSSLMode;
	// internal/config validates the value before it gets here.
	SSLMode string
	// Schema, when set, becomes the connection's search_path so every table is
	// created in and read from that schema instead of the server default
	// ("public"). The schema must already exist. Empty keeps the default.
	Schema string
}

var postgresSchemaName = regexp.MustCompile(`^[a-z_][a-z0-9_]{0,62}$`)

// ValidatePostgresSchemaName accepts only lowercase unquoted identifiers, so the
// name means the same thing in CREATE SCHEMA and in search_path (which folds
// case), and cannot smuggle anything into the connection string.
func ValidatePostgresSchemaName(name string) error {
	if !postgresSchemaName.MatchString(name) || strings.HasPrefix(name, "pg_") {
		return fmt.Errorf("invalid schema name %q: use lowercase letters, digits and underscores, starting with a letter or underscore, and not starting with pg_", name)
	}
	return nil
}

// BuildPostgresConnString constructs a PostgreSQL connection string from
// already-resolved parameters.
//
// Built via net/url rather than fmt.Sprintf so credentials containing URL
// metacharacters (@ : / ? #) are percent-encoded instead of corrupting the
// DSN, and so IPv6 hosts get their brackets.
func BuildPostgresConnString(p PostgresEnv) string {
	sslMode := p.SSLMode
	if sslMode == "" {
		sslMode = DefaultPostgresSSLMode
	}

	query := url.Values{"sslmode": {sslMode}}
	if p.Schema != "" {
		query.Set("search_path", p.Schema)
	}

	u := &url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(p.User, p.Password),
		Host:     net.JoinHostPort(p.Host, p.Port),
		Path:     "/" + p.Database,
		RawQuery: query.Encode(),
	}
	return u.String()
}
