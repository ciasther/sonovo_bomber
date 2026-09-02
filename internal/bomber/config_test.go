package bomber

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScoreStorePersistsHistory(t *testing.T) {
	root := t.TempDir()
	store := NewScoreStore(root)
	entries := []ScoreEntry{
		{ID: "a", EventID: "event", Nick: "ALA", Score: 100, SurvivalMS: 5000, CreatedAt: time.Now()},
		{ID: "b", EventID: "event", Nick: "JAN", Score: 200, SurvivalMS: 6000, CreatedAt: time.Now().Add(time.Second)},
	}
	for _, entry := range entries {
		if err := store.Append(entry); err != nil {
			t.Fatal(err)
		}
	}
	f, err := os.OpenFile(store.Path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString("uszkodzony wiersz\n")
	_ = f.Close()

	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 2 || loaded[0].ID != "a" || loaded[1].ID != "b" {
		t.Fatalf("wczytano niepoprawna historie: %+v", loaded)
	}
}

func TestScoreStoreChecksWritableDirectory(t *testing.T) {
	store := NewScoreStore(t.TempDir())
	if err := store.CheckWritable(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Dir(store.Path)); err != nil {
		t.Fatalf("katalog wynikow nie powstal: %v", err)
	}
}

func TestBrandingConfigAndFingerprint(t *testing.T) {
	root := t.TempDir()
	assets := filepath.Join(root, "assets")
	if err := os.MkdirAll(assets, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := DefaultConfig()
	cfg.EventID = "test-event"
	cfg.Partner.Primary = "#112233"
	write := func() {
		data, err := json.Marshal(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(assets, "branding.json"), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write()
	first, err := LoadBranding(root)
	if err != nil {
		t.Fatal(err)
	}
	if first.Primary != RGB(0x11, 0x22, 0x33) || first.Config.RoundSeconds != 75 {
		t.Fatalf("branding=%+v", first)
	}
	cfg.Partner.Primary = "#334455"
	write()
	second, err := LoadBranding(root)
	if err != nil {
		t.Fatal(err)
	}
	if first.Fingerprint == second.Fingerprint {
		t.Fatal("fingerprint nie zmienil sie po zmianie konfiguracji")
	}
}

func TestInvalidRoundAndPINUseSafeDefaults(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RoundSeconds = 12
	cfg.IdleReturnSeconds = 2
	cfg.AdminPIN = "1"
	if err := validateConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.RoundSeconds != 75 || cfg.IdleReturnSeconds != 15 || cfg.AdminPIN != "2468" {
		t.Fatalf("round=%d idle=%d pin=%s", cfg.RoundSeconds, cfg.IdleReturnSeconds, cfg.AdminPIN)
	}
}
