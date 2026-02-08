# 🚀 Hướng Dẫn Chạy HTTP Simulation System

> **Quick Start Guide** - Chạy hệ thống giả lập HTTP với phân tích và xử lý an toàn file thực thi

## 📍 Vị Trí Module

```
D:\PROJECT\Project\pack-a-mal\service-simulation-module\
```

## ⚡ Chạy Nhanh 3 Bước

### 1️⃣ Khởi Động Services

```powershell
cd D:\PROJECT\Project\pack-a-mal\service-simulation-module
docker-compose up -d
```

**Kết quả:**
```
[+] Running 3/3
 ✔ Network service-simulation-module_simulation_network  Created
 ✔ Container inetsim                                     Started
 ✔ Container service-simulation                          Started
```

⏱️ Đợi 10-15 giây để containers khởi động

### 2️⃣ Kiểm Tra

```powershell
curl http://localhost:5000/status -UseBasicParsing
```

**Kết quả:**
```json
{
  "service": "http-simulation",
  "status": "running",
  "version": "2.0",
  "features": [
    "http_analysis",
    "request_classification",
    "safe_executable_handling",
    "adaptive_response"
  ]
}
```

✅ Nếu thấy `"status": "running"` → Thành công!

### 3️⃣ Test Thử

```powershell
# Download executable (sẽ được sandbox)
curl http://localhost:5000/tools/installer.exe -OutFile test.exe

# Xem file đã sandbox
docker exec service-simulation ls -la /logs/executables/
```

**Kết quả:**
```
total 12
drwxr-xr-x 1 root root  512 Feb  8 07:03 .
drwxrwxrwx 1 root root  512 Feb  8 07:03 ..
-rw-r--r-- 1 root root  183 Feb  8 07:03 a1b2c3d4e5f67890_installer.exe
-rw-r--r-- 1 root root  965 Feb  8 07:03 a1b2c3d4e5f67890_installer.exe.metadata.json
-rw-r--r-- 1 root root  250 Feb  8 07:03 executable_requests.log
```

## 🎯 Các Tính Năng Chính

| Tính Năng | Mô Tả |
|-----------|-------|
| 🔍 **Request Analysis** | Phân tích chi tiết HTTP requests |
| 🏷️ **Classification** | Phân loại 9 categories tự động |
| 🛡️ **Security Detection** | Phát hiện XSS, SQLi, path traversal |
| 📦 **Safe Executable** | Sandbox file thực thi (không rủi ro) |
| 🍯 **Honeypot** | Track suspicious downloads |
| 📊 **Logging** | Chi tiết metadata + request logs |

## 📖 Tài Liệu Đầy Đủ

Xem file hướng dẫn chi tiết trong thư mục module:

```
service-simulation-module/
├── QUICKSTART.md              ⭐ Hướng dẫn chạy nhanh
├── HTTP_SIMULATION_GUIDE.md   📚 Tài liệu đầy đủ
├── QUICK_REFERENCE.md         📋 Quick reference
├── README.md                  📖 Overview
├── demo_http_simulation.py    🎬 Demo script
└── test_http_simulation.py    🧪 Test suite
```

### Quick Links

- [QUICKSTART.md](service-simulation-module/QUICKSTART.md) - Bắt đầu ngay
- [HTTP_SIMULATION_GUIDE.md](service-simulation-module/HTTP_SIMULATION_GUIDE.md) - Tài liệu chi tiết 
- [QUICK_REFERENCE.md](service-simulation-module/QUICK_REFERENCE.md) - Tham khảo nhanh
- [REPORT_HTTP_EXTENSION.md](Reports/REPORT_HTTP_EXTENSION.md) - Báo cáo kỹ thuật

## 🧪 Chạy Demo

```powershell
cd service-simulation-module

# Cài đặt requests
pip install requests

# Chạy demo (9 scenarios)
python demo_http_simulation.py

# Chạy tests (12 tests)
python test_http_simulation.py
```

## 🎬 Ví Dụ Sử Dụng

### Test Static Content
```powershell
curl http://localhost:5000/styles/main.css
curl http://localhost:5000/images/logo.png
```

**Kết quả (CSS):**
```css
/* Simulated CSS file */
body { font-family: Arial, sans-serif; }
```

**Kết quả (Image):**
→ Trả về 1x1 transparent PNG placeholder

### Test API Simulation
```powershell
curl http://localhost:5000/api/v1/users
```

**Kết quả:**
```json
{
  "status": "success",
  "timestamp": "2026-02-08T07:04:10.123456",
  "data": {
    "message": "API simulation response",
    "request_path": "/api/v1/users",
    "simulated": true
  }
}
```

### Test Executable Download
```powershell
# Safe download (low risk)
curl http://localhost:5000/installer.exe -OutFile installer.exe

# Suspicious download (medium risk - honeypot)
curl http://localhost:5000/malware.exe -H "User-Agent: Malware" -OutFile malware.exe
```

**Kiểm tra file đã download:**
```powershell
Get-Content installer.exe
```

**Kết quả:**
```
MZ
# SIMULATED EXECUTABLE
# Request ID: a1b2c3d4e5f67890
# Timestamp: 2026-02-08T07:03:57.152974
# Original file: installer.exe
# Platform: windows
# SAFE FOR ANALYSIS - NO REAL CODE
```

→ File được sandbox an toàn, không có code thực thi!

### Test Attack Detection
```powershell
# XSS
curl "http://localhost:5000/search?q=<script>alert('xss')</script>"

# Path traversal  
curl "http://localhost:5000/../../../etc/passwd"

# SQL injection
curl "http://localhost:5000/api?id=1' OR '1'='1"
```

