# Burma 2D - App Config Server

🚀 **Standalone server for managing app configuration, version updates, and in-app messages**

## 🎯 Purpose

This is a **backup/independent** server specifically for:
- ✅ App version management & update notifications
- ✅ In-app message/announcement system
- ✅ No authentication required (simple GET endpoint)
- ✅ Separate from main Burma 2D server (runs on port 8080)

## 📡 API Endpoints

### Main Endpoint (Used by Android App)
```bash
GET http://localhost:8080/api/burma2d/config
```

**Response:**
```json
{
  "app_version": {
    "latest_version": "1.5.0",
    "latest_version_code": 150,
    "minimum_version_code": 100,
    "force_update": false,
    "update_title": "🎉 New Update Available!",
    "update_message": "We've added exciting new features...",
    "download_url": "https://github.com/Lainlain/burma2daio/releases/latest",
    "whats_new": [
      "✨ Brand new App Config system",
      "💬 Real-time chat with auto-reconnect",
      "🎁 Interactive gift system"
    ],
    "release_date": "2025-11-04"
  },
  "in_app_messages": [
    {
      "id": "welcome_2025",
      "type": "info",
      "title": "🎊 Welcome to Burma 2D 2025!",
      "message": "Thank you for using our app...",
      "priority": 5,
      "start_date": "2025-11-01",
      "end_date": "2025-12-31",
      "show_once": true,
      "dismissible": true
    }
  ]
}
```

### Individual Endpoints

#### Get Version Only
```bash
GET http://localhost:8080/api/burma2d/version
```

#### Get Messages Only
```bash
GET http://localhost:8080/api/burma2d/messages
```

#### Update Config (Admin)
```bash
POST http://localhost:8080/api/burma2d/config
Content-Type: application/json

{
  "app_version": {...},
  "in_app_messages": [...]
}
```

#### Health Check
```bash
GET http://localhost:8080/health
```

## 🚀 Quick Start

### 1. Install Dependencies
```bash
cd appconfig-server
go mod download
```

### 2. Build & Run
```bash
# Build
go build -o appconfig-server main.go

# Run
./appconfig-server
```

### 3. Test
```bash
# Test main endpoint
curl http://localhost:8080/api/burma2d/config | jq

# Test version endpoint
curl http://localhost:8080/api/burma2d/version | jq

# Test messages endpoint
curl http://localhost:8080/api/burma2d/messages | jq

# Health check
curl http://localhost:8080/health
```

## 📱 Android Integration

Update `Constants.kt`:
```kotlin
object Constants {
    // For local testing (emulator)
    const val CONFIG_BASE_URL = "http://10.0.2.2:8080/"
    
    // For production
    // const val CONFIG_BASE_URL = "https://config.num.guru/"
}
```

Update `AppConfigService.kt`:
```kotlin
private const val BASE_URL = Constants.CONFIG_BASE_URL

interface AppConfigService {
    @GET("api/burma2d/config")
    suspend fun getAppConfig(): AppConfig
}
```

## 🎨 Features

### App Version Management
- **Version Comparison**: Shows update dialog if new version available
- **Force Update**: Can require minimum version (blocks app if too old)
- **What's New**: List of new features in latest version
- **Download URL**: Direct link to APK/Play Store

### In-App Messages
- **4 Message Types**:
  - `info` - General announcements (blue badge)
  - `warning` - Important notices (orange badge)
  - `promo` - Promotions/offers (gold badge)
  - `alert` - Critical alerts (red badge)

