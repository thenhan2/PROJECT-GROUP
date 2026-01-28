# Hướng Dẫn Chạy Network Simulation

## Bước 1: Khởi động Docker Services

```powershell
# Di chuyển vào thư mục dynamic-analysis
cd D:\PROJECT\Project\pack-a-mal\dynamic-analysis

# Khởi động INetSim và Service Simulation
docker-compose -f docker-compose.network-sim.yml up -d
```

**Kết quả mong đợi:**
```
✔ Network pack-a-mal-network    Created
✔ Container pack-a-mal-inetsim  Healthy  
✔ Container pack-a-mal-sim-api  Started (healthy)
```

## Bước 2: Kiểm tra Services đang chạy

```powershell
# Xem trạng thái containers
docker ps --filter "name=pack-a-mal"
```

**Kết quả mong đợi:** Cả 2 containers hiển thị status `(healthy)`

## Bước 3: Test INetSim HTTP Service

```powershell
# Test HTTP service (port 8080)
curl.exe http://localhost:8080
```

**Kết quả mong đợi:** Trả về trang HTML với nội dung "INetSim default HTML page"

## Bước 4: Test Service Simulation API

```powershell
# Test API status
curl.exe http://localhost:5000/status
```

**Kết quả mong đợi:**
```json
{"service":"simulation","status":"running"}
```

## Bước 5: Cấu hình Environment Variables

```powershell
# Set biến môi trường cho network simulation
$env:OSSF_NETWORK_SIMULATION_ENABLED = "true"
$env:OSSF_INETSIM_DNS_ADDR = "172.20.0.2:53"
$env:OSSF_INETSIM_HTTP_ADDR = "172.20.0.2:80"

# Kiểm tra
Write-Host "Network Simulation: $env:OSSF_NETWORK_SIMULATION_ENABLED"
Write-Host "DNS Server: $env:OSSF_INETSIM_DNS_ADDR"
Write-Host "HTTP Server: $env:OSSF_INETSIM_HTTP_ADDR"
```

## Bước 6: Chạy Go Unit Tests

```powershell
# Di chuyển vào thư mục networksim
cd internal\networksim

# Chạy tests
go test -v

# Kết quả: Tất cả tests phải PASS
```

**Kết quả mong đợi:**
```
=== RUN   TestDefaultConfig
--- PASS: TestDefaultConfig (0.00s)
=== RUN   TestIsURLAlive
--- PASS: TestIsURLAlive (0.XXs)
=== RUN   TestShouldRedirectToINetSim
--- PASS: TestShouldRedirectToINetSim (0.00s)
PASS
ok      github.com/ossf/package-analysis/internal/networksim    X.XXXs
```

## Bước 7: Test với Sample Malicious Package

```powershell
# Quay lại dynamic-analysis
cd ..\..

# Di chuyển vào sample packages
cd sample_packages\malicious_network_package

# Cài đặt package
pip install -e .

# Chạy test
python test_network.py
```

**Kết quả mong đợi:** 
```
============================================================
Malicious Network Package - Connecting to dead URL
============================================================

[*] Target URL: http://malicious-c2-server.example.com/api/data
[*] Attempting connection...
[-] Connection failed: HTTPConnectionPool(...): Max retries exceeded...
============================================================
```

**Lưu ý:** Package sẽ thử kết nối tới dead URLs. Connection failed là behavior đúng khi chưa cấu hình DNS redirection. Khi network simulation được enable trong Go code, traffic sẽ được redirect đến INetSim.

## Bước 8: Xem Logs

```powershell
# Xem logs của INetSim
docker logs pack-a-mal-inetsim

# Xem logs của Service Simulation
docker logs pack-a-mal-sim-api

# Xem logs file (nếu cần)
Get-Content "..\..\service-simulation-module\shared\logs\inetsim\service.log" -Tail 20
```

## Bước 9: Dừng Services (khi hoàn thành)

```powershell
# Quay lại dynamic-analysis
cd D:\PROJECT\Project\pack-a-mal\dynamic-analysis

# Dừng tất cả services
docker-compose -f docker-compose.network-sim.yml down
```


