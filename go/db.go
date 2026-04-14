package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // Pure Go SQLite sürücüsü, CGO gerektirmez
)

// DB SQLite veritabanı işlemlerini yönetir
type DB struct {
	conn *sql.DB
}

// NewDB veritabanını açar veya oluşturur
func NewDB(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path) // modernc.org/sqlite için driver adı "sqlite"
	if err != nil {
		return nil, fmt.Errorf("DB açılamadı: %w", err)
	}

	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS fav (name TEXT, url TEXT);
		CREATE TABLE IF NOT EXISTS settings (key TEXT PRIMARY KEY, value TEXT);
	`)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("tablo oluşturulamadı: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close veritabanını kapatır
func (db *DB) Close() {
	if db.conn != nil {
		db.conn.Close()
	}
}

// AddFavorite favorilere kanal ekler; zaten varsa false döner
func (db *DB) AddFavorite(name, url string) (bool, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM fav WHERE name=? AND url=?", name, url).Scan(&count)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}
	_, err = db.conn.Exec("INSERT INTO fav (name, url) VALUES (?, ?)", name, url)
	return err == nil, err
}

// RemoveFavorite favorilerden kanal siler
func (db *DB) RemoveFavorite(name, url string) error {
	_, err := db.conn.Exec("DELETE FROM fav WHERE name=? AND url=?", name, url)
	return err
}

// GetFavorites tüm favori kanalları döner
func (db *DB) GetFavorites() ([]Channel, error) {
	rows, err := db.conn.Query("SELECT name, url FROM fav")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var channels []Channel
	for rows.Next() {
		var ch Channel
		if err := rows.Scan(&ch.Name, &ch.URL); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}

// GetAutoURL kayıtlı otomatik URL'yi döner
func (db *DB) GetAutoURL() (string, bool) {
	var v string
	err := db.conn.QueryRow("SELECT value FROM settings WHERE key='auto_url'").Scan(&v)
	if err != nil {
		return "", false
	}
	return v, true
}

// SetAutoURL otomatik URL'yi kaydeder
func (db *DB) SetAutoURL(url string) error {
	_, err := db.conn.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES ('auto_url', ?)", url)
	return err
}

// ClearAutoURL otomatik URL'yi siler
func (db *DB) ClearAutoURL() error {
	_, err := db.conn.Exec("DELETE FROM settings WHERE key='auto_url'")
	return err
}

// GetLanguage kayıtlı dil tercihini döner ("tr" veya "en"); yoksa "tr"
func (db *DB) GetLanguage() string {
	var v string
	err := db.conn.QueryRow("SELECT value FROM settings WHERE key='language'").Scan(&v)
	if err != nil || (v != "tr" && v != "en") {
		return "tr"
	}
	return v
}

// SetLanguage dil tercihini kaydeder
func (db *DB) SetLanguage(lang string) error {
	_, err := db.conn.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES ('language', ?)", lang)
	return err
}
