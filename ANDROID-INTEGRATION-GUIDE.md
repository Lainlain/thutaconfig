# 📱 Android App Integration with App Config Server

## 🎯 Overview

Connect your Android app to the standalone App Config Server to enable:
- ✅ **Update Dialog** - Shows when new version is available
- ✅ **In-App Messages** - Rich announcements with images and actions

## 🔧 Step 1: Update Constants.kt

Add a separate base URL for the config server:

```kotlin
// File: Kotlin/MVVM/app/src/main/java/com/burma/royal2d/util/Constants.kt

object Constants {
    // Main Burma 2D API (existing)
    const val BASE_URL = "http://10.0.2.2:4545/"
    
    // NEW: App Config Server (port 8080)
    const val CONFIG_BASE_URL = "http://10.0.2.2:8080/"
    
    // Other constants...
}
```

**Important URL Notes:**
- **Emulator**: Use `http://10.0.2.2:8080/` (10.0.2.2 = host machine)
- **Physical Device (same WiFi)**: Use `http://192.168.x.x:8080/` (your PC's IP)
- **Production**: Use `https://config.num.guru/` (when deployed to VPS)

## 🔧 Step 2: Update AppConfigService.kt

Point the Retrofit service to the new config server:

```kotlin
// File: Kotlin/MVVM/app/src/main/java/com/burma/royal2d/data/api/AppConfigService.kt

import com.burma.royal2d.util.Constants
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import retrofit2.http.GET

interface AppConfigService {
    @GET("api/burma2d/config")
    suspend fun getAppConfig(): AppConfig
}

object AppConfigServiceBuilder {
    private val retrofit = Retrofit.Builder()
        .baseUrl(Constants.CONFIG_BASE_URL)  // Changed from BASE_URL
        .addConverterFactory(GsonConverterFactory.create())
        .build()

    val service: AppConfigService = retrofit.create(AppConfigService::class.java)
}
```

## ✅ Step 3: Verify Setup

### Check Server is Running
```bash
# On your development machine
curl http://localhost:8080/health
# Should return: {"status":"healthy","time":"2025-11-05T..."}

curl http://localhost:8080/api/burma2d/config | jq
# Should return full app config with version and messages
```

### Check Network Configuration

**For Emulator:**
```bash
# From emulator terminal/adb shell
adb shell ping 10.0.2.2
# Should successfully ping your host machine
```

**For Physical Device:**
```bash
# Find your PC's IP address
ip addr show | grep "inet 192"
# or
ifconfig | grep "inet 192"

# Use that IP in Constants.CONFIG_BASE_URL
# Example: http://192.168.1.100:8080/
```

## 🧪 Step 4: Test in Android Studio

### Build and Run
```bash
cd Kotlin/MVVM
./gradlew installDebug

# Or use Android Studio:
# 1. Open project in Android Studio
# 2. Select emulator/device
# 3. Click Run ▶️
```

### Expected Behavior on App Launch

**500ms after MainActivity starts:**
1. `AppConfigViewModel` fetches config from `http://10.0.2.2:8080/api/burma2d/config`
2. **If new version available** → Shows `UpdateDialogFragment`
3. **If active in-app message exists** → Shows `InAppMessageBottomSheet`

### Logcat Debugging

Watch for these logs:
```
D/AppConfigViewModel: Checking for updates...
D/AppConfigViewModel: Config received: AppConfig(...)
D/UpdateDialog: Showing update dialog for version 1.5.0
D/InAppMessage: Showing message: welcome_2025
```

## 🎨 Testing Different Scenarios

### Scenario 1: Test Update Dialog

**Edit server config** (appconfig-server/main.go):
```go
LatestVersion:      "2.0.0",
LatestVersionCode:  200,  // Higher than your app's version code
ForceUpdate:        false,
UpdateTitle:        "🔥 Major Update Available!",
UpdateMessage:      "Lots of new features!",
WhatsNew: []string{
    "✨ New Feature 1",
    "🚀 New Feature 2",
},
```

**Rebuild server:**
```bash
cd appconfig-server
go build -o appconfig-server main.go
killall appconfig-server
nohup ./appconfig-server > server.log 2>&1 &
```

**Launch app** → You should see professional update dialog with:
- Version comparison (Your: 1.0.0 → Latest: 2.0.0)
- What's New bullet list
- Download button

### Scenario 2: Test Force Update

**Edit server config:**
```go
MinimumVersionCode: 200,  // Higher than your app's version code
ForceUpdate:        true,
UpdateTitle:        "⚠️ Critical Update Required",
UpdateMessage:      "You must update to continue using the app",
```

**Launch app** → Dialog appears with:
- ❌ No close button (non-cancelable)
- ⚠️ Warning message
- Only "Download Update" action

### Scenario 3: Test In-App Messages

**Edit server config (add new message):**
```go
{
    ID:          "black_friday",
    Type:        "promo",
    Title:       "🎉 Black Friday Sale!",
    Message:     "50% off all premium features!",
    ImageURL:    "https://example.com/banner.jpg",
    ActionText:  "Claim Offer",
    ActionURL:   "http://localhost:4545/promo",
    Priority:    15,  // Highest priority
    StartDate:   "2025-11-05",  // Today
    EndDate:     "2025-11-06",  // Tomorrow
    ShowOnce:    false,
    Dismissible: true,
},
```

**Launch app** → Bottom sheet appears with:
- 🏷️ PROMO badge (gold background)
- Image loaded via Glide
- "Claim Offer" button → Opens URL

### Scenario 4: Test Show Once

**Set message:**
```go
ShowOnce:    true,
```

**First launch** → Message shows  
**Dismiss message**  
**Restart app** → Message won't show again (tracked in SharedPreferences)

## 🐛 Troubleshooting

### Issue: App can't connect to server

**Check 1: Server is running**
```bash
curl http://localhost:8080/health
# Should return {"status":"healthy"}
```

**Check 2: Correct URL in Constants.kt**
```kotlin
// For emulator - MUST use 10.0.2.2
const val CONFIG_BASE_URL = "http://10.0.2.2:8080/"

// For physical device - Use your PC's IP
const val CONFIG_BASE_URL = "http://192.168.1.100:8080/"
```

**Check 3: Firewall**
```bash
# Allow port 8080
sudo ufw allow 8080
```

**Check 4: Cleartext traffic allowed (AndroidManifest.xml)**
```xml
<application
    android:usesCleartextTraffic="true"
    ...>
```

### Issue: Update dialog not showing

**Check 1: Version codes**
```kotlin
// In app/build.gradle
versionCode 150

// Server must have higher version
LatestVersionCode:  200  // Must be > 150
```

**Check 2: Delay in MainActivity**
```kotlin
lifecycleScope.launch {
    delay(500)  // Give time for data fetch
    observeAppConfig()
}
```

### Issue: In-app message not showing

**Check 1: Date range**
```go
StartDate:   "2025-11-05",  // Must be <= today
EndDate:     "2025-12-31",  // Must be >= today
```

**Check 2: Already shown (show_once)**
```kotlin
// Clear SharedPreferences to reset
adb shell pm clear com.burma.royal2d
```

**Check 3: Priority sorting**
```kotlin
// ViewModel filters and sorts by priority DESC
// Message with priority 15 shows before priority 5
```

### Issue: JSON parsing errors

**Check Logcat:**
```
E/AppConfigViewModel: Failed to fetch config: ...
```

**Test endpoint directly:**
```bash
curl -s http://localhost:8080/api/burma2d/config | jq
# Check if JSON is valid
```

## 📊 Current Test Data

The server comes with **3 sample messages**:

1. **Welcome Message** (INFO)
   - Priority: 5
   - Valid: Nov 1 - Dec 31, 2025
   - Show once: true

2. **November Promo** (PROMO)
   - Priority: 10
   - Valid: Nov 1 - Nov 30, 2025
   - Show once: false

3. **Maintenance Alert** (WARNING)
   - Priority: 8
   - Valid: Nov 4 - Nov 10, 2025
   - Show once: false

**Which message shows?**  
→ Highest priority that's not been dismissed (if show_once)  
→ In this case: **November Promo (priority 10)**

## 🚀 Production Deployment

### Server Deployment (VPS)
```bash
# Build for Linux
cd appconfig-server
GOOS=linux GOARCH=amd64 go build -o appconfig-server main.go

# SCP to VPS
scp appconfig-server user@your-vps:/opt/burma2d/

# SSH to VPS
ssh user@your-vps
cd /opt/burma2d

# Run with systemd (create service file)
sudo nano /etc/systemd/system/appconfig-server.service
```

**Service file:**
```ini
[Unit]
Description=Burma 2D App Config Server
After=network.target

[Service]
Type=simple
User=burma2d
WorkingDirectory=/opt/burma2d
ExecStart=/opt/burma2d/appconfig-server
Restart=always

[Install]
WantedBy=multi-user.target
```

**Start service:**
```bash
sudo systemctl daemon-reload
sudo systemctl enable appconfig-server
sudo systemctl start appconfig-server
```

### Android App (Production)

**Update Constants.kt:**
```kotlin
const val CONFIG_BASE_URL = "https://config.num.guru/"
```

**Setup Nginx reverse proxy:**
```nginx
server {
    listen 80;
    server_name config.num.guru;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
    }
}
```

## ✅ Checklist

Before testing:
- [ ] App config server is running on port 8080
- [ ] `curl http://localhost:8080/health` returns healthy
- [ ] Constants.CONFIG_BASE_URL is set correctly (10.0.2.2 for emulator)
- [ ] AndroidManifest has `usesCleartextTraffic="true"`
- [ ] Server has test data with today's date range
- [ ] App version code is lower than server's latest_version_code

Launch app and verify:
- [ ] Logcat shows "Checking for updates..."
- [ ] Update dialog appears (if version code < latest)
- [ ] In-app message bottom sheet appears
- [ ] Message has correct badge color
- [ ] Action button opens URL (if provided)
- [ ] Dismissing message works
- [ ] Show-once tracking works (message doesn't reappear)

---

**Next Steps**: Update your Android app's Constants.kt and test the UI! 🚀
