package bomber

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Config struct {
	EventID           string        `json:"event_id"`
	EventName         string        `json:"event_name"`
	RoundSeconds      int           `json:"round_seconds"`
	IdleReturnSeconds int           `json:"idle_return_seconds"`
	AdminPIN          string        `json:"admin_pin"`
	Partner           PartnerConfig `json:"partner"`
}

type PartnerConfig struct {
	Name            string `json:"name"`
	Headline        string `json:"headline"`
	SummaryTitle    string `json:"summary_title"`
	SummaryText     string `json:"summary_text"`
	Primary         string `json:"primary"`
	Secondary       string `json:"secondary"`
	Accent          string `json:"accent"`
	StartBackground string `json:"start_background"`
	PartnerLogo     string `json:"partner_logo"`
	SummaryAd       string `json:"summary_ad"`
	QR              string `json:"qr"`
	ProductItem     string `json:"product_item"`
}

type Branding struct {
	Config          Config
	Primary         Color
	Secondary       Color
	Accent          Color
	StartBackground *Sprite
	PartnerLogo     *Sprite
	SummaryAd       *Sprite
	QR              *Sprite
	ProductItem     *Sprite
	Fingerprint     string
	LoadedAt        time.Time
}

func DefaultConfig() Config {
	return Config{
		EventID:           "default-event",
		EventName:         "BOMBER RUSH",
		RoundSeconds:      75,
		IdleReturnSeconds: 15,
		AdminPIN:          "2468",
		Partner: PartnerConfig{
			Name:         "PARTNER",
			Headline:     "ZBIERZ PRODUKTY. ZDOBADZ NAJWYZSZY WYNIK.",
			SummaryTitle: "GRATULACJE!",
			SummaryText:  "ZESKANUJ KOD QR I SPRAWDZ SZCZEGOLY.",
			Primary:      "#63F3FF",
			Secondary:    "#FF4FD8",
			Accent:       "#FFB547",
		},
	}
}

func LoadBranding(root string) (*Branding, error) {
	assetsDir := filepath.Join(root, "assets")
	cfg, fingerprint, err := readBrandingConfig(root)
	if err != nil {
		return nil, err
	}
	b := &Branding{
		Config:      cfg,
		Primary:     ParseHexColor(cfg.Partner.Primary, RGB(99, 243, 255)),
		Secondary:   ParseHexColor(cfg.Partner.Secondary, RGB(255, 79, 216)),
		Accent:      ParseHexColor(cfg.Partner.Accent, RGB(255, 181, 71)),
		Fingerprint: fingerprint,
		LoadedAt:    time.Now(),
	}
	var loadErrs []string
	load := func(name string, dst **Sprite) {
		if strings.TrimSpace(name) == "" {
			return
		}
		path := filepath.Join(assetsDir, filepath.Base(name))
		s, imageErr := LoadSprite(path)
		if imageErr != nil {
			loadErrs = append(loadErrs, fmt.Sprintf("%s: %v", filepath.Base(path), imageErr))
			return
		}
		*dst = s
	}
	load(cfg.Partner.StartBackground, &b.StartBackground)
	load(cfg.Partner.PartnerLogo, &b.PartnerLogo)
	load(cfg.Partner.SummaryAd, &b.SummaryAd)
	load(cfg.Partner.QR, &b.QR)
	load(cfg.Partner.ProductItem, &b.ProductItem)
	if len(loadErrs) > 0 {
		return b, errors.New(strings.Join(loadErrs, "; "))
	}
	return b, nil
}

func BrandingFingerprint(root string) (string, error) {
	_, fingerprint, err := readBrandingConfig(root)
	return fingerprint, err
}

func readBrandingConfig(root string) (Config, string, error) {
	assetsDir := filepath.Join(root, "assets")
	configPath := filepath.Join(assetsDir, "branding.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, "", fmt.Errorf("odczyt branding.json: %w", err)
	}
	cfg := DefaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, "", fmt.Errorf("parsowanie branding.json: %w", err)
	}
	if err := validateConfig(&cfg); err != nil {
		return Config{}, "", err
	}
	var fp strings.Builder
	fp.Write(data)
	for _, name := range []string{
		cfg.Partner.StartBackground,
		cfg.Partner.PartnerLogo,
		cfg.Partner.SummaryAd,
		cfg.Partner.QR,
		cfg.Partner.ProductItem,
	} {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		path := filepath.Join(assetsDir, filepath.Base(name))
		st, statErr := os.Stat(path)
		if statErr != nil {
			fmt.Fprintf(&fp, "|%s:missing", filepath.Base(path))
			continue
		}
		fmt.Fprintf(&fp, "|%s:%d:%d", filepath.Base(path), st.Size(), st.ModTime().UnixNano())
	}
	return cfg, HashString(fp.String()), nil
}

