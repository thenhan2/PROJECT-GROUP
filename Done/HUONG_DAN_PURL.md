# 🎯 HƯỚNG DẪN SỬ DỤNG PACKAGE URL (pURL)

## 📋 Package URL (pURL) là gì?

**Package URL (pURL)** là một định dạng chuẩn quốc tế để xác định các package phần mềm từ mọi hệ sinh thái bằng một chuỗi định danh duy nhất.

**Format chuẩn:**
```
pkg:ecosystem/namespace/name@version
```

**Ví dụ cụ thể:**
- PyPI: `pkg:pypi/requests@2.31.0`
- npm: `pkg:npm/@babel/core@7.22.0`
- Maven: `pkg:maven/org.springframework/spring-core@6.0.0`

---

## 🎯 Lợi ích khi sử dụng pURL

✅ **Định danh chuẩn quốc tế**: Một format cho tất cả ecosystems  
✅ **Dễ chia sẻ**: Copy-paste một chuỗi thay vì nhiều tham số  
✅ **SBOM tương thích**: Tích hợp với Software Bill of Materials  
✅ **Automation-friendly**: Dễ dàng cho CI/CD pipelines  

---

## 🔍 LỰA CHỌN CÁCH SỬ DỤNG

> **💡 CẢ 3 CÁCH ĐỀU CHO KẾT QUẢ GIỐNG NHAU!**  
> Phần 1, 2, 3 chỉ khác nhau về **cách thức** chứ **mục đích đều là phân tích package bằng pURL**.

### Bạn nên dùng cách nào?

| Tình huống | Nên dùng | Lý do |
|------------|----------|-------|
| 🪟 **Đang dùng Windows** | 👉 **PHẦN 2 (Web API)** | Không cần WSL/Ubuntu, setup đơn giản |
| 🐧 **Có Ubuntu/WSL** | 👉 **PHẦN 1 hoặc 2** | Cả 2 đều chạy được, tùy thích |
| 🧪 **Test nhanh** | 👉 **PHẦN 2 (Web API)** | Dễ nhất, có script Python sẵn |
| 🤖 **CI/CD Pipeline** | 👉 **PHẦN 1 (Command-line)** | Tích hợp tốt hơn cho automation |
| 🌐 **Tích hợp Web App** | 👉 **PHẦN 2 (Web API)** | REST API, dễ gọi từ frontend |
| 📦 **Xử lý batch** | 👉 **PHẦN 1 hoặc 2** | Cả 2 đều hỗ trợ |

### Tóm tắt:

**PHẦN 1 - Command-line Tool:**
- ✅ Nhanh, chạy local
- ❌ Chỉ chạy trên Ubuntu/WSL
- ❌ Cần build Go binary
- 🎯 Phù hợp: CI/CD, automation, batch processing

**PHẦN 2 - Web API:**
- ✅ Chạy mọi nơi (Windows/Linux/Mac)
- ✅ Setup đơn giản (chỉ cần Python)
- ✅ Có sẵn REST API
- ⚠️ Cần Django server chạy
- 🎯 Phù hợp: Testing, web integration, team collaboration

**PHẦN 3 - Ecosystems:**
- 📚 Danh sách tất cả ecosystems được hỗ trợ
- 📖 Tham khảo format pURL cho từng ecosystem

---

## ⚡ QUICK START (Chọn 1 trong 2 cách)

### 🪟 Trên Windows → Dùng Web API (5 phút):

```powershell
# 1. Di chuyển vào thư mục web
cd web\packamal

# 2. Khởi động Django server
python manage.py runserver

# 3. Gửi request (Terminal mới)
curl -X POST "http://localhost:8000/api/analyze/" -H "Content-Type: application/json" -d "{\"purl\": \"pkg:pypi/requests@2.31.0\"}"
```

### 🐧 Trên Ubuntu/WSL → Dùng Command-line (5 phút):

```bash
# 1. Build tool
cd dynamic-analysis
make build

# 2. Phân tích
./analyze -purl "pkg:pypi/requests@2.31.0"
```

---

## 🚀 PHẦN 1: SỬ DỤNG COMMAND-LINE TOOL

