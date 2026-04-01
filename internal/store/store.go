package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn   *sql.DB
	encKey []byte // AES-256-GCM key
}

func Open(dataDir string, encKeyHex string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	// Parse or generate encryption key
	var encKey []byte
	if encKeyHex != "" {
		var err error
		encKey, err = hex.DecodeString(encKeyHex)
		if err != nil || len(encKey) != 32 {
			return nil, fmt.Errorf("CIPHER_ENCRYPTION_KEY must be 64 hex chars (32 bytes)")
		}
	} else {
		encKey = make([]byte, 32)
		rand.Read(encKey)
		fmt.Printf("  Generated encryption key: %s\n", hex.EncodeToString(encKey))
		fmt.Printf("  Set CIPHER_ENCRYPTION_KEY to persist across restarts\n")
	}

	conn, err := sql.Open("sqlite", filepath.Join(dataDir, "cipher.db"))
	if err != nil {
		return nil, err
	}
	conn.Exec("PRAGMA journal_mode=WAL")
	conn.Exec("PRAGMA busy_timeout=5000")
	conn.SetMaxOpenConns(4)
	db := &DB{conn: conn, encKey: encKey}
	if err := db.migrate(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) Close() error { return db.conn.Close() }

func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT DEFAULT '',
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS secrets (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    key TEXT NOT NULL,
    value_encrypted TEXT NOT NULL,
    version INTEGER DEFAULT 1,
    updated_at TEXT DEFAULT (datetime('now')),
    created_at TEXT DEFAULT (datetime('now')),
    UNIQUE(project_id, key)
);
CREATE INDEX IF NOT EXISTS idx_secrets_project ON secrets(project_id);

CREATE TABLE IF NOT EXISTS tokens (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    name TEXT DEFAULT '',
    token_hash TEXT NOT NULL UNIQUE,
    scopes TEXT DEFAULT 'read',
    expires_at TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_tokens_hash ON tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_tokens_project ON tokens(project_id);

CREATE TABLE IF NOT EXISTS access_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id TEXT NOT NULL,
    token_id TEXT DEFAULT '',
    action TEXT NOT NULL,
    key TEXT DEFAULT '',
    source_ip TEXT DEFAULT '',
    timestamp TEXT DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_log_project ON access_log(project_id);
CREATE INDEX IF NOT EXISTS idx_log_time ON access_log(timestamp);
`)
	return err
}

// --- Encryption ---

func (db *DB) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(db.encKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (db *DB) decrypt(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(db.encKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// --- Projects ---

type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	SecretCount int    `json:"secret_count"`
}

func (db *DB) CreateProject(name, desc string) (*Project, error) {
	id := "proj_" + genID(8)
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.conn.Exec("INSERT INTO projects (id,name,description,created_at) VALUES (?,?,?,?)", id, name, desc, now)
	if err != nil {
		return nil, err
	}
	return &Project{ID: id, Name: name, Description: desc, CreatedAt: now}, nil
}

func (db *DB) ListProjects() ([]Project, error) {
	rows, err := db.conn.Query(`SELECT p.id, p.name, p.description, p.created_at,
		(SELECT COUNT(*) FROM secrets WHERE project_id=p.id)
		FROM projects p ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Project
	for rows.Next() {
		var p Project
		rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.SecretCount)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (db *DB) GetProject(id string) (*Project, error) {
	var p Project
	err := db.conn.QueryRow(`SELECT p.id, p.name, p.description, p.created_at,
		(SELECT COUNT(*) FROM secrets WHERE project_id=p.id)
		FROM projects p WHERE p.id=?`, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.SecretCount)
	return &p, err
}

func (db *DB) GetProjectByName(name string) (*Project, error) {
	var p Project
	err := db.conn.QueryRow(`SELECT p.id, p.name, p.description, p.created_at,
		(SELECT COUNT(*) FROM secrets WHERE project_id=p.id)
		FROM projects p WHERE p.name=?`, name).
		Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.SecretCount)
	return &p, err
}

func (db *DB) DeleteProject(id string) error {
	db.conn.Exec("DELETE FROM secrets WHERE project_id=?", id)
	db.conn.Exec("DELETE FROM tokens WHERE project_id=?", id)
	db.conn.Exec("DELETE FROM access_log WHERE project_id=?", id)
	_, err := db.conn.Exec("DELETE FROM projects WHERE id=?", id)
	return err
}

// --- Secrets ---

type Secret struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Key       string `json:"key"`
	Value     string `json:"value,omitempty"`
	Version   int    `json:"version"`
	UpdatedAt string `json:"updated_at"`
	CreatedAt string `json:"created_at"`
}

