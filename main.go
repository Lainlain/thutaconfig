package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// loadEnvFile reads key=value pairs from a .env file and sets them as
// environment variables (only if not already set by the OS).
func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // no .env file — that's fine
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if os.Getenv(key) == "" { // don't override real env vars
			os.Setenv(key, val)
		}
	}
}

var db *sql.DB

// AppVersion represents app version and update information
type AppVersion struct {
	LatestVersion      string   `json:"lv"`
	LatestVersionCode  int      `json:"lvc"`
	MinimumVersionCode int      `json:"mvc"`
	ForceUpdate        bool     `json:"fu"`
	UpdateTitle        string   `json:"ut"`
	UpdateMessage      string   `json:"um"`
	WhatsNew           []string `json:"wn"`
	ReleaseDate        string   `json:"rd"`
	DownloadURL        string   `json:"du"`      // Universal / fallback APK link
	DownloadURLv8a     string   `json:"du_v8a"`  // arm64-v8a APK (modern phones)
	DownloadURLv7a     string   `json:"du_v7a"`  // armeabi-v7a APK (older phones)
	UpdateButtonText   string   `json:"ubt"`     // e.g. "ယခုဒေါင်းလုပ်"
	LaterButtonText    string   `json:"lbt"`     // e.g. "နောက်မှ" (empty = hide)
}

// InAppMessage represents an in-app announcement shown as a bottom sheet
type InAppMessage struct {
	ID          string `json:"i"`
	Type        string `json:"tp"` // info | promo | warning
	Title       string `json:"tt"`
	Message     string `json:"mg"`
	ImageURL    string `json:"iu,omitempty"`
	ActionText  string `json:"at,omitempty"`
	ActionURL   string `json:"au,omitempty"`
	Priority    int    `json:"pr"`
	StartDate   string `json:"sd"`
	EndDate     string `json:"ed"`
	ShowOnce    bool   `json:"so"`
	Dismissible bool   `json:"dm"`
}

// Banner represents a home-screen ad/promo banner slide
type Banner struct {
	ID       string `json:"id"`
	ImageURL string `json:"image_url"`
	LinkURL  string `json:"link_url"`
	Title    string `json:"title"`
	IsActive bool   `json:"is_active"`
	SortOrder int   `json:"sort_order"`
}

// AppConfig is the complete configuration response
type AppConfig struct {
	AppVersion    AppVersion     `json:"av"`
	InAppMessages []InAppMessage `json:"ms"`
}

// isDuplicateColumnError returns true when SQLite rejects an ALTER TABLE
// because the column already exists.
func isDuplicateColumnError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "duplicate column name")
}