> **Yêu cầu:** Ubuntu/WSL (không chạy được trên Windows)  
> **Khi nào dùng:** CI/CD, automation, batch processing  
> **Nếu dùng Windows:** → Xem **PHẦN 2 (Web API)** sẽ dễ hơn!

### So sánh: Command-Line vs Web API

| Tiêu chí | Command-Line Tool | Web API |
|----------|-------------------|---------|
| **Hệ điều hành** | ❌ Chỉ Ubuntu/WSL | ✅ Windows/Linux/Mac |
| **Cài đặt** | ❌ Cần build Go binary | ✅ Chỉ cần pip install |
| **Tốc độ** | ✅ Nhanh (chạy local) | ⚠️ Tùy server |
| **Dễ sử dụng** | ⚠️ Command-line | ✅ REST API |
| **Phù hợp cho** | CI/CD, automation | Web apps, testing |

### Bước 1.1: Build Analyze Tool (Trên Ubuntu/WSL)

```bash
# Di chuyển vào thư mục dynamic-analysis
cd dynamic-analysis/

# Build binary
make build

# Hoặc build thủ công
cd cmd/analyze
go build -o analyze .
```

**⚠️ Lưu ý:** Tool này cần Unix syscalls nên chỉ chạy được trên Linux/WSL, không chạy trên Windows.

### Bước 1.2: Phân tích package với pURL

#### Cách 1: Dùng pURL (Khuyến nghị ✅)

```bash
# Phân tích Python package
./analyze -purl "pkg:pypi/requests@2.31.0"

# Phân tích npm package (có scope)
./analyze -purl "pkg:npm/@babel/core@7.22.0"

# Phân tích Maven package (có namespace)
./analyze -purl "pkg:maven/org.springframework/spring-core@6.0.0"

# Phân tích phiên bản mới nhất (không cần version)
./analyze -purl "pkg:pypi/flask"
```

#### Cách 2: Dùng tham số truyền thống (Vẫn hoạt động)

```bash
# Cách cũ vẫn được hỗ trợ
./analyze -ecosystem pypi -package requests -version 2.31.0
```


> **Yêu cầu:** Python (bất kỳ HĐH nào)  
> **Khi nào dùng:** Testing, web integration, Windows users  
> **Ưu điểm:**  
> - ✅ Không cần build Go binary  
> - ✅ Chạy trên Windows/Linux/Mac  
> - ✅ Setup đơn giản với pip
> - ✅ Không cần build như command-line tool  
> - ✅ Chạy được trên Windows (không cần WSL/Ubuntu)  
> - ✅ Chỉ cần khởi động Django server là dùng được ngay  

### Bước 2.1: Khởi động Web Server

#### Trên Windows (Khuyến nghị):

```powershell
# Di chuyển vào thư mục web
cd web\packamal\

# Tạo virtual environment (lần đầu tiên)
python -m venv venv

# Kích hoạt virtual environment
.\venv\Scripts\activate

# Cài đặt dependencies
pip install -r requirements.txt

# Chạy migrations
python manage.py migrate

# Khởi động server
python manage.py runserver 0.0.0.0:8000
```

#### Trên Ubuntu/Linux:

```bash
# Di chuyển vào thư mục web
cd web/packamal/

# Tạo virtual environment (lần đầu tiên)
python3 -m venv venv

# Kích hoạt virtual environment
source venv/bin/activate

# Cài đặt dependencies
pip install -r requirements.txt

# Chạy migrations
python manage.py migrate

# Khởi động server
python manage.py runserver 0.0.0.0:8000
```

**✅ Server chạy tại:** http://localhost:8000

### Bước 2.2: Gửi Request với pURL

#### Cách 1: Sử dụng curl (Linux/Mac/Windows PowerShell):

```bash
# Phân tích bằng pURL (Khuyến nghị ✅)
curl -X POST "http://localhost:8000/api/analyze/" \
  -H "Content-Type: application/json" \
  -d '{"purl": "pkg:pypi/requests@2.31.0"}'

# Response sẽ trả về task_id và status_url để theo dõi tiến trình
```

**Response mẫu:**
```json
{
  "task_id": 123,
  "status": "queued",
  "status_url": "http://localhost:8000/api/tasks/123/status/",
  "result_url": "http://localhost:8000/reports/pypi/requests/2.31.0.json",
  "message": "Analysis queued successfully"
}
```