### Analyze Request
```powershell
$body = @{
    method = "GET"
    url = "/download/malware.exe"
    headers = @{"User-Agent" = "Python/3.9"}
    client_ip = "192.168.1.100"
} | ConvertTo-Json

curl http://localhost:5000/analyze -Method Post -Body $body -ContentType "application/json"
```

**Kết quả:**
```json
{
  "classification": {
    "category": "executable_download",
    "sub_category": ".exe",
    "confidence": 0.95,
    "intent": "download_executable",
    "recommended_action": "sandbox_and_serve"
  },
  "analysis": {
    "method": "GET",
    "url": "/download/malware.exe",
    "file_extension": ".exe",
    "is_executable_request": true,
    "security_flags": {
      "risk_level": "low",
      "suspicious_patterns_found": []
    }
  },
  "summary": "GET request to /download/malware.exe from 192.168.1.100 (executable download)"
}
```

### View Logs
```powershell
# Xem executable download logs
curl http://localhost:5000/logs/executables

# Xem container logs
docker-compose logs -f service-simulation
```

**Kết quả logs mẫu:**
```
* Serving Flask app 'api.server'
* Debug mode: off
* Running on all addresses (0.0.0.0)
* Running on http://127.0.0.1:5000
Press CTRL+C to quit

172.20.0.1 - - [08/Feb/2026 07:03:57] "GET /status HTTP/1.1" 200 -
172.20.0.1 - - [08/Feb/2026 07:03:57] "GET /test.exe HTTP/1.1" 200 -
172.20.0.1 - - [08/Feb/2026 07:04:10] "GET /malware.exe HTTP/1.1" 200 -
172.20.0.1 - - [08/Feb/2026 07:04:10] "POST /analyze HTTP/1.1" 200 -
```

→ Mỗi dòng hiển thị: **Client IP**, **Timestamp**, **HTTP Method**, **URL**, **Status Code**

## � Hiểu Logs Của Hệ Thống

### Container Logs (docker logs)
```powershell
docker logs service-simulation --tail 20
# hoặc
docker-compose logs -f service-simulation  # realtime
```

**Logs hiển thị:**
1. **Startup logs** - Khi service khởi động
2. **HTTP access logs** - Mỗi request được log với:
   - Client IP (172.20.0.1)
   - Timestamp [08/Feb/2026 07:03:57]
   - HTTP Method và URL: "GET /status HTTP/1.1"
   - Status code: 200

### Executable Request Logs
```powershell
# Xem logs JSON của executable requests
docker exec service-simulation cat /logs/executables/executable_requests.log
```

**Format:**
```json
{"type": "executable_request", "request_id": "a1b2c3d4", "timestamp": "2026-02-08T07:03:57", "filename": "installer.exe", "extension": ".exe", "platform": "windows", "client_ip": "172.20.0.1", "risk_level": "low", "is_suspicious": false}
```

### Metadata Files
```powershell
# Xem metadata chi tiết của 1 request
docker exec service-simulation cat /logs/executables/*.metadata.json | python -m json.tool
```

**Chứa đầy đủ:**
- Request ID, timestamp
- Client info (IP, User-Agent)
- Risk assessment
- Handling strategy
- Security flags

## �📊 API Endpoints

| Endpoint | Method | Chức Năng |
|----------|--------|-----------|
| `/status` | GET | Service status |
| `/analyze` | POST | Phân tích request |
| `/simulate` | POST | Simulate & respond |
| `/logs/executables` | GET | View executable logs |
| `/*` | ANY | Auto-handle all requests |

## 🛑 Dừng Services

```powershell
cd service-simulation-module

# Dừng containers
docker-compose stop

# Dừng và xóa containers
docker-compose down
```

## 🔄 Restart/Rebuild

```powershell
# Restart
docker-compose restart

# Rebuild
docker-compose up -d --build
```

## 🐛 Troubleshooting Nhanh

| Vấn Đề | Giải Pháp |
|--------|-----------|
| Port in use | `docker-compose down` rồi `up` lại |
| Container không start | Xem logs: `docker-compose logs` |
| Import error | Rebuild: `docker-compose build --no-cache` |
| Connection refused | Đợi thêm vài giây, check `docker-compose ps` |

## 🎓 Kiến Trúc Hệ Thống

```
HTTP Request
    ↓
HTTPRequestAnalyzer (phân tích)
    ↓
RequestClassifier (phân loại)
    ↓
ResponseHandler (tạo response)
    ↓
SafeExecutableHandler (xử lý executables)
    ↓
Sandbox Storage + Logs
```

## 🔒 Bảo Mật

✅ **Không có executable thật nào được serve**  
✅ **Mọi file được sandbox hoàn toàn**  
✅ **Chi tiết logging cho forensics**  
✅ **Risk-based response strategies**

## 💡 Tips

1. Luôn check status trước khi test
2. Dùng demo script để hiểu flow
3. Check sandbox files để xem metadata
4. Xem logs realtime khi debug
5. Đọc QUICKSTART.md trong module folder

## 📞 Hỗ Trợ

1. Đọc [QUICKSTART.md](service-simulation-module/QUICKSTART.md)
2. Xem [HTTP_SIMULATION_GUIDE.md](service-simulation-module/HTTP_SIMULATION_GUIDE.md)
3. Check [REPORT_HTTP_EXTENSION.md](Reports/REPORT_HTTP_EXTENSION.md)
4. Xem logs: `docker-compose logs`

---

**Version:** 2.0  
**Ready:** ✅ Production  
**Updated:** February 8, 2026

🚀 **[Bắt Đầu Ngay](service-simulation-module/QUICKSTART.md)**