// InitDB initializes the SQLite database
func InitDB(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}

	createTables := `
	CREATE TABLE IF NOT EXISTS app_version (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		latest_version TEXT NOT NULL,
		latest_version_code INTEGER NOT NULL,
		minimum_version_code INTEGER NOT NULL,
		force_update BOOLEAN NOT NULL,
		update_title TEXT NOT NULL,
		update_message TEXT NOT NULL,
		whats_new TEXT NOT NULL,
		release_date TEXT NOT NULL,
		download_url TEXT NOT NULL DEFAULT '',
		download_url_v8a TEXT NOT NULL DEFAULT '',
		download_url_v7a TEXT NOT NULL DEFAULT '',
		update_button_text TEXT NOT NULL DEFAULT 'ယခုဒေါင်းလုပ်',
		later_button_text TEXT NOT NULL DEFAULT 'နောက်မှ',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS in_app_messages (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		title TEXT NOT NULL,
		message TEXT NOT NULL,
		image_url TEXT,
		action_text TEXT,
		action_url TEXT,
		priority INTEGER NOT NULL,
		start_date TEXT NOT NULL,
		end_date TEXT NOT NULL,
		show_once BOOLEAN NOT NULL,
		dismissible BOOLEAN NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS banners (
		id TEXT PRIMARY KEY,
		image_url TEXT NOT NULL,
		link_url TEXT NOT NULL DEFAULT '',
		title TEXT NOT NULL DEFAULT '',
		is_active BOOLEAN NOT NULL DEFAULT 1,
		sort_order INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err = db.Exec(createTables); err != nil {
		return err
	}

	// Migrate existing DB: add new columns if they don't exist yet
	migrations := []string{
		`ALTER TABLE app_version ADD COLUMN download_url TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE app_version ADD COLUMN download_url_v8a TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE app_version ADD COLUMN download_url_v7a TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE app_version ADD COLUMN update_button_text TEXT NOT NULL DEFAULT 'ယခုဒေါင်းလုပ်'`,
		`ALTER TABLE app_version ADD COLUMN later_button_text TEXT NOT NULL DEFAULT 'နောက်မှ'`,
	}
	for _, m := range migrations {
		if _, err = db.Exec(m); err != nil {
			if !isDuplicateColumnError(err) {
				return err
			}
		}
	}

	if err = insertDefaultData(); err != nil {
		return err
	}

	log.Println("✅ Database initialized successfully")
	return nil
}

func insertDefaultData() error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM app_version").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		whatsNew := []string{
			"2D / 3D ထိုးကြေး Real-time နှင့် ကြည့်ရှုနိုင်",
			"ငွေဖြည့် / ငွေထုတ် မှတ်တမ်းကြည့်နိုင်",
			"မြန်မာဘာသာဖြင့် အသိပေးချက်များ",
			"ဂဏန်းရလဒ်ထွက်သည့်အခါ notification ရနိုင်",
			"Bug fixes and performance improvements",
		}
		whatsNewJSON, _ := json.Marshal(whatsNew)

		_, err = db.Exec(`
			INSERT INTO app_version (
				id, latest_version, latest_version_code, minimum_version_code,
				force_update, update_title, update_message,
				whats_new, release_date,
				download_url, download_url_v8a, download_url_v7a,
				update_button_text, later_button_text
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			1,
			"1.0.0", 1, 1,
			false,
			"အပ်ဒိတ်အသစ်ရှိပါသည်",
			"အသစ်သောအင်္ဂါရပ်များနှင့် တိုးတက်မှုများ ထည့်သွင်းထားပါသည်",
			string(whatsNewJSON),
			time.Now().Format("2006-01-02"),
			"",  // download_url (universal fallback)
			"",  // download_url_v8a — set via POST /api/power2d/config
			"",  // download_url_v7a — set via POST /api/power2d/config
			"ယခုဒေါင်းလုပ်",
			"နောက်မှ",
		)
		if err != nil {
			return err
		}

		messages := []InAppMessage{
			{
				ID:          "welcome_power2d",
				Type:        "info",
				Title:       "Power 2D မှ ကြိုဆိုပါသည်",
				Message:     "Power2D သို့ ကြိုဆိုပါသည်။ 2D / 3D ထိုးနိုင်ပြီး ငွေဖြည့် / ငွေထုတ် လွယ်ကူစွာ ပြုလုပ်နိုင်ပါသည်",
				ImageURL:    "",
				ActionText:  "",
				ActionURL:   "",
				Priority:    5,
				StartDate:   time.Now().Format("2006-01-02"),
				EndDate:     "2099-12-31",
				ShowOnce:    true,
				Dismissible: true,
			},
		}
		for _, msg := range messages {
			_, err = db.Exec(`
				INSERT INTO in_app_messages (
					id, type, title, message, image_url, action_text, action_url,
					priority, start_date, end_date, show_once, dismissible
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				msg.ID, msg.Type, msg.Title, msg.Message, msg.ImageURL,
				msg.ActionText, msg.ActionURL, msg.Priority, msg.StartDate,
				msg.EndDate, msg.ShowOnce, msg.Dismissible,
			)
			if err != nil {
				return err
			}
		}
		log.Println("✅ Default data inserted")
	}
	return nil
}

// ── AppVersion CRUD ───────────────────────────────────────────────────────────

func GetAppVersion() (*AppVersion, error) {
	var version AppVersion
	var whatsNewJSON string
	err := db.QueryRow(`
		SELECT latest_version, latest_version_code, minimum_version_code,
		       force_update, update_title, update_message,
		       whats_new, release_date,
		       download_url, download_url_v8a, download_url_v7a,
		       update_button_text, later_button_text
		FROM app_version WHERE id = 1
	`).Scan(
		&version.LatestVersion, &version.LatestVersionCode, &version.MinimumVersionCode,
		&version.ForceUpdate, &version.UpdateTitle, &version.UpdateMessage,
		&whatsNewJSON, &version.ReleaseDate,
		&version.DownloadURL, &version.DownloadURLv8a, &version.DownloadURLv7a,
		&version.UpdateButtonText, &version.LaterButtonText,
	)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal([]byte(whatsNewJSON), &version.WhatsNew); err != nil {
		version.WhatsNew = []string{}
	}
	return &version, nil
}

func UpdateAppVersion(version *AppVersion) error {
	whatsNewJSON, err := json.Marshal(version.WhatsNew)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		UPDATE app_version SET
			latest_version = ?, latest_version_code = ?, minimum_version_code = ?,
			force_update = ?, update_title = ?, update_message = ?,
			whats_new = ?, release_date = ?,
			download_url = ?, download_url_v8a = ?, download_url_v7a = ?,
			update_button_text = ?, later_button_text = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`,
		version.LatestVersion, version.LatestVersionCode, version.MinimumVersionCode,
		version.ForceUpdate, version.UpdateTitle, version.UpdateMessage,
		string(whatsNewJSON), version.ReleaseDate,
		version.DownloadURL, version.DownloadURLv8a, version.DownloadURLv7a,
		version.UpdateButtonText, version.LaterButtonText,
	)
	return err
}