#### Cách 2: Sử dụng Python (Dễ nhất cho Windows):

```python
import requests
import time
import json

# Gửi request phân tích
response = requests.post(
    "http://localhost:8000/api/analyze/",
    json={"purl": "pkg:pypi/flask@3.0.0"}
)

result = response.json()
print("Task created:", json.dumps(result, indent=2))

# Lấy task_id và status_url để theo dõi
task_id = result['task_id']
status_url = result['status_url']

# Polling để kiểm tra khi nào hoàn thành
print("\nWaiting for analysis to complete...")
while True:
    status_response = requests.get(status_url)
    status = status_response.json()
    
    print(f"Status: {status['status']}")
    
    if status['status'] == 'completed':
        print("\n✅ Analysis completed!")
        print(f"Download report: {result['result_url']}")
        break
    elif status['status'] == 'failed':
        print("\n❌ Analysis failed!")
        print(f"Error: {status.get('error_message')}")
        break
    
    time.sleep(5)  # Đợi 5 giây trước khi check lại
```

#### Cách 3: Sử dụng Postman (GUI):

1. Mở Postman
2. Method: **POST**
3. URL: `http://localhost:8000/api/analyze/`
4. Headers:
   - `Content-Type`: `application/json`
5. Body (raw JSON):
```json
{
  "purl": "pkg:pypi/requests@2.31.0"
}
```
6. Click **Send**

#### Cách 4: Dùng tham số truyền thống (vẫn được hỗ trợ):

```bash
curl -X POST "http://localhost:8000/api/analyze/" \
  -H "Content-Type: application/json" \
  -d '{
    "ecosystem": "pypi",
    "package_name": "requests",
    "package_version": "2.31.0"
  }'
```

---

## 📚 PHẦN 3: ECOSYSTEMS ĐƯỢC HỖ TRỢ

| Ecosystem | pURL Format | Ví dụ |
|-----------|-------------|--------|
| **PyPI** (Python) | `pkg:pypi/name@version` | `pkg:pypi/django@5.0.0` |
| **npm** (Node.js) | `pkg:npm/name@version` | `pkg:npm/express@4.18.0` |
| **npm scoped** | `pkg:npm/@scope/name@version` | `pkg:npm/@babel/core@7.22.0` |
| **Maven** (Java) | `pkg:maven/group/artifact@version` | `pkg:maven/org.springframework/spring-core@6.0.0` |
| **RubyGems** (Ruby) | `pkg:gem/name@version` | `pkg:gem/rails@7.0.0` |
| **Packagist** (PHP) | `pkg:composer/vendor/package@version` | `pkg:composer/symfony/console@6.0.0` |
| **Crates.io** (Rust) | `pkg:cargo/name@version` | `pkg:cargo/serde@1.0.0` |

---

## 🧪 PHẦN 4: CHẠY TEST SCRIPTS

### Test Web API trên Windows (Dễ nhất - Không cần Docker/Ubuntu!)

```powershell
# Terminal 1: Khởi động Django server
cd web\packamal
python manage.py runserver

# Terminal 2: Chạy test script
cd web\packamal
python test_purl_web_api.py "pkg:pypi/requests@2.31.0"

# Hoặc chạy interactive mode
python test_purl_web_api.py
```

**Script sẽ tự động:**
- ✅ Gửi request phân tích với pURL
- ✅ Theo dõi tiến trình (polling)
- ✅ Hiển thị kết quả khi hoàn thành
- ✅ Download link cho report JSON

### Test Command-Line Tool trên Ubuntu/WSL

```bash
# Di chuyển vào thư mục examples/purl
cd dynamic-analysis/examples/purl/

# Cho phép execute permission
chmod +x test_purl_ubuntu.sh

# Chạy test suite đầy đủ
./test_purl_ubuntu.sh
```

**Kết quả mong đợi:** 6/6 tests passed ✅

### Demo với Python Script (Không cần build)

```bash
# Chạy Python demo
python test_purl_parsing.py "pkg:pypi/requests@2.31.0"
```

### Batch Processing