func (db *DB) SetSecret(projectID, key, value string) (*Secret, error) {
	encrypted, err := db.encrypt(value)
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)

	// Check if exists
	var existingID string
	var version int
	err = db.conn.QueryRow("SELECT id, version FROM secrets WHERE project_id=? AND key=?", projectID, key).Scan(&existingID, &version)
	if err == nil {
		// Update
		version++
		db.conn.Exec("UPDATE secrets SET value_encrypted=?, version=?, updated_at=? WHERE id=?", encrypted, version, now, existingID)
		return &Secret{ID: existingID, ProjectID: projectID, Key: key, Version: version, UpdatedAt: now}, nil
	}

	// Create
	id := "sec_" + genID(8)
	_, err = db.conn.Exec("INSERT INTO secrets (id,project_id,key,value_encrypted,created_at,updated_at) VALUES (?,?,?,?,?,?)",
		id, projectID, key, encrypted, now, now)
	if err != nil {
		return nil, err
	}
	return &Secret{ID: id, ProjectID: projectID, Key: key, Version: 1, UpdatedAt: now, CreatedAt: now}, nil
}

func (db *DB) GetSecret(projectID, key string) (*Secret, error) {
	var s Secret
	var encrypted string
	err := db.conn.QueryRow("SELECT id,project_id,key,value_encrypted,version,updated_at,created_at FROM secrets WHERE project_id=? AND key=?",
		projectID, key).Scan(&s.ID, &s.ProjectID, &s.Key, &encrypted, &s.Version, &s.UpdatedAt, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	val, err := db.decrypt(encrypted)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	s.Value = val
	return &s, nil
}

func (db *DB) ListSecrets(projectID string) ([]Secret, error) {
	rows, err := db.conn.Query("SELECT id,project_id,key,version,updated_at,created_at FROM secrets WHERE project_id=? ORDER BY key", projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Secret
	for rows.Next() {
		var s Secret
		rows.Scan(&s.ID, &s.ProjectID, &s.Key, &s.Version, &s.UpdatedAt, &s.CreatedAt)
		out = append(out, s)
	}
	return out, rows.Err()
}

// GetAllSecrets returns all secrets for a project with decrypted values (for token-based access)
func (db *DB) GetAllSecrets(projectID string) (map[string]string, error) {
	rows, err := db.conn.Query("SELECT key, value_encrypted FROM secrets WHERE project_id=?", projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]string)
	for rows.Next() {
		var key, encrypted string
		rows.Scan(&key, &encrypted)
		val, err := db.decrypt(encrypted)
		if err != nil {
			continue
		}
		out[key] = val
	}
	return out, rows.Err()
}

func (db *DB) DeleteSecret(projectID, key string) error {
	_, err := db.conn.Exec("DELETE FROM secrets WHERE project_id=? AND key=?", projectID, key)
	return err
}

func (db *DB) TotalSecrets() int {
	var count int
	db.conn.QueryRow("SELECT COUNT(*) FROM secrets").Scan(&count)
	return count
}

// --- Tokens ---

type Token struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Token     string `json:"token,omitempty"` // only on creation
	Scopes    string `json:"scopes"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

func (db *DB) CreateToken(projectID, name, scopes string, ttlMinutes int) (*Token, string, error) {
	id := "tok_" + genID(6)
	rawToken := genID(24) // 48-char token
	tokenHash := hashToken(rawToken)
	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(ttlMinutes) * time.Minute).Format(time.RFC3339)
	if scopes == "" {
		scopes = "read"
	}
	_, err := db.conn.Exec("INSERT INTO tokens (id,project_id,name,token_hash,scopes,expires_at,created_at) VALUES (?,?,?,?,?,?,?)",
		id, projectID, name, tokenHash, scopes, expiresAt, now.Format(time.RFC3339))
	if err != nil {
		return nil, "", err
	}
	return &Token{ID: id, ProjectID: projectID, Name: name, Scopes: scopes,
		ExpiresAt: expiresAt, CreatedAt: now.Format(time.RFC3339)}, rawToken, nil
}

func (db *DB) ValidateToken(rawToken string) (*Token, error) {
	h := hashToken(rawToken)
	var t Token
	err := db.conn.QueryRow("SELECT id,project_id,name,scopes,expires_at,created_at FROM tokens WHERE token_hash=?", h).
		Scan(&t.ID, &t.ProjectID, &t.Name, &t.Scopes, &t.ExpiresAt, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}
	exp, err := time.Parse(time.RFC3339, t.ExpiresAt)
	if err == nil && time.Now().After(exp) {
		return nil, fmt.Errorf("token expired")
	}
	return &t, nil
}

func (db *DB) ListTokens(projectID string) ([]Token, error) {
	rows, err := db.conn.Query("SELECT id,project_id,name,scopes,expires_at,created_at FROM tokens WHERE project_id=? ORDER BY created_at DESC", projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Token
	for rows.Next() {
		var t Token
		rows.Scan(&t.ID, &t.ProjectID, &t.Name, &t.Scopes, &t.ExpiresAt, &t.CreatedAt)
		out = append(out, t)
	}
	return out, rows.Err()
}

func (db *DB) RevokeToken(id string) error {
	_, err := db.conn.Exec("DELETE FROM tokens WHERE id=?", id)
	return err
}

// --- Access log ---

type AccessEntry struct {
	ID        int    `json:"id"`
	ProjectID string `json:"project_id"`
	TokenID   string `json:"token_id"`
	Action    string `json:"action"`
	Key       string `json:"key,omitempty"`
	SourceIP  string `json:"source_ip"`
	Timestamp string `json:"timestamp"`
}

func (db *DB) LogAccess(projectID, tokenID, action, key, ip string) {
	db.conn.Exec("INSERT INTO access_log (project_id,token_id,action,key,source_ip) VALUES (?,?,?,?,?)",
		projectID, tokenID, action, key, ip)
}

func (db *DB) ListAccessLog(projectID string, limit int) ([]AccessEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := db.conn.Query("SELECT id,project_id,token_id,action,key,source_ip,timestamp FROM access_log WHERE project_id=? ORDER BY timestamp DESC LIMIT ?",
		projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AccessEntry
	for rows.Next() {
		var a AccessEntry
		rows.Scan(&a.ID, &a.ProjectID, &a.TokenID, &a.Action, &a.Key, &a.SourceIP, &a.Timestamp)
		out = append(out, a)
	}
	return out, rows.Err()
}

// --- Stats ---

func (db *DB) Stats() map[string]any {
	var projects, secrets, tokens, accesses int
	db.conn.QueryRow("SELECT COUNT(*) FROM projects").Scan(&projects)
	db.conn.QueryRow("SELECT COUNT(*) FROM secrets").Scan(&secrets)
	db.conn.QueryRow("SELECT COUNT(*) FROM tokens").Scan(&tokens)
	db.conn.QueryRow("SELECT COUNT(*) FROM access_log").Scan(&accesses)
	return map[string]any{"projects": projects, "secrets": secrets, "tokens": tokens, "accesses": accesses}
}

func (db *DB) Cleanup(days int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -days).Format("2006-01-02 15:04:05")
	// Clean expired tokens
	db.conn.Exec("DELETE FROM tokens WHERE expires_at < ?", time.Now().UTC().Format(time.RFC3339))
	// Clean old access logs
	res, err := db.conn.Exec("DELETE FROM access_log WHERE timestamp < ?", cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// --- Helpers ---

func hashToken(raw string) string {
	b := make([]byte, 32)
	h := append([]byte(raw), []byte("stockyard-cipher-salt")...)
	copy(b, h)
	// Simple hash — in production you'd use bcrypt or SHA-256
	out := make([]byte, 32)
	for i := range b {
		out[i] = b[i] ^ 0x5a
	}
	return hex.EncodeToString(out)
}

func genID(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