// ── InAppMessage CRUD ─────────────────────────────────────────────────────────

func GetInAppMessages() ([]InAppMessage, error) {
	rows, err := db.Query(`
		SELECT id, type, title, message, image_url, action_text, action_url,
		       priority, start_date, end_date, show_once, dismissible
		FROM in_app_messages ORDER BY priority DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []InAppMessage
	for rows.Next() {
		var msg InAppMessage
		var imageURL, actionText, actionURL sql.NullString
		err = rows.Scan(
			&msg.ID, &msg.Type, &msg.Title, &msg.Message,
			&imageURL, &actionText, &actionURL,
			&msg.Priority, &msg.StartDate, &msg.EndDate,
			&msg.ShowOnce, &msg.Dismissible,
		)
		if err != nil {
			return nil, err
		}
		msg.ImageURL = imageURL.String
		msg.ActionText = actionText.String
		msg.ActionURL = actionURL.String
		messages = append(messages, msg)
	}
	return messages, nil
}

func UpsertInAppMessage(msg *InAppMessage) error {
	_, err := db.Exec(`
		INSERT INTO in_app_messages (
			id, type, title, message, image_url, action_text, action_url,
			priority, start_date, end_date, show_once, dismissible
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			type = excluded.type, title = excluded.title, message = excluded.message,
			image_url = excluded.image_url, action_text = excluded.action_text,
			action_url = excluded.action_url, priority = excluded.priority,
			start_date = excluded.start_date, end_date = excluded.end_date,
			show_once = excluded.show_once, dismissible = excluded.dismissible,
			updated_at = CURRENT_TIMESTAMP
	`,
		msg.ID, msg.Type, msg.Title, msg.Message, msg.ImageURL,
		msg.ActionText, msg.ActionURL, msg.Priority, msg.StartDate,
		msg.EndDate, msg.ShowOnce, msg.Dismissible,
	)
	return err
}

func DeleteInAppMessage(id string) error {
	_, err := db.Exec("DELETE FROM in_app_messages WHERE id = ?", id)
	return err
}

// ── Banner CRUD ───────────────────────────────────────────────────────────────

func GetBanners(activeOnly bool) ([]Banner, error) {
	query := `SELECT id, image_url, link_url, title, is_active, sort_order FROM banners`
	if activeOnly {
		query += ` WHERE is_active = 1`
	}
	query += ` ORDER BY sort_order ASC, created_at ASC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var banners []Banner
	for rows.Next() {
		var b Banner
		if err := rows.Scan(&b.ID, &b.ImageURL, &b.LinkURL, &b.Title, &b.IsActive, &b.SortOrder); err != nil {
			return nil, err
		}
		banners = append(banners, b)
	}
	if banners == nil {
		banners = []Banner{}
	}
	return banners, nil
}

func UpsertBanner(b *Banner) error {
	_, err := db.Exec(`
		INSERT INTO banners (id, image_url, link_url, title, is_active, sort_order)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			image_url = excluded.image_url,
			link_url  = excluded.link_url,
			title     = excluded.title,
			is_active = excluded.is_active,
			sort_order = excluded.sort_order,
			updated_at = CURRENT_TIMESTAMP
	`, b.ID, b.ImageURL, b.LinkURL, b.Title, b.IsActive, b.SortOrder)
	return err
}

func DeleteBanner(id string) error {
	_, err := db.Exec("DELETE FROM banners WHERE id = ?", id)
	return err
}

// ── Handlers ──────────────────────────────────────────────────────────────────

func GetAppConfigHandler(c *gin.Context) {
	version, err := GetAppVersion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get version: " + err.Error()})
		return
	}
	messages, err := GetInAppMessages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, AppConfig{AppVersion: *version, InAppMessages: messages})
	log.Println("✅ App config sent to client")
}