```bash
# Phân tích nhiều packages từ file
chmod +x analyze-with-purl.sh
./analyze-with-purl.sh purl-examples.txt
```

---

## 📖 PHẦN 5: VÍ DỤ THỰC TÊ

### Ví dụ 1: Phân tích package Python nghi ngờ malware

```bash
./analyze -purl "pkg:pypi/malicious-package@1.0.0"
```

### Ví dụ 2: Kiểm tra tất cả dependencies trong SBOM

```bash
# Tạo file sbom-purls.txt với nội dung:
pkg:pypi/requests@2.31.0
pkg:npm/express@4.18.0
pkg:maven/org.springframework/spring-core@6.0.0

# Phân tích tất cả
while read purl; do
  ./analyze -purl "$purl"
done < sbom-purls.txt
```

### Ví dụ 3: CI/CD Integration

```yaml
# .github/workflows/security-scan.yml
name: Security Scan

on: [push]

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - name: Analyze package
        run: |
          ./analyze -purl "pkg:pypi/${{ github.event.repository.name }}@${{ github.ref_name }}"
```

---

## 🔍 PHẦN 6: TROUBLESHOOTING

### Lỗi: "Invalid pURL format"

**Nguyên nhân:** Sai format pURL

**Giải pháp:**
```bash
# Sai ❌
./analyze -purl "pypi/requests@2.31.0"

# Đúng ✅
./analyze -purl "pkg:pypi/requests@2.31.0"
```

### Lỗi: Build fail trên Windows

**Nguyên nhân:** Code sử dụng Unix syscalls không có trên Windows

**Giải pháp:** 
- Sử dụng WSL (Windows Subsystem for Linux)
- Hoặc dùng Ubuntu VM/Docker container

### Lỗi: "Package not found"

**Nguyên nhân:** Package không tồn tại hoặc version sai

**Giải pháp:**
```bash
# Kiểm tra package tồn tại trước
# PyPI
curl https://pypi.org/pypi/requests/2.31.0/json

# npm
curl https://registry.npmjs.org/express/4.18.0
```

---

## 📁 PHẦN 7: TÀI LIỆU THAM KHẢO

### Tài liệu trong project:

1. **[dynamic-analysis/docs/PURL_GUIDE.md](dynamic-analysis/docs/PURL_GUIDE.md)**  
   → Hướng dẫn chi tiết về implementation

2. **[dynamic-analysis/examples/purl/README.md](dynamic-analysis/examples/purl/README.md)**  
   → Quick start guide cho pURL examples

3. **[dynamic-analysis/examples/purl/PURL_EXAMPLES.md](dynamic-analysis/examples/purl/PURL_EXAMPLES.md)**  
   → Ví dụ nâng cao và use cases

### Tài liệu external:

- **[pURL Specification](https://github.com/package-url/purl-spec)** - Đặc tả chính thức
- **[packageurl-go](https://github.com/package-url/packageurl-go)** - Library Go chúng ta sử dụng

---

## 💡 TIPS & BEST PRACTICES

### 1. Luôn dùng pURL khi có thể

```bash
# Thay vì
./analyze -ecosystem pypi -package requests -version 2.31.0

# Dùng
./analyze -purl "pkg:pypi/requests@2.31.0"
```

### 2. Batch processing nên dùng file list

```bash
# Tạo file purls.txt
echo "pkg:pypi/requests@2.31.0" >> purls.txt
echo "pkg:npm/express@4.18.0" >> purls.txt

# Chạy batch
while read purl; do ./analyze -purl "$purl"; done < purls.txt
```

### 3. Log kết quả ra file

```bash
./analyze -purl "pkg:pypi/requests@2.31.0" > results.json 2>&1
```

### 4. Version mới nhất

```bash
# Bỏ @version để phân tích version mới nhất
./analyze -purl "pkg:pypi/django"
```

---

## 🎓 PHẦN 8: BÀI TẬP THỰC HÀNH

### Bài 1: Phân tích package Python

```bash
# Yêu cầu: Phân tích package 'flask' version 3.0.0
# TODO: Viết command pURL của bạn ở đây
```

<details>
<summary>Đáp án</summary>

```bash
./analyze -purl "pkg:pypi/flask@3.0.0"
```
</details>

### Bài 2: Phân tích scoped npm package

```bash
# Yêu cầu: Phân tích package '@vue/cli' version 5.0.0
# TODO: Viết command pURL của bạn ở đây
```

<details>
<summary>Đáp án</summary>

```bash
./analyze -purl "pkg:npm/@vue/cli@5.0.0"
```
</details>

### Bài 3: Phân tích Maven package với namespace

```bash
# Yêu cầu: Phân tích package 'com.google.guava:guava' version 32.1.0
# TODO: Viết command pURL của bạn ở đây
```

<details>
<summary>Đáp án</summary>

```bash
./analyze -purl "pkg:maven/com.google.guava/guava@32.1.0"
```
</details>

---

## ❓ FAQ (Câu hỏi thường gặp)

### Q1: Phần 1, 2, 3 có khác nhau về kết quả không?

**A:** ❌ **KHÔNG!** Cả 3 cách đều cho **kết quả phân tích giống hệt nhau**. Chỉ khác nhau về cách thức:
- **PHẦN 1 (Command-line):** Dùng lệnh terminal để phân tích
- **PHẦN 2 (Web API):** Gửi HTTP request để phân tích  
- **PHẦN 3:** Danh sách các ecosystems được hỗ trợ (tham khảo)

→ Chọn cách nào tùy vào môi trường và sở thích của bạn!

### Q2: Tôi dùng Windows, nên chọn phần nào?

**A:** 👉 **PHẦN 2 (Web API)** - Đơn giản nhất cho Windows!
- Không cần WSL/Ubuntu
- Chỉ cần Python và Django
- Setup trong 5 phút

### Q3: Tôi có Ubuntu, nên dùng phần nào?

**A:** 👉 **Cả 2 đều được!** Tùy mục đích:
- **Automation/CI/CD:** Dùng Phần 1 (Command-line)
- **Testing/Web app:** Dùng Phần 2 (Web API)

### Q4: pURL có thay thế hoàn toàn cách cũ không?

**A:** Không, cả hai cách đều được hỗ trợ. pURL là cách khuyến nghị nhưng tham số truyền thống (`-ecosystem`, `-package`, `-version` cho command-line hoặc `ecosystem`, `package_name`, `package_version` cho Web API) vẫn hoạt động.

### Q5: Tôi phải build cả 2 (command-line và web) không?

**A:** ❌ **KHÔNG CẦN!** Chọn 1 trong 2:
- **Chỉ dùng Web API:** Không cần build gì, chỉ cần `pip install -r requirements.txt`
- **Chỉ dùng Command-line:** Chỉ cần build Go binary trên Ubuntu/WSL

### Q6: Web API có cần Docker không?

**A:** ✅ **KHÔNG CẦN!** Web API chạy được ngay trên Windows với Python. Docker chỉ cần cho production deployment hoặc test với network simulation.

### Q7: pURL có case-sensitive không?

**A:** Có, ecosystem phải viết thường (`pypi` không phải `PyPI`), nhưng package name tùy thuộc vào ecosystem.

### Q8: Làm sao biết package tồn tại?

**A:** Kiểm tra trực tiếp trên registry:
- PyPI: https://pypi.org/project/package-name/
- npm: https://www.npmjs.com/package/package-name
- Maven: https://search.maven.org/

### Q9: pURL có hỗ trợ version ranges không?

**A:** Không, pURL chỉ hỗ trợ version cụ thể. Để dùng version mới nhất, bỏ qua phần `@version`.

---

## 👥 HỖ TRỢ

Nếu gặp vấn đề hoặc có câu hỏi:

1. Kiểm tra [Troubleshooting](#-phần-6-troubleshooting) section
2. Xem [PURL_GUIDE.md](dynamic-analysis/docs/PURL_GUIDE.md) để biết chi tiết
3. Chạy test scripts để verify setup: `./test_purl_ubuntu.sh`
4. Liên hệ team leader

---

## 📝 CHANGELOG

- **2026-02-08**: Tạo hướng dẫn pURL cho team
- Implementation hoàn thành với 6 ecosystems được hỗ trợ
- Test suite passed 6/6 trên Ubuntu/WSL

---

**Chúc các bạn sử dụng pURL thành công! 🚀**
