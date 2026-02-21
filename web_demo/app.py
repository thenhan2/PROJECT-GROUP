"""
Pack-a-mal Demo Web App (Standalone Flask)
Không dependencies ngoài Flask - chạy độc lập hoàn toàn
"""
import subprocess
import json
import os
import requests as req_lib
from pathlib import Path
from flask import Flask, render_template, jsonify, Response, stream_with_context

app = Flask(__name__)

BASE_DIR = Path(__file__).resolve().parent.parent
DYNAMIC_ANALYSIS_DIR = BASE_DIR / "dynamic-analysis"
SAMPLE_PKG_DIR = DYNAMIC_ANALYSIS_DIR / "sample_packages" / "malicious_network_package"
GO_TEST_DIR = DYNAMIC_ANALYSIS_DIR / "internal" / "networksim"

# Dùng venv Python để chạy demo scripts - venv có sitecustomize.py force UTF-8
VENV_PYTHON = Path(__file__).resolve().parent / "venv" / "Scripts" / "python.exe"


def run_cmd(command, cwd=None, env_extra=None, timeout=60):
    env = os.environ.copy()
    env["PYTHONIOENCODING"] = "utf-8"
    env["PYTHONUTF8"] = "1"
    if env_extra:
        env.update(env_extra)
    try:
        result = subprocess.run(
            command, shell=True, cwd=cwd,
            capture_output=True, text=True,
            encoding="utf-8", errors="replace",
            timeout=timeout, env=env
        )
        return {"ok": result.returncode == 0, "out": result.stdout, "err": result.stderr}
    except subprocess.TimeoutExpired:
        return {"ok": False, "out": "", "err": "Timeout sau 60 giây"}
    except Exception as e:
        return {"ok": False, "out": "", "err": str(e)}


def stream_cmd(command, cwd=None, env_extra=None):
    """Generator: stream output từng dòng qua SSE - đọc raw bytes để tránh encoding issues"""
    env = os.environ.copy()
    env["PYTHONIOENCODING"] = "utf-8"
    env["PYTHONUTF8"] = "1"
    if env_extra:
        env.update(env_extra)
    try:
        proc = subprocess.Popen(
            command, shell=True, cwd=cwd,
            stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
            env=env, bufsize=0  # raw bytes, không qua text mode
        )
        buf = b""
        while True:
            chunk = proc.stdout.read(1)
            if not chunk and proc.poll() is not None:
                break
            if chunk == b"\n" or (not chunk and buf):
                line = buf.decode("utf-8", errors="replace").rstrip("\r")
                buf = b""
                if line or chunk == b"\n":
                    yield f"data: {json.dumps(line)}\n\n"
            elif chunk:
                buf += chunk
        if buf:
            yield f"data: {json.dumps(buf.decode('utf-8', errors='replace'))}\n\n"
        proc.wait()
        yield f"data: {json.dumps('__DONE__:' + str(proc.returncode))}\n\n"
    except Exception as e:
        yield f"data: {json.dumps('__ERROR__:' + str(e))}\n\n"


def make_sse_response(gen_func):
    """Tạo SSE Response với đầy đủ headers để tắt buffering"""
    resp = Response(stream_with_context(gen_func()), mimetype="text/event-stream")
    resp.headers["Cache-Control"] = "no-cache"
    resp.headers["X-Accel-Buffering"] = "no"
    resp.headers["Connection"] = "keep-alive"
    return resp


# ─── Routes ────────────────────────────────────────────────────────────────────

@app.route("/")
def index():
    return render_template("index.html")


# ── Docker ──────────────────────────────────────────────────────────────────────

