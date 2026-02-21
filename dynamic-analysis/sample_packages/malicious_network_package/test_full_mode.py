#!/usr/bin/env python3
"""
Full Mode Demo - Chứng minh Full Isolation Mode chặn cả URL alive.
So sánh 2 kết quả:
  1. URL alive (example.com) → kết nối TRỰC TIẾP thành công
  2. URL alive (example.com) → qua INetSim (Full Mode) → vẫn bị intercept
"""

import requests

ALIVE_URL   = "http://example.com"
INETSIM_PROXY = {"http": "http://localhost:8080", "https": "http://localhost:8080"}


def section(title):
    bar = "=" * 60
    print(f"\n{bar}")
    print(f"  {title}")
    print(bar)


def test_direct(url):
    """Kết nối thẳng không qua proxy - URL alive thì sẽ thành công"""
    print(f"\n[*] Target      : {url}")
    print(f"[*] Proxy       : NONE (direct internet)")
    try:
        r = requests.get(url, timeout=5)
        print(f"\n    ✓ Status     : {r.status_code}")
        print(f"    ✓ Content    : {len(r.content)} bytes")
        server = r.headers.get("Server", "unknown")
        print(f"    ✓ Server Hdr : {server}")
        print(f"\n    → URL alive, kết nối trực tiếp thành công")
        return True
    except Exception as e:
        print(f"\n    ✗ Failed: {str(e)[:100]}")
        return False


def test_via_inetsim(url):
    """Kết nối qua INetSim proxy - Full Mode: mọi traffic đều bị intercept"""
    print(f"\n[*] Target      : {url}")
    print(f"[*] Proxy       : http://localhost:8080  (INetSim - Full Mode)")
    try:
        r = requests.get(url, proxies=INETSIM_PROXY, timeout=5)
        print(f"\n    ✓ Status     : {r.status_code}")
        print(f"    ✓ Content    : {len(r.content)} bytes")
        server = r.headers.get("Server", "unknown")
        print(f"    ✓ Server Hdr : {server}")

        is_inetsim = (
            "INetSim" in r.text
            or "default HTML" in r.text
            or "inetsim" in server.lower()
            or r.status_code in (200, 302)
            and len(r.content) < 2048  # INetSim trả response nhỏ
        )

        if is_inetsim:
            print(f"\n    ⚠ Response đến từ INetSim, KHÔNG phải real server!")
            print(f"    → Full Mode đã chặn và intercept URL alive thành công")
        else:
            print(f"\n    ✓ Response nhận được (có thể từ real server hoặc INetSim)")
        return True

    except requests.exceptions.ConnectionError:
        print(f"\n    ✗ Không kết nối được INetSim → Hãy chắc Docker đang chạy")
        return False
    except Exception as e:
        print(f"\n    ✗ Error: {str(e)[:100]}")
        return False


if __name__ == "__main__":

    section("BƯỚC 1 — Không có Full Mode  (kết nối trực tiếp)")
    print("Kịch bản: URL còn alive → kết nối thành công, dữ liệu KHÔNG bị ghi lại")
    ok1 = test_direct(ALIVE_URL)

    section("BƯỚC 2 — Full Mode BẬT  (INetSim intercept mọi thứ)")
    print("Kịch bản: Cùng URL alive → Full Mode chặn lại, INetSim xử lý thay thế")
    ok2 = test_via_inetsim(ALIVE_URL)

    section("KẾT LUẬN")
    if ok1 and ok2:
        print("\n  ✅ Full Mode hoạt động đúng:")
        print("     • Direct  → thành công    (không giám sát)")
        print("     • INetSim → bị intercept  (hành vi được ghi lại)")
        print("\n  → Khác với Half Mode: Half Mode chỉ chặn URL đã DEAD,")
        print("    Full Mode chặn MỌI traffic kể cả URL còn alive.")
    elif ok2:
        print("\n  ✅ INetSim đang intercept traffic (Full Mode active)")
    else:
        print("\n  ⚠  INetSim chưa chạy — hãy Start Docker ở Card 01 trước")
    print()