func validateConfig(cfg *Config) error {
	cfg.EventID = strings.TrimSpace(cfg.EventID)
	cfg.EventName = strings.TrimSpace(cfg.EventName)
	cfg.Partner.Name = strings.TrimSpace(cfg.Partner.Name)
	if cfg.EventID == "" {
		return errors.New("event_id nie moze byc pusty")
	}
	if cfg.EventName == "" {
		cfg.EventName = "BOMBER RUSH"
	}
	if cfg.RoundSeconds != 60 && cfg.RoundSeconds != 75 && cfg.RoundSeconds != 90 {
		cfg.RoundSeconds = 75
	}
	if cfg.IdleReturnSeconds < 8 || cfg.IdleReturnSeconds > 60 {
		cfg.IdleReturnSeconds = 15
	}
	if len(cfg.AdminPIN) < 4 || len(cfg.AdminPIN) > 12 {
		cfg.AdminPIN = "2468"
	}
	if cfg.Partner.Name == "" {
		cfg.Partner.Name = "PARTNER"
	}
	return nil
}

func LoadSprite(path string) (*Sprite, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	return SpriteFromImage(img), nil
}

type ScoreEntry struct {
	ID           string    `json:"id"`
	EventID      string    `json:"event_id"`
	Nick         string    `json:"nick"`
	Score        int       `json:"score"`
	SurvivalMS   int64     `json:"survival_ms"`
	BestCombo    int       `json:"best_combo"`
	Waves        int       `json:"waves"`
	CreatedAt    time.Time `json:"created_at"`
	CompletedRun bool      `json:"completed_run"`
}

type ScoreStore struct {
	Path string
}

func NewScoreStore(root string) *ScoreStore {
	return &ScoreStore{Path: filepath.Join(root, "data", "scores.jsonl")}
}

func (s *ScoreStore) CheckWritable() error {
	dir := filepath.Dir(s.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("brak dostepu do katalogu wynikow: %w", err)
	}
	f, err := os.CreateTemp(dir, ".write-check-*")
	if err != nil {
		return fmt.Errorf("nie mozna zapisywac wynikow: %w", err)
	}
	name := f.Name()
	closeErr := f.Close()
	removeErr := os.Remove(name)
	if closeErr != nil {
		return fmt.Errorf("nie mozna zamknac testu zapisu: %w", closeErr)
	}
	if removeErr != nil {
		return fmt.Errorf("nie mozna usunac testu zapisu: %w", removeErr)
	}
	return nil
}

func (s *ScoreStore) Append(entry ScoreEntry) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(s.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	return enc.Encode(entry)
}

func (s *ScoreStore) Load() ([]ScoreEntry, error) {
	f, err := os.Open(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var entries []ScoreEntry
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		var e ScoreEntry
		if json.Unmarshal(scanner.Bytes(), &e) == nil && e.ID != "" {
			entries = append(entries, e)
		}
	}
	return entries, scanner.Err()
}

func RankEntries(entries []ScoreEntry, eventID string, day *time.Time) []ScoreEntry {
	filtered := make([]ScoreEntry, 0, len(entries))
	for _, e := range entries {
		if eventID != "" && e.EventID != eventID {
			continue
		}
		if day != nil {
			y, m, d := e.CreatedAt.In(day.Location()).Date()
			dy, dm, dd := day.Date()
			if y != dy || m != dm || d != dd {
				continue
			}
		}
		filtered = append(filtered, e)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].Score != filtered[j].Score {
			return filtered[i].Score > filtered[j].Score
		}
		if filtered[i].SurvivalMS != filtered[j].SurvivalMS {
			return filtered[i].SurvivalMS > filtered[j].SurvivalMS
		}
		return filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
	})
	return filtered
}

func PositionOf(entries []ScoreEntry, id string) int {
	for i, e := range entries {
		if e.ID == id {
			return i + 1
		}
	}
	return 0
}

func savePNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriterSize(f, 1<<20)
	if err := png.Encode(w, img); err != nil {
		return err
	}
	return w.Flush()
}

func SavePNG(path string, img image.Image) error { return savePNG(path, img) }