@app.route("/api/docker/<action>")
def docker(action):
    compose = DYNAMIC_ANALYSIS_DIR / "docker-compose.network-sim.yml"

    if action == "status":
        r = run_cmd('docker ps --filter "name=pack-a-mal" --format "{{.Names}}|{{.Status}}"')
        lines = [l for l in r["out"].strip().splitlines() if l]
        containers = [{"name": p.split("|")[0], "status": p.split("|")[1]} for p in lines if "|" in p]
        running = any("healthy" in c["status"].lower() or "up" in c["status"].lower() for c in containers)
        return jsonify({"ok": True, "running": running, "containers": containers})

    elif action == "start":
        def gen():
            yield f"data: {json.dumps('🚀 Đang khởi động Docker services...')}\n\n"
            yield from stream_cmd(
                f'docker-compose -f "{compose}" up -d --remove-orphans 2>&1',
                cwd=str(DYNAMIC_ANALYSIS_DIR),
                env_extra={"COMPOSE_PROGRESS": "plain", "NO_COLOR": "1"}
            )
            yield f"data: {json.dumps('─' * 40)}\n\n"
            yield f"data: {json.dumps('📋 Kiểm tra trạng thái containers...')}\n\n"
            yield from stream_cmd('docker ps --filter "name=pack-a-mal" --format "table {{.Names}}\\t{{.Status}}"')
        return make_sse_response(gen)

    elif action == "stop":
        def gen():
            yield f"data: {json.dumps('🛑 Đang dừng Docker services...')}\n\n"
            yield from stream_cmd(
                f'docker-compose -f "{compose}" down --remove-orphans 2>&1',
                cwd=str(DYNAMIC_ANALYSIS_DIR),
                env_extra={"COMPOSE_PROGRESS": "plain", "NO_COLOR": "1"}
            )
        return make_sse_response(gen)

    return jsonify({"ok": False, "msg": "Unknown action"})


# ── Test services ────────────────────────────────────────────────────────────────

@app.route("/api/test/<svc>")
def test_svc(svc):
    if svc == "http":
        r = run_cmd("curl.exe -s --max-time 5 http://localhost:8080", timeout=10)
        ok = r["ok"] and len(r["out"]) > 0
        return jsonify({"ok": ok, "label": "INetSim HTTP :8080",
                        "out": r["out"][:600] if ok else r["err"]})
    elif svc == "api":
        r = run_cmd("curl.exe -s --max-time 5 http://localhost:5000/status", timeout=10)
        ok = r["ok"] and len(r["out"]) > 0
        return jsonify({"ok": ok, "label": "Service Simulation API :5000",
                        "out": r["out"] if ok else r["err"]})
    return jsonify({"ok": False})


# ── Demo package ─────────────────────────────────────────────────────────────────

@app.route("/stream/demo/<mode>")
def demo_pkg(mode):
    if mode == "without":
        script = SAMPLE_PKG_DIR / "test_network.py"
        def gen():
            yield f"data: {json.dumps('🚫 Chạy KHÔNG có Network Simulation...')}\n\n"
            yield f"data: {json.dumps('─' * 50)}\n\n"
            yield from stream_cmd(f'"{VENV_PYTHON}" "{script}"', cwd=str(SAMPLE_PKG_DIR))
        return make_sse_response(gen)

    elif mode == "with":
        script = SAMPLE_PKG_DIR / "test_with_inetsim.py"
        def gen():
            yield f"data: {json.dumps('✅ Chạy CÓ Network Simulation (INetSim)...')}\n\n"
            yield f"data: {json.dumps('─' * 50)}\n\n"
            yield from stream_cmd(f'"{VENV_PYTHON}" "{script}"', cwd=str(SAMPLE_PKG_DIR))
        return make_sse_response(gen)

    elif mode == "full":
        script = SAMPLE_PKG_DIR / "test_full_mode.py"
        def gen():
            yield f"data: {json.dumps('🔴 Demo Full Isolation Mode...')}\n\n"
            yield f"data: {json.dumps('─' * 50)}\n\n"
            yield from stream_cmd(f'"{VENV_PYTHON}" "{script}"', cwd=str(SAMPLE_PKG_DIR))
        return make_sse_response(gen)

    return jsonify({"ok": False})


# ── Compare mode (before / after) ──────────────────────────────────────────────

@app.route("/api/compare/<mode>")
def compare_mode(mode):
    """Trả JSON {before, after} để UI hiển thị response so sánh"""
    if mode == "half":
        url = "http://malicious-c2-server.example.com/api/data"
    elif mode == "full":
        url = "http://example.com"
    else:
        return jsonify({"ok": False, "msg": "Unknown mode"})

    proxies = {"http": "http://localhost:8080", "https": "http://localhost:8080"}

    def do_req(use_proxy):
        try:
            r = req_lib.get(url, proxies=proxies if use_proxy else None,
                            timeout=6, allow_redirects=True)
            return {
                "ok": True,
                "status": r.status_code,
                "server": r.headers.get("Server", ""),
                "content_type": r.headers.get("Content-Type", ""),
                "body": r.text[:300],
                "size": len(r.content)
            }
        except req_lib.exceptions.ConnectionError as e:
            msg = str(e)
            # rút gọn message dài
            if "Max retries" in msg:
                msg = "Connection refused / Max retries exceeded"
            return {"ok": False, "error": msg[:180]}
        except req_lib.exceptions.Timeout:
            return {"ok": False, "error": "Request timed out (6s)"}
        except Exception as e:
            return {"ok": False, "error": str(e)[:180]}

    return jsonify({"ok": True, "mode": mode, "url": url,
                    "before": do_req(False), "after": do_req(True)})


