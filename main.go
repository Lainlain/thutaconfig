package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

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
}

// InAppMessage represents an in-app announcement/message
type InAppMessage struct {
	ID          string `json:"i"`
	Type        string `json:"tp"`
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

// AdConfig represents ad-related configuration
type AdConfig struct {
	NativeAdTimerSeconds int `json:"nt"` // Timer in seconds (60, 90, 120, etc.)
}

// AdUnits represents AdMob ad unit IDs
type AdUnits struct {
	BannerAdUnit       string `json:"ba"`
	InterstitialAdUnit string `json:"ia"`
	NativeAdUnit       string `json:"na"`
	AppOpenAdUnit      string `json:"oa"`
}

// AppConfig is the complete configuration response
type AppConfig struct {
	AppVersion    AppVersion     `json:"av"`
	InAppMessages []InAppMessage `json:"ms"`
	AdConfig      AdConfig       `json:"ac"`
	AdUnits       AdUnits        `json:"adu"`
}

// InitDB initializes the SQLite database
func InitDB(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// Test connection
	if err = db.Ping(); err != nil {
		return err
	}

	// Create tables
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

	CREATE TABLE IF NOT EXISTS ad_config (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		native_ad_timer_seconds INTEGER NOT NULL DEFAULT 60,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS ad_units (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		banner_ad_unit TEXT NOT NULL,
		interstitial_ad_unit TEXT NOT NULL,
		native_ad_unit TEXT NOT NULL,
		app_open_ad_unit TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err = db.Exec(createTables); err != nil {
		return err
	}

	// Insert default data if empty
	if err = insertDefaultData(); err != nil {
		return err
	}

	log.Println("✅ Database initialized successfully")
	return nil
}

// insertDefaultData inserts default configuration if database is empty
func insertDefaultData() error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM app_version").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// Insert default app version
		whatsNew := []string{
			"🎯 Real-time 2D lottery live numbers",
			"🎁 Gifts & rewards system",
			"📰 Paper/Guide section with beautiful UI",
			"📺 3D lottery live results",
			"📅 Calendar and lottery history",
			"🐛 Performance improvements and bug fixes",
		}
		whatsNewJSON, _ := json.Marshal(whatsNew)

		_, err = db.Exec(`
			INSERT INTO app_version (
				id, latest_version, latest_version_code, minimum_version_code,
				force_update, update_title, update_message,
				whats_new, release_date
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			1,
			"1.0.1",
			2,
			1, // minimum supported version code
			false,
			"🎉 New Update Available!",
			"We've added exciting new features and improvements to enhance your experience!",
			string(whatsNewJSON),
			time.Now().Format("2006-01-02"),
		)
		if err != nil {
			return err
		}

		// Insert default messages
		messages := []InAppMessage{
			{
				ID:          "welcome_2026",
				Type:        "info",
				Title:       "🎊 Welcome to 2D Thu Ta!",
				Message:     "Thank you for using 2D Thu Ta! We're committed to providing you with the best lottery experience. Check out our new features and enjoy real-time updates!",
				ImageURL:    "",
				ActionText:  "",
				ActionURL:   "",
				Priority:    5,
				StartDate:   "2025-11-01",
				EndDate:     "2025-12-31",
				ShowOnce:    true,
				Dismissible: true,
			},
			{
				ID:          "promo_nov_2025",
				Type:        "promo",
				Title:       "🎁 Special November Offer!",
				Message:     "Get access to premium features this month! Exclusive analysis, predictions, and more. Don't miss out on this limited-time offer!",
				ImageURL:    "",
				ActionText:  "Learn More",
				ActionURL:   "https://play.google.com/store/apps/details?id=com.twod.thuta.twodthuta",
				Priority:    10,
				StartDate:   "2025-11-01",
				EndDate:     "2025-11-30",
				ShowOnce:    false,
				Dismissible: true,
			},
			{
				ID:          "maintenance_alert",
				Type:        "warning",
				Title:       "⚠️ Scheduled Maintenance",
				Message:     "We'll be performing system maintenance on November 10th from 2:00 AM to 4:00 AM. The app may be temporarily unavailable during this time.",
				ImageURL:    "",
				ActionText:  "",
				ActionURL:   "",
				Priority:    8,
				StartDate:   "2025-11-04",
				EndDate:     "2025-11-10",
				ShowOnce:    false,
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

	// Insert default ad config if empty
	var adCount int
	err = db.QueryRow("SELECT COUNT(*) FROM ad_config").Scan(&adCount)
	if err != nil {
		return err
	}

	if adCount == 0 {
		_, err = db.Exec(`
			INSERT INTO ad_config (id, native_ad_timer_seconds)
			VALUES (1, 60)
		`)
		if err != nil {
			return err
		}
		log.Println("✅ Default ad config inserted (60 seconds)")
	}

	// Insert default ad units if empty
	var adUnitsCount int
	err = db.QueryRow("SELECT COUNT(*) FROM ad_units").Scan(&adUnitsCount)
	if err != nil {
		return err
	}

	if adUnitsCount == 0 {
		_, err = db.Exec(`
			INSERT INTO ad_units (
				id, banner_ad_unit, interstitial_ad_unit, 
				native_ad_unit, app_open_ad_unit
			) VALUES (?, ?, ?, ?, ?)`,
			1,
			"ca-app-pub-3940256099942544/6300978111", // Test banner
			"ca-app-pub-3940256099942544/1033173712", // Test interstitial
			"ca-app-pub-3940256099942544/2247696110", // Test native
			"ca-app-pub-3940256099942544/9257395921", // Test app open
		)
		if err != nil {
			return err
		}
		log.Println("✅ Default ad units inserted (test ads)")
	}

	return nil
}

// GetAppVersion retrieves the app version from database
func GetAppVersion() (*AppVersion, error) {
	var version AppVersion
	var whatsNewJSON string

	err := db.QueryRow(`
		SELECT latest_version, latest_version_code, minimum_version_code,
		       force_update, update_title, update_message,
		       whats_new, release_date
		FROM app_version WHERE id = 1
	`).Scan(
		&version.LatestVersion,
		&version.LatestVersionCode,
		&version.MinimumVersionCode,
		&version.ForceUpdate,
		&version.UpdateTitle,
		&version.UpdateMessage,
		&whatsNewJSON,
		&version.ReleaseDate,
	)

	if err != nil {
		return nil, err
	}

	// Parse whats_new JSON
	if err = json.Unmarshal([]byte(whatsNewJSON), &version.WhatsNew); err != nil {
		version.WhatsNew = []string{}
	}

	return &version, nil
}

// UpdateAppVersion updates the app version in database
func UpdateAppVersion(version *AppVersion) error {
	whatsNewJSON, err := json.Marshal(version.WhatsNew)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		UPDATE app_version SET
			latest_version = ?,
			latest_version_code = ?,
			minimum_version_code = ?,
			force_update = ?,
			update_title = ?,
			update_message = ?,
			whats_new = ?,
			release_date = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`,
		version.LatestVersion,
		version.LatestVersionCode,
		version.MinimumVersionCode,
		version.ForceUpdate,
		version.UpdateTitle,
		version.UpdateMessage,
		string(whatsNewJSON),
		version.ReleaseDate,
	)

	return err
}

// GetInAppMessages retrieves all in-app messages from database
func GetInAppMessages() ([]InAppMessage, error) {
	rows, err := db.Query(`
		SELECT id, type, title, message, image_url, action_text, action_url,
		       priority, start_date, end_date, show_once, dismissible
		FROM in_app_messages
		ORDER BY priority DESC
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

// UpsertInAppMessage inserts or updates an in-app message
func UpsertInAppMessage(msg *InAppMessage) error {
	_, err := db.Exec(`
		INSERT INTO in_app_messages (
			id, type, title, message, image_url, action_text, action_url,
			priority, start_date, end_date, show_once, dismissible
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			type = excluded.type,
			title = excluded.title,
			message = excluded.message,
			image_url = excluded.image_url,
			action_text = excluded.action_text,
			action_url = excluded.action_url,
			priority = excluded.priority,
			start_date = excluded.start_date,
			end_date = excluded.end_date,
			show_once = excluded.show_once,
			dismissible = excluded.dismissible,
			updated_at = CURRENT_TIMESTAMP
	`,
		msg.ID, msg.Type, msg.Title, msg.Message, msg.ImageURL,
		msg.ActionText, msg.ActionURL, msg.Priority, msg.StartDate,
		msg.EndDate, msg.ShowOnce, msg.Dismissible,
	)

	return err
}

// DeleteInAppMessage deletes an in-app message by ID
func DeleteInAppMessage(id string) error {
	_, err := db.Exec("DELETE FROM in_app_messages WHERE id = ?", id)
	return err
}

// GetAdConfig retrieves the ad configuration from database
func GetAdConfig() (*AdConfig, error) {
	var config AdConfig

	err := db.QueryRow(`
		SELECT native_ad_timer_seconds
		FROM ad_config WHERE id = 1
	`).Scan(&config.NativeAdTimerSeconds)

	if err != nil {
		return nil, err
	}

	return &config, nil
}

// UpdateAdConfig updates the ad configuration in database
func UpdateAdConfig(config *AdConfig) error {
	_, err := db.Exec(`
		UPDATE ad_config
		SET native_ad_timer_seconds = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`, config.NativeAdTimerSeconds)

	return err
}

// GetAdUnits retrieves the ad units from database
func GetAdUnits() (*AdUnits, error) {
	var units AdUnits

	err := db.QueryRow(`
		SELECT banner_ad_unit, interstitial_ad_unit, 
		       native_ad_unit, app_open_ad_unit
		FROM ad_units WHERE id = 1
	`).Scan(&units.BannerAdUnit, &units.InterstitialAdUnit,
		&units.NativeAdUnit, &units.AppOpenAdUnit)

	if err != nil {
		return nil, err
	}

	return &units, nil
}

// UpdateAdUnits updates the ad units in database
func UpdateAdUnits(units *AdUnits) error {
	_, err := db.Exec(`
		UPDATE ad_units
		SET banner_ad_unit = ?, interstitial_ad_unit = ?, 
		    native_ad_unit = ?, app_open_ad_unit = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`, units.BannerAdUnit, units.InterstitialAdUnit,
		units.NativeAdUnit, units.AppOpenAdUnit)

	return err
}

// GetAppConfigHandler returns the current app configuration
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

	adConfig, err := GetAdConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ad config: " + err.Error()})
		return
	}

	adUnits, err := GetAdUnits()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ad units: " + err.Error()})
		return
	}

	config := AppConfig{
		AppVersion:    *version,
		InAppMessages: messages,
		AdConfig:      *adConfig,
		AdUnits:       *adUnits,
	}

	c.JSON(http.StatusOK, config)
	log.Printf("✅ App config sent to client (ad timer: %d seconds)\n", adConfig.NativeAdTimerSeconds)
}

// UpdateAppConfigHandler allows updating the configuration
func UpdateAppConfigHandler(c *gin.Context) {
	var newConfig AppConfig
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update app version
	if err := UpdateAppVersion(&newConfig.AppVersion); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update version: " + err.Error()})
		return
	}

	// Delete all existing messages and insert new ones
	if _, err := db.Exec("DELETE FROM in_app_messages"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear messages: " + err.Error()})
		return
	}

	// Insert new messages
	for _, msg := range newConfig.InAppMessages {
		if err := UpsertInAppMessage(&msg); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert message: " + err.Error()})
			return
		}
	}

	// Update ad config
	if err := UpdateAdConfig(&newConfig.AdConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update ad config: " + err.Error()})
		return
	}

	// Update ad units
	if err := UpdateAdUnits(&newConfig.AdUnits); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update ad units: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration updated successfully",
		"config":  newConfig,
	})
	log.Printf("✅ App config updated in database (ad timer: %d seconds)\n", newConfig.AdConfig.NativeAdTimerSeconds)
}

// GetAdConfigOnlyHandler returns only ad configuration
func GetAdConfigOnlyHandler(c *gin.Context) {
	adConfig, err := GetAdConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, adConfig)
	log.Printf("✅ Ad config sent (timer: %d seconds)", adConfig.NativeAdTimerSeconds)
}

// UpdateAdConfigOnlyHandler updates only the ad configuration
func UpdateAdConfigOnlyHandler(c *gin.Context) {
	var adConfig AdConfig
	if err := c.ShouldBindJSON(&adConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate timer value (must be between 30 and 300 seconds)
	if adConfig.NativeAdTimerSeconds < 30 || adConfig.NativeAdTimerSeconds > 300 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Timer must be between 30 and 300 seconds"})
		return
	}

	if err := UpdateAdConfig(&adConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ad config updated successfully",
		"nt":      adConfig.NativeAdTimerSeconds,
	})
	log.Printf("✅ Ad timer updated to %d seconds", adConfig.NativeAdTimerSeconds)
}

// GetAdUnitsOnlyHandler returns only ad units
func GetAdUnitsOnlyHandler(c *gin.Context) {
	units, err := GetAdUnits()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, units)
	log.Println("✅ Ad units info sent")
}

// UpdateAdUnitsHandler updates ad units configuration
func UpdateAdUnitsHandler(c *gin.Context) {
	var units AdUnits
	if err := c.ShouldBindJSON(&units); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate ad unit IDs (must not be empty)
	if units.BannerAdUnit == "" || units.InterstitialAdUnit == "" ||
		units.NativeAdUnit == "" || units.AppOpenAdUnit == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All ad unit IDs must be provided"})
		return
	}

	if err := UpdateAdUnits(&units); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ad units updated successfully",
		"ba":      units.BannerAdUnit,
		"ia":      units.InterstitialAdUnit,
		"na":      units.NativeAdUnit,
		"oa":      units.AppOpenAdUnit,
	})
	log.Println("✅ Ad units updated successfully")
}

// GetAppVersionHandler returns only version information
func GetAppVersionHandler(c *gin.Context) {
	version, err := GetAppVersion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, version)
	log.Println("✅ App version info sent")
}

// GetInAppMessagesHandler returns only active messages
func GetInAppMessagesHandler(c *gin.Context) {
	messages, err := GetInAppMessages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
	log.Println("✅ In-app messages sent")
}

func main() {
	// Initialize database
	dbPath := "./appconfig.db"
	if err := InitDB(dbPath); err != nil {
		log.Fatal("❌ Failed to initialize database: ", err)
	}
	defer db.Close()

	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))

	// API Routes
	api := r.Group("/api/2dthuta")
	{
		// Main config endpoint (used by Android app)
		api.GET("/config", GetAppConfigHandler)

		// Individual endpoints
		api.GET("/version", GetAppVersionHandler)
		api.GET("/messages", GetInAppMessagesHandler)
		api.GET("/adconfig", GetAdConfigOnlyHandler)

		// Admin endpoint to update config
		api.POST("/config", UpdateAppConfigHandler)
		api.POST("/adconfig", UpdateAdConfigOnlyHandler)

		// Ad units endpoints
		api.GET("/adunits", GetAdUnitsOnlyHandler)
		api.POST("/adunits", UpdateAdUnitsHandler)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"time":     time.Now().Format(time.RFC3339),
			"database": "sqlite3",
		})
	})

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":     "2D Thu Ta - App Config Server",
			"version":  "1.0.0",
			"status":   "running",
			"database": "SQLite3",
			"endpoints": []string{
				"GET  /api/2dthuta/config   - Get full app configuration",
				"GET  /api/2dthuta/version  - Get version info only",
				"GET  /api/2dthuta/messages - Get in-app messages only",
				"POST /api/2dthuta/config   - Update configuration (admin)",
				"GET  /health               - Health check",
			},
		})
	})

	port := "4598"
	log.Printf("🚀 2D Thu Ta App Config Server starting on port %s...\n", port)
	log.Printf("📡 Main endpoint: http://localhost:%s/api/2dthuta/config\n", port)
	log.Printf("💾 Database: %s (SQLite3)\n", dbPath)
	log.Printf("🏥 Health check:  http://localhost:%s/health\n", port)
	log.Println("✅ Server ready to serve app configuration!")

	if err := r.Run(":" + port); err != nil {
		log.Fatal("❌ Failed to start server: ", err)
	}
}