func UpdateAppConfigHandler(c *gin.Context) {
	var newConfig AppConfig
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := UpdateAppVersion(&newConfig.AppVersion); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update version: " + err.Error()})
		return
	}
	if _, err := db.Exec("DELETE FROM in_app_messages"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear messages: " + err.Error()})
		return
	}
	for _, msg := range newConfig.InAppMessages {
		if err := UpsertInAppMessage(&msg); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert message: " + err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Configuration updated successfully", "config": newConfig})
	log.Println("✅ App config updated")
}

func GetAppVersionHandler(c *gin.Context) {
	version, err := GetAppVersion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, version)
	log.Println("✅ App version info sent")
}

func GetInAppMessagesHandler(c *gin.Context) {
	messages, err := GetInAppMessages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, messages)
	log.Println("✅ In-app messages sent")
}

// ── Banner Handlers ───────────────────────────────────────────────────────────

func GetBannersHandler(c *gin.Context) {
	banners, err := GetBanners(true) // active only for Flutter app
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, banners)
}

func GetAllBannersHandler(c *gin.Context) {
	banners, err := GetBanners(false) // all banners for admin
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, banners)
}

func UpsertBannerHandler(c *gin.Context) {
	var b Banner
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// For PUT /admin/banners/:id, use the URL param as the ID
	if pathID := c.Param("id"); pathID != "" {
		b.ID = pathID
	}
	if b.ID == "" {
		b.ID = fmt.Sprintf("banner_%d", time.Now().UnixNano())
	}
	if err := UpsertBanner(&b); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("✅ Banner upserted: %s", b.ID)
	c.JSON(http.StatusOK, b)
}

func DeleteBannerHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
		return
	}
	if err := DeleteBanner(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("✅ Banner deleted: %s", id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ── Rate limiter ─────────────────────────────────────────────────────────────

type rateBucket struct {
	count    int
	resetAt  time.Time
}

var (
	rateMu      sync.Mutex
	rateBuckets = make(map[string]*rateBucket)
)

const (
	rateWindow   = time.Minute
	rateMaxReads = 60  // Flutter app polling — generous
	rateMaxAdmin = 20  // Admin writes — strict
)

func clientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	return c.ClientIP()
}

func checkRate(ip string, max int) bool {
	rateMu.Lock()
	defer rateMu.Unlock()
	b, ok := rateBuckets[ip]
	if !ok || time.Now().After(b.resetAt) {
		rateBuckets[ip] = &rateBucket{count: 1, resetAt: time.Now().Add(rateWindow)}
		return true
	}
	if b.count >= max {
		return false
	}
	b.count++
	return true
}

// ── Middleware ────────────────────────────────────────────────────────────────

// readKeyMiddleware: allows Flutter app (read-only key) OR admin key for GET endpoints.
func readKeyMiddleware(readKey, adminKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := clientIP(c)
		if !checkRate(ip, rateMaxReads) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		provided := c.GetHeader("X-API-Key")
		if provided == "" {
			provided = c.Query("api_key")
		}
		// Also accept the admin key on GET routes (admin dashboard needs to load config)
		adminProvided := c.GetHeader("X-Admin-Key")
		if provided != readKey && adminProvided != adminKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

// adminKeyMiddleware: only accepts the admin write key — never shared with apps.
func adminKeyMiddleware(adminKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := clientIP(c)
		if !checkRate(ip+":admin", rateMaxAdmin) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		provided := c.GetHeader("X-Admin-Key")
		if provided == "" || provided != adminKey {
			// Log suspicious attempts
			log.Printf("[SECURITY] ⚠️  Rejected admin write attempt from %s", ip)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

// ── Image Upload ──────────────────────────────────────────────────────────────

func UploadImageHandler(c *gin.Context) {
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no image file provided"})
		return
	}
	defer file.Close()

	// Only allow images
	ct := header.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only image files are allowed"})
		return
	}

	// Build unique filename: timestamp + original name
	ext := ".jpg"
	if idx := strings.LastIndex(header.Filename, "."); idx >= 0 {
		ext = header.Filename[idx:]
	}
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := "./uploads/" + filename

	if err := os.MkdirAll("./uploads", 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create uploads dir"})
		return
	}

	out, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save file"})
		return
	}
	defer out.Close()

	buf := make([]byte, 32*1024)
	for {
		n, readErr := file.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "write error"})
				return
			}
		}
		if readErr != nil {
			break
		}
	}

	// Return the public URL
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	// Use PUBLIC_URL env var if set (needed when server is behind proxy or accessed from phones)
	publicURL := os.Getenv("PUBLIC_URL")
	var url string
	if publicURL != "" {
		url = fmt.Sprintf("%s/uploads/%s", strings.TrimRight(publicURL, "/"), filename)
	} else {
		host := c.Request.Host
		url = fmt.Sprintf("%s://%s/uploads/%s", scheme, host, filename)
	}
	log.Printf("✅ Image uploaded: %s", url)
	c.JSON(http.StatusOK, gin.H{"url": url})
}