- **Smart Filtering**:
  - Date range (start_date to end_date)
  - Priority sorting (higher = shown first)
  - Show once tracking (won't show again if dismissed)

- **Rich Content**:
  - Optional images
  - Action buttons with URLs
  - Dismissible/non-dismissible options

## 🛠️ Configuration Management

### Update Version Info
Edit `main.go` → `currentConfig.AppVersion`:
```go
AppVersion: AppVersion{
    LatestVersion:      "1.6.0",  // Displayed version
    LatestVersionCode:  160,       // Numeric for comparison
    MinimumVersionCode: 150,       // Minimum required version
    ForceUpdate:        true,      // Block app if below minimum
    UpdateTitle:        "Critical Update Required",
    UpdateMessage:      "Please update to continue using the app",
    DownloadURL:        "https://...",
    WhatsNew: []string{
        "🔥 New feature 1",
        "⚡ Performance boost",
    },
}
```

### Add In-App Message
Append to `currentConfig.InAppMessages`:
```go
{
    ID:          "black_friday_2025",
    Type:        "promo",
    Title:       "🎉 Black Friday Sale!",
    Message:     "50% off premium features...",
    ImageURL:    "https://...",
    ActionText:  "Get Offer",
    ActionURL:   "https://...",
    Priority:    10,
    StartDate:   "2025-11-25",
    EndDate:     "2025-11-30",
    ShowOnce:    false,
    Dismissible: true,
}
```

## 🔧 Production Deployment

### Systemd Service
Create `/etc/systemd/system/appconfig-server.service`:
```ini
[Unit]
Description=Burma 2D App Config Server
After=network.target

[Service]
Type=simple
User=burma2d
WorkingDirectory=/opt/burma2d/appconfig-server
ExecStart=/opt/burma2d/appconfig-server/appconfig-server
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### Commands
```bash
sudo systemctl daemon-reload
sudo systemctl enable appconfig-server
sudo systemctl start appconfig-server
sudo systemctl status appconfig-server

# View logs
journalctl -u appconfig-server -f
```

### Nginx Reverse Proxy
```nginx
server {
    listen 80;
    server_name config.num.guru;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 📊 Default Test Data

The server comes with **3 sample messages**:

1. **Welcome Message** (INFO)
   - Shows once on first app launch
   - Valid Nov 1 - Dec 31, 2025

2. **November Promo** (PROMO)
   - Repeatable promotional message
   - Valid for November 2025

3. **Maintenance Alert** (WARNING)
   - System maintenance notification
   - Valid Nov 4 - Nov 10, 2025

## 🔄 Dynamic Updates

The config is **stored in memory** (currentConfig variable). To persist changes:

### Option 1: JSON File (Simple)
```go
// Load on startup
configFile, _ := os.ReadFile("config.json")
json.Unmarshal(configFile, &currentConfig)

// Save on update
configData, _ := json.MarshalIndent(currentConfig, "", "  ")
os.WriteFile("config.json", configData, 0644)
```

### Option 2: Database (Advanced)
- Store in SQLite/PostgreSQL
- Version history tracking
- Rollback capability

### Option 3: Environment Variables
```go
latestVersion := os.Getenv("LATEST_VERSION")
downloadURL := os.Getenv("DOWNLOAD_URL")
```

## 🧪 Testing Scenarios

### Test Update Dialog
1. Set `latest_version_code: 200` in server
2. App version code is 150
3. Launch app → See update dialog with "What's New"

### Test Force Update
1. Set `minimum_version_code: 160`, `force_update: true`
2. App version code is 150
3. Launch app → Non-cancelable update dialog

### Test In-App Message
1. Add message with `start_date: today`, `end_date: tomorrow`
2. Launch app after 500ms
3. See bottom sheet with message content

### Test Show Once
1. Message with `show_once: true`
2. Dismiss message
3. Restart app → Message won't show again

## 🎯 Port Configuration

Default: **8080** (different from main Burma 2D server on 4545)

To change port, edit `main.go`:
```go
port := "9000"  // Your preferred port
```

## 📈 Future Enhancements

- [ ] Admin web panel for config management
- [ ] A/B testing for messages
- [ ] Analytics tracking (message views, update downloads)
- [ ] Push notification integration
- [ ] Multi-language support
- [ ] Message scheduling
- [ ] User segmentation (show different messages to different users)

## 🐛 Troubleshooting

### Server won't start
```bash
# Check if port 8080 is in use
lsof -i :8080
# Kill process using port
kill -9 <PID>
```

### Android can't connect (emulator)
```kotlin
// Use 10.0.2.2 instead of localhost
const val CONFIG_BASE_URL = "http://10.0.2.2:8080/"
```

### Android can't connect (physical device)
```kotlin
// Use your computer's IP address
const val CONFIG_BASE_URL = "http://192.168.1.100:8080/"
```

### CORS errors
- Already configured with `AllowOrigins: []string{"*"}`
- Check browser console for exact error

## 📝 License

Part of Burma 2D 2025 project.

---

**Created**: November 4, 2025  
**Purpose**: Standalone app configuration server (backup system)  
**No Authentication Required**: Simple GET endpoints for easy integration
# burma2dconfig
# 2dexpectconfig
# thutaconfig
# bettingconfig
