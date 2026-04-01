# 🎉 App Config Server Setup Complete!

**Created**: November 5, 2025  
**Purpose**: Standalone backup server for app configuration management

---

## ✅ What Was Created

### 📁 New Directory: `appconfig-server/`

```
appconfig-server/
├── main.go                         # Go server with app config endpoints
├── go.mod                          # Go module dependencies
├── go.sum                          # Dependency checksums
├── README.md                       # Complete server documentation
├── ANDROID-INTEGRATION-GUIDE.md   # Android app integration guide
├── build-and-run.sh               # Quick build + run script
├── test-endpoints.sh              # API endpoint testing script
└── appconfig-server               # Compiled binary (13 MB)
```

### 🚀 Server Features

**Running on**: `http://localhost:8080`

**API Endpoints**:
- `GET /api/burma2d/config` - Full app configuration (main endpoint)
- `GET /api/burma2d/version` - App version info only
- `GET /api/burma2d/messages` - In-app messages only
- `POST /api/burma2d/config` - Update configuration (admin)
- `GET /health` - Health check

**Test Data Included**:
1. ✨ App Version 1.5.0 (latest_version_code: 150)
2. 🎊 Welcome Message (INFO type)
3. 🎁 November Promo (PROMO type)
4. ⚠️ Maintenance Alert (WARNING type)

---

## 🔧 Android Integration Changes

### 1. Updated `Constants.kt`
```kotlin
// Added new constant for config server
const val CONFIG_BASE_URL = "http://10.0.2.2:8080/"
```

### 2. Updated `AppConfigViewModel.kt`
```kotlin
// Changed Retrofit base URL to use CONFIG_BASE_URL
.baseUrl(Constants.CONFIG_BASE_URL)  // Separate server on port 8080
```

---

## 🧪 Server Status

### ✅ Server Running
```
Process ID: 79718
Port: 8080
Status: Healthy
```

### ✅ Tested Endpoints
```bash
# Health check
curl http://localhost:8080/health
# Response: {"status":"healthy","time":"2025-11-05T06:23:47+06:30"}

# App config
curl http://localhost:8080/api/burma2d/config
# Response: Full JSON with app_version and in_app_messages (3 messages)
```

---

## 📱 Next Steps to Test UI

### 1. Build & Install Android App
```bash
cd Kotlin/MVVM
./gradlew installDebug

# Or use Android Studio: Run ▶️
```

### 2. Expected Behavior
**500ms after app launch:**
1. ViewModel fetches config from `http://10.0.2.2:8080/api/burma2d/config`
2. Checks if app version < latest version (150)
3. Shows **Update Dialog** if update available
4. Shows **In-App Message Bottom Sheet** with highest priority message

### 3. Test Scenarios

**Scenario A: Test Update Dialog**
- Edit `appconfig-server/main.go`
- Change `LatestVersionCode: 200` (higher than current app)
- Rebuild server: `go build -o appconfig-server main.go`
- Restart server: `killall appconfig-server && nohup ./appconfig-server > server.log 2>&1 &`
- Launch app → See update dialog with "What's New" list

**Scenario B: Test In-App Message**
- App currently shows message with **priority 10** (November Promo)
- Displays in bottom sheet with PROMO badge (gold)
- "Learn More" button opens http://localhost:4545

**Scenario C: Test Force Update**
- Edit `MinimumVersionCode: 200` and `ForceUpdate: true`
- Launch app → Non-cancelable dialog appears

---

## 🛠️ Useful Commands

### Server Management
```bash
# Start server
cd appconfig-server
nohup ./appconfig-server > server.log 2>&1 & echo $!

# Check if running
curl http://localhost:8080/health

# Stop server
killall appconfig-server

# View logs
tail -f appconfig-server/server.log

# Rebuild after changes
cd appconfig-server
go build -o appconfig-server main.go
killall appconfig-server
nohup ./appconfig-server > server.log 2>&1 &
```

### Testing
```bash
# Test all endpoints
cd appconfig-server
./test-endpoints.sh

# Individual endpoint tests
curl -s http://localhost:8080/api/burma2d/config | jq
curl -s http://localhost:8080/api/burma2d/version | jq
curl -s http://localhost:8080/api/burma2d/messages | jq
```