// fixLocalhostURLs replaces localhost/127.0.0.1 image URLs stored in the DB
// with the correct PUBLIC_URL. Runs once on startup.
func fixLocalhostURLs(publicURL string) {
	pub := strings.TrimRight(publicURL, "/")
	// Replace both localhost:PORT and 127.0.0.1:PORT variants
	replacements := []string{
		"http://localhost:4598",
		"https://localhost:4598",
		"http://127.0.0.1:4598",
		"https://127.0.0.1:4598",
	}
	total := 0
	for _, old := range replacements {
		res, err := db.Exec(
			`UPDATE in_app_messages SET image_url = REPLACE(image_url, ?, ?) WHERE image_url LIKE ?`,
			old, pub, "%"+old+"%",
		)
		if err == nil {
			n, _ := res.RowsAffected()
			total += int(n)
		}
		res2, err2 := db.Exec(
			`UPDATE banners SET image_url = REPLACE(image_url, ?, ?) WHERE image_url LIKE ?`,
			old, pub, "%"+old+"%",
		)
		if err2 == nil {
			n, _ := res2.RowsAffected()
			total += int(n)
		}
	}
	if total > 0 {
		log.Printf("✅ Fixed %d image URL(s): localhost → %s", total, pub)
	}
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	// Load .env file first (values are ignored if already set by OS env)
	loadEnvFile(".env")

	dbPath := "./appconfig.db"
	if err := InitDB(dbPath); err != nil {
		log.Fatal("❌ Failed to initialize database: ", err)
	}
	defer db.Close()

	// Fix any stored localhost image URLs → use real PUBLIC_URL
	publicURL := os.Getenv("PUBLIC_URL")
	if publicURL != "" {
		fixLocalhostURLs(publicURL)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Read keys from environment
	readKey := os.Getenv("POWER2D_READ_KEY")
	if readKey == "" {
		readKey = "power2d-client-readonly-2026"
		log.Printf("⚠️  POWER2D_READ_KEY not set — using default (change in production!)")
	}
	adminKey := os.Getenv("POWER2D_ADMIN_KEY")
	if adminKey == "" {
		adminKey = "power2d-admin-write-2026-secret"
		log.Printf("⚠️  POWER2D_ADMIN_KEY not set — using default (MUST change in production!)")
	}
	log.Printf("🔑 Two-key security enabled (read + admin write)")

	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept", "X-API-Key", "X-Admin-Key"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))

	api := r.Group("/api/power2d")
	{
		// READ routes — Flutter app uses X-API-Key (read-only) OR admin dashboard uses X-Admin-Key
		api.GET("/config", readKeyMiddleware(readKey, adminKey), GetAppConfigHandler)
		api.GET("/version", readKeyMiddleware(readKey, adminKey), GetAppVersionHandler)
		api.GET("/messages", readKeyMiddleware(readKey, adminKey), GetInAppMessagesHandler)
		api.GET("/banners", readKeyMiddleware(readKey, adminKey), GetBannersHandler)          // Flutter: active banners only

		// WRITE routes — Admin dashboard only, uses X-Admin-Key (never in APK)
		api.POST("/config", adminKeyMiddleware(adminKey), UpdateAppConfigHandler)
		api.GET("/admin/banners", adminKeyMiddleware(adminKey), GetAllBannersHandler)         // Admin: all banners
		api.POST("/admin/banners", adminKeyMiddleware(adminKey), UpsertBannerHandler)         // Admin: create/update
		api.PUT("/admin/banners/:id", adminKeyMiddleware(adminKey), UpsertBannerHandler)      // Admin: update
		api.DELETE("/admin/banners/:id", adminKeyMiddleware(adminKey), DeleteBannerHandler)   // Admin: delete
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
			"db":     "sqlite3",
		})
	})

	// Serve uploaded images statically — e.g. http://localhost:4598/uploads/abc.jpg
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		log.Printf("⚠️  Could not create uploads dir: %v", err)
	}
	r.Static("/uploads", "./uploads")

	// Image upload — admin key required
	r.POST("/api/power2d/upload-image", adminKeyMiddleware(adminKey), UploadImageHandler)

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":    "Power 2D - App Config Server",
			"version": "2.0.0",
			"status":  "running",
			"endpoints": []string{
				"GET  /api/power2d/config   - Full config (version + messages)",
				"POST /api/power2d/config   - Update config (admin)",
				"GET  /api/power2d/version  - Version check only",
				"GET  /api/power2d/messages - In-app messages only",
				"GET  /health               - Health check",
			},
		})
	})

	port := "4598"
	log.Printf("🚀 Power 2D Config Server on port %s\n", port)
	log.Printf("📡 http://localhost:%s/api/power2d/config\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("❌ Failed to start server: ", err)
	}
}