# ── Go unit tests ─────────────────────────────────────────────────────────────────

@app.route("/stream/gotest")
def go_test():
    env = {
        "OSSF_NETWORK_SIMULATION_ENABLED": "true",
        "OSSF_INETSIM_DNS_ADDR": "172.20.0.2:53",
        "OSSF_INETSIM_HTTP_ADDR": "172.20.0.2:80",
    }
    def gen():
        yield f"data: {json.dumps('🧪 Đang chạy Go Unit Tests...')}\n\n"
        yield f"data: {json.dumps('─' * 50)}\n\n"
        yield from stream_cmd("go test -v ./...", cwd=str(GO_TEST_DIR), env_extra=env)
    return make_sse_response(gen)


# ── Network mode info ─────────────────────────────────────────────────────────────

@app.route("/api/mode/<name>")
def mode_info(name):
    modes = {
        "full": {
            "title": "Full Mode",
            "icon": "🔴",
            "desc": "Toàn bộ traffic bị chặn và redirect tới INetSim. Không có kết nối internet thật – môi trường cách ly hoàn toàn.",
            "dns": "INetSim DNS  172.20.0.2:53",
            "http": "INetSim HTTP  172.20.0.2:80",
            "safety": "Maximum",
            "color": "danger",
            "usecase": "Phân tích malware chưa biết nguồn gốc, môi trường sandbox hoàn toàn cách ly.",
            "flow": [
                "Package gửi request tới domain bất kỳ",
                "DNS bị chặn – toàn bộ resolve qua INetSim DNS",
                "INetSim trả về IP giả 127.0.0.1 cho mọi domain",
                "HTTP request bị redirect tới INetSim HTTP server",
                "INetSim trả về response giả lập, ghi lại toàn bộ hành vi"
            ]
        },
        "half": {
            "title": "Half Mode",
            "icon": "🟠",
            "desc": "Chặn và giả lập các URL đã chết (dead URLs). Các domain còn alive được kết nối bình thường.",
            "dns": "Conditional: dead → INetSim DNS / alive → System DNS",
            "http": "Dead URL → INetSim HTTP / Alive URL → Direct",
            "safety": "Medium",
            "color": "warning",
            "usecase": "Phát hiện malware dùng C2 server đã dead – redirect để giả lập response thay vì để fail.",
            "flow": [
                "Package gửi request tới một URL",
                "Hệ thống kiểm tra URL còn alive không",
                "✅ URL alive → cho qua kết nối internet trực tiếp",
                "❌ URL dead → redirect DNS & HTTP tới INetSim",
                "INetSim giả lập response, ghi lại hành vi của dead URL"
            ]
        }
    }
    data = modes.get(name)
    if not data:
        return jsonify({"ok": False})
    return jsonify({"ok": True, "mode": data})


# ── Package info ──────────────────────────────────────────────────────────────────

@app.route("/api/package-info")
def package_info():
    return jsonify({
        "ok": True,
        "name": "malicious-network-package",
        "version": "0.1.0",
        "description": "Package mẫu giả lập hành vi malware kết nối C2 server",
        "urls": [
            "http://malicious-c2-server.example.com/api/data",
            "http://evil-domain.net/payload",
            "http://dead-c2-server.com/beacon"
        ],
        "files": [
            {"name": "test_network.py", "purpose": "Test KHÔNG có simulation → ❌ fail"},
            {"name": "test_with_inetsim.py", "purpose": "Test CÓ INetSim → ✅ success"}
        ]
    })


if __name__ == "__main__":
    print("=" * 60)
    print("🚀  Pack-a-mal Demo Dashboard")
    print("    http://127.0.0.1:5500")
    print("=" * 60)
    app.run(host="127.0.0.1", port=5500, debug=False)