### Android App
```bash
# Build and install
cd Kotlin/MVVM
./gradlew installDebug

# Clear app data (reset show-once messages)
adb shell pm clear com.burma.royal2d

# Watch Logcat
adb logcat | grep -E "AppConfigViewModel|UpdateDialog|InAppMessage"
```

---

## 📊 Server Configuration

### Current Test Data

**App Version:**
- Latest Version: 1.5.0 (code: 150)
- Minimum Required: 1.0.0 (code: 100)
- Force Update: No
- What's New: 6 features listed

**In-App Messages:**
1. **Welcome 2025** (INFO)
   - Priority: 5
   - Date Range: Nov 1 - Dec 31, 2025
   - Show Once: Yes
   
2. **November Promo** (PROMO) ⭐ Highest Priority
   - Priority: 10
   - Date Range: Nov 1 - Nov 30, 2025
   - Show Once: No
   - Action: "Learn More" → http://localhost:4545
   
3. **Maintenance Alert** (WARNING)
   - Priority: 8
   - Date Range: Nov 4 - Nov 10, 2025
   - Show Once: No

---

## 🎨 UI Components Ready

### ✅ Update Dialog (`UpdateDialogFragment`)
- Professional dark theme with gold accents
- Version comparison display (Your: X.X.X → Latest: Y.Y.Y)
- "What's New" bullet list with emojis
- Download button → Opens GitHub/Play Store
- Force update support (non-cancelable)

### ✅ In-App Message Bottom Sheet (`InAppMessageBottomSheet`)
- Type badges: INFO (blue), WARNING (orange), PROMO (gold), ALERT (red)
- Image support via Glide
- Rich message content (scrollable)
- Optional action button
- Dismissible/non-dismissible modes
- Show-once tracking via SharedPreferences

---

## 📝 Important Notes

### For Emulator Testing
- **MUST use**: `http://10.0.2.2:8080/`
- `10.0.2.2` = Host machine's localhost
- `localhost` won't work in emulator

### For Physical Device Testing
- Find your PC's IP: `ip addr show | grep "inet 192"`
- Update Constants.kt: `http://192.168.x.x:8080/`
- Ensure firewall allows port 8080
- Device must be on same WiFi network

### Server Independence
- Runs on **port 8080** (different from main Burma2D server on 4545)
- No authentication required (simple GET endpoints)
- No database needed (config stored in memory)
- Can run standalone or alongside main server

---

## 🚀 Production Deployment

When ready to deploy to VPS:

1. **Build for Linux**:
   ```bash
   GOOS=linux GOARCH=amd64 go build -o appconfig-server main.go
   ```

2. **Copy to VPS**:
   ```bash
   scp appconfig-server user@vps:/opt/burma2d/
   ```

3. **Create systemd service** (see README.md for complete setup)

4. **Update Android app**:
   ```kotlin
   const val CONFIG_BASE_URL = "https://config.num.guru/"
   ```

---

## ✅ Success Checklist

- [x] Go server created and built successfully
- [x] Server running on port 8080
- [x] All 5 endpoints tested and working
- [x] Test data configured (1 version + 3 messages)
- [x] README.md with complete documentation
- [x] ANDROID-INTEGRATION-GUIDE.md with step-by-step setup
- [x] Build scripts created (build-and-run.sh, test-endpoints.sh)
- [x] Android Constants.kt updated with CONFIG_BASE_URL
- [x] Android AppConfigViewModel updated to use config server
- [ ] Android app built and tested with live data (next step)

---

## 🎯 Ready to Test!

**Your standalone app config server is now running!**

To see the UI in action:
1. Open Android Studio
2. Build and run the app on emulator
3. Wait 500ms after MainActivity loads
4. **In-App Message Bottom Sheet** should appear with "November Promo"
5. To trigger **Update Dialog**: Edit server config to set `LatestVersionCode: 200`

**Server URL**: http://localhost:8080  
**Android connects to**: http://10.0.2.2:8080  
**Process ID**: 79718

---

**All files created in**: `/home/lainlain/Desktop/Burma2D/appconfig-server/`

Enjoy your new app configuration system! 🚀
