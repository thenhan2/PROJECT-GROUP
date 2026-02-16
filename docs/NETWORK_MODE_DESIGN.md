# Network Mode Design - Full Mode vs Half Mode

## 📋 Tổng Quan

Thiết kế hai chế độ phân tích mạng cho Pack-A-Mal để kiểm soát linh hoạt việc xử lý traffic từ malware samples:

- **Full Mode (Isolated Mode)**: Không có giao tiếp nào được phép rời khỏi môi trường phân tích. Tất cả các giao thức được xử lý bởi service simulation.
- **Half Mode (Transparent Proxy Mode)**: Traffic được chặn, phân tích, có thể sửa đổi và quyết định forward đến đích thật hoặc block.

---

## 🎯 Mục Tiêu

### Full Mode
- **Mục đích**: Phân tích malware mà hoàn toàn KHÔNG có rủi ro kết nối internet thực
- **Use case**: 
  - Phân tích malware nguy hiểm, chưa rõ hành vi
  - Môi trường lab không được phép có kết nối ra ngoài
  - Phân tích nhanh để xem malware cố gắng kết nối đến đâu

### Half Mode  
- **Mục đích**: Phân tích malware với khả năng tương tác có kiểm soát với internet thực
- **Use case**:
  - Nghiên cứu C2 (Command & Control) infrastructure
  - Thu thập payload từ server thực
  - Phân tích phishing campaigns
  - Theo dõi exfiltration data

---

## 🏗️ Kiến Trúc Hiện Tại

### Thành Phần Đã Có

```
┌─────────────────────────────────────────────────────────┐
│  1. INetSim (172.20.0.2)                                │
│     - DNS, SMTP, FTP, HTTP/HTTPS simulation             │
│     - Hoàn toàn isolated                                │
│                                                         │
│  2. FakeNet-NG (172.20.0.3)                            │
│     - HTTP/HTTPS interception                           │
│     - Traffic analysis                                  │
│     - Response injection                                │
│                                                         │
│  3. NetworkSimulator (Go - networksim.go)              │
│     - Check URL alive                                   │
│     - Redirect to INetSim if URL dead                  │
│     - Binary choice: Real or Simulated                 │
│                                                         │
│  4. HTTP Simulation Service                            │
│     - Request analysis & classification                 │
│     - Safe executable handling                          │
│     - Response generation                               │
└─────────────────────────────────────────────────────────┘
```

### Hạn Chế Hiện Tại

❌ **Không có chế độ Hybrid**: Chỉ có "real" hoặc "simulated"
❌ **Không thể sửa đổi traffic**: Không thể modify request/response
❌ **Không thể selective forwarding**: Không quyết định được request nào forward, request nào block
❌ **Thiếu traffic monitoring**: Không có layer để observe và log tất cả traffic

---

## 🚀 Thiết Kế Mới: Network Mode Controller

### Architecture Overview

```
┌────────────────────────────────────────────────────────────────┐
│                    MALWARE SAMPLE                              │
└────────────────────────┬───────────────────────────────────────┘
                         │ All Network Traffic
                         ↓
┌────────────────────────────────────────────────────────────────┐
│           NETWORK MODE CONTROLLER (New Component)              │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │  Traffic Interceptor                                     │ │
│  │  - Capture all egress traffic                           │ │
│  │  - Deep packet inspection                               │ │
│  │  - Protocol identification                              │ │
│  └────────────────────┬─────────────────────────────────────┘ │
│                       │                                        │
│                       ↓                                        │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │  Mode Router                                             │ │
│  │  ┌──────────────┐         ┌──────────────┐             │ │
│  │  │  FULL MODE   │         │  HALF MODE   │             │ │
│  │  │  (Isolated)  │         │  (Proxy)     │             │ │
│  │  └──────────────┘         └──────────────┘             │ │
│  └────────────────────┬─────────────────────┬──────────────┘ │
│                       │                     │                 │
└───────────────────────┼─────────────────────┼─────────────────┘
                        │                     │
           ┌────────────┘                     └──────────────┐
           ↓                                                 ↓
┌─────────────────────┐                    ┌──────────────────────────┐
│   FULL MODE PATH    │                    │    HALF MODE PATH        │
│                     │                    │                          │
│  ┌───────────────┐  │                    │  ┌────────────────────┐  │
│  │  INetSim      │  │                    │  │  Traffic Analyzer  │  │
│  │  DNS/SMTP/FTP │  │                    │  │  - Deep inspection │  │
│  └───────────────┘  │                    │  │  - Risk assessment│  │
│                     │                    │  └─────────┬──────────┘  │
│  ┌───────────────┐  │                    │            │             │
│  │  FakeNet-NG   │  │                    │            ↓             │
│  │  HTTP/HTTPS   │  │                    │  ┌────────────────────┐  │
│  └───────────────┘  │                    │  │  Decision Engine   │  │
│                     │                    │  │  - Policy check    │  │
│  ┌───────────────┐  │                    │  │  - Whitelist/Black │  │
│  │  Service Sim  │  │                    │  └─────────┬──────────┘  │
│  │  Custom APIs  │  │                    │            │             │
│  └───────────────┘  │                    │     ┌──────┴──────┐      │
│                     │                    │     ↓             ↓      │
│  ALL → Simulation   │                    │  Forward      Block      │
│  NO external traffic│                    │    ↓           ↓         │
└─────────────────────┘                    │  ┌──────┐   ┌──────┐    │
                                           │  │ Real │   │ Fake │    │
                                           │  │Internet  │Response│   │
                                           │  └──────┘   └──────┘    │
                                           │     ↓          ↓         │
                                           │  ┌──────────────────┐   │
                                           │  │ Modifier         │   │
                                           │  │ - Strip headers  │   │
                                           │  │ - Inject content │   │
                                           │  │ - Log everything │   │
                                           │  └──────────────────┘   │
                                           └──────────────────────────┘
                                                      ↓
                                           ┌────────────────────────┐
                                           │  Logging & Analytics   │
                                           │  - Full PCAP           │
                                           │  - HTTP logs           │
                                           │  - DNS queries         │
                                           │  - Decisions made      │
                                           └────────────────────────┘
```

---

## 📊 Chi Tiết Thiết Kế

### 1. Full Mode (Isolated Mode)

#### Đặc điểm
```yaml
Mode: FULL
Isolation Level: COMPLETE
Internet Access: DENIED
Traffic Handling: ALL_SIMULATED
Security Level: MAXIMUM
```

#### Flow Chart
```
Malware Request
      ↓
[Check Mode = FULL?] → YES
      ↓
[Protocol Detection]
      ↓
  ┌───┴─────┬──────────┬─────────┬──────────┐
  ↓         ↓          ↓         ↓          ↓
 DNS      HTTP      HTTPS      SMTP       FTP
  ↓         ↓          ↓         ↓          ↓
INetSim  FakeNet-NG  Service  INetSim   INetSim
         /ServiceSim   Sim
      ↓
[Generate Fake Response]
      ↓
[Log Activity]
      ↓
Return to Malware
```

#### Configuration Example
```yaml
network_mode:
  mode: "full"
  isolation:
    block_all_external: true
    allow_dns_resolution: false  # Use INetSim DNS only
    allow_http_external: false
    allow_https_external: false
  
  services:
    dns:
      handler: "inetsim"
      address: "172.20.0.2:53"
      default_response: "127.0.0.1"
    
    http:
      handler: "fakenet-ng"
      address: "172.20.0.3:80"
      response_mode: "adaptive"  # smart, static, honeypot
    
    https:
      handler: "service-simulation"
      address: "172.20.0.4:443"
      ssl_interception: true
    
    smtp:
      handler: "inetsim"
      address: "172.20.0.2:25"
    
    ftp:
      handler: "inetsim"
      address: "172.20.0.2:21"
  
  logging:
    level: "verbose"
    capture_pcap: true
    log_all_requests: true
    log_responses: true
```

#### Implementation Points
- ✅ **Đã có**: INetSim, FakeNet-NG, Service Simulation
- 🔄 **Cần mở rộng**: 
  - Central controller để route traffic
  - Configuration system cho mode selection
  - Enhanced logging với correlation

---

### 2. Half Mode (Transparent Proxy Mode)

#### Đặc điểm
```yaml
Mode: HALF
Isolation Level: CONTROLLED
Internet Access: SELECTIVE
Traffic Handling: INSPECT_AND_DECIDE
Security Level: MEDIUM-HIGH
```

#### Flow Chart
```
Malware Request
      ↓
[Check Mode = HALF?] → YES
      ↓
[Deep Packet Inspection]
      ↓
[Extract Request Details]
  - Destination IP/Domain
  - Protocol
  - Headers
  - Payload
      ↓
[Apply Decision Rules]
      ↓
  ┌───┴────────┬─────────────┬──────────────┐
  ↓            ↓             ↓              ↓
FORWARD     MODIFY       BLOCK         SIMULATE
  ↓            ↓             ↓              ↓
[Send to    [Modify       [Drop         [Use Full
 Real       Request/      Request]       Mode Path]
 Internet]  Response]        ↓              ↓
  ↓            ↓          [Return         [Fake
  ↓         [Forward      Error]         Response]
  ↓          Modified]      
  ↓            ↓
[Record Response]
  ↓
[Apply Response Filters]
  - Strip malicious payloads
  - Sanitize data
  - Inject markers
      ↓
[Log Everything]
      ↓
Return to Malware
```

#### Decision Engine Rules

```go
type DecisionRule struct {
    Name        string
    Priority    int
    Condition   Condition
    Action      Action
    Modifier    *Modifier
}

// Example Rules
var HalfModeRules = []DecisionRule{
    {
        Name:     "Block Known C2 Servers",
        Priority: 100,
        Condition: Condition{
            Type: "domain_blacklist",
            Values: []string{
                "*.malware-c2.com",
                "evil.example.com",
                "192.168.1.100",
            },
        },
        Action: "BLOCK",
    },
    {
        Name:     "Allow Legitimate CDNs",
        Priority: 90,
        Condition: Condition{
            Type: "domain_whitelist",
            Values: []string{
                "*.cloudflare.com",
                "*.akamai.com",
                "*.fastly.com",
            },
        },
        Action: "FORWARD",
    },
    {
        Name:     "Intercept Executable Downloads",
        Priority: 80,
        Condition: Condition{
            Type: "content_type",
            Values: []string{
                "application/x-msdownload",
                "application/executable",
            },
        },
        Action: "MODIFY",
        Modifier: &Modifier{
            Type: "sandbox_executable",
            SaveOriginal: true,
            ReplaceWith: "fake_payload",
        },
    },
    {
        Name:     "Monitor Data Exfiltration",
        Priority: 70,
        Condition: Condition{
            Type: "method_and_size",
            Method: "POST",
            MinSize: 1024 * 1024, // > 1MB
        },
        Action: "INSPECT_AND_FORWARD",
        Modifier: &Modifier{
            Type: "log_full_content",
            StripSensitiveData: true,
        },
    },
    {
        Name:     "Block Unknown Protocols",
        Priority: 50,
        Condition: Condition{
            Type: "protocol",
            Values: []string{"IRC", "Telnet", "SMB"},
        },
        Action: "BLOCK",
    },
    {
        Name:     "Default - Simulate",
        Priority: 1,
        Condition: Condition{
            Type: "default",
        },
        Action: "SIMULATE",
    },
}
```

#### Configuration Example
```yaml
network_mode:
  mode: "half"
  
  proxy:
    transparent: true
    listen_address: "0.0.0.0"
    dns_interception: true
    ssl_interception: true
    ssl_cert_path: "/certs/proxy-ca.crt"
  
  decision_engine:
    default_action: "simulate"  # forward, block, simulate
    
    rules:
      - name: "block_known_c2"
        priority: 100
        condition:
          type: "domain_blacklist"
          source: "file:///config/c2-blacklist.txt"
        action: "block"
        log_level: "alert"
      
      - name: "allow_cdns"
        priority: 90
        condition:
          type: "domain_whitelist"
          domains:
            - "*.cloudflare.com"
            - "*.akamai.com"
        action: "forward"
      
      - name: "intercept_executables"
        priority: 80
        condition:
          type: "file_extension"
          extensions: [".exe", ".dll", ".ps1", ".sh"]
        action: "modify"
        modifier:
          type: "sandbox_file"
          save_original: true
          replace_with: "honeypot"
      
      - name: "monitor_exfiltration"
        priority: 70
        condition:
          type: "upload_detection"
          methods: ["POST", "PUT"]
          min_size: 1048576  # 1MB
        action: "inspect_and_forward"
        modifier:
          type: "content_logging"
          strip_pii: true
      
      - name: "simulate_social_media"
        priority: 60
        condition:
          type: "domain_pattern"
          patterns:
            - "*.facebook.com"
            - "*.twitter.com"
        action: "simulate"
        service: "social-media-sim"
  
  traffic_modifier:
    enabled: true
    
    request_modifiers:
      - strip_auth_headers: true
      - inject_tracking_headers: true
      - sanitize_user_agent: false
    
    response_modifiers:
      - strip_executable_content: true
      - inject_watermark: true
      - limit_response_size: 10485760  # 10MB
  
  logging:
    level: "debug"
    capture_pcap: true
    log_decisions: true
    log_modifications: true
    separate_logs_per_destination: true
```

---

## 🔧 Implementation Plan

### Phase 1: Core Infrastructure (2-3 weeks)

#### 1.1 Network Mode Controller
```
Location: dynamic-analysis/internal/networkmode/
Files:
  - controller.go       # Main controller
  - mode.go            # Mode definitions
  - interceptor.go     # Traffic interception
  - router.go          # Mode-based routing
```

**Key Components**:
```go
package networkmode

type Mode string

const (
    ModeFull Mode = "full"
    ModeHalf Mode = "half"
)

type Controller struct {
    mode           Mode
    config         *Config
    interceptor    *TrafficInterceptor
    decisionEngine *DecisionEngine
    modifier       *TrafficModifier
    logger         *Logger
}

func NewController(mode Mode, config *Config) *Controller {
    return &Controller{
        mode:           mode,
        config:         config,
        interceptor:    NewInterceptor(),
        decisionEngine: NewDecisionEngine(config.Rules),
        modifier:       NewModifier(config.Modifiers),
        logger:         NewLogger(config.Logging),
    }
}

func (c *Controller) HandleRequest(req *Request) (*Response, error) {
    switch c.mode {
    case ModeFull:
        return c.handleFullMode(req)
    case ModeHalf:
        return c.handleHalfMode(req)
    default:
        return nil, ErrInvalidMode
    }
}
```

#### 1.2 Traffic Interceptor
```go
package networkmode

import (
    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
)

type TrafficInterceptor struct {
    pcapHandle *pcap.Handle
    analyzer   *PacketAnalyzer
}

func (ti *TrafficInterceptor) CapturePacket() (*Packet, error) {
    // Capture raw packet
    // Parse protocol
    // Extract request details
    // Return structured Packet
}

type Packet struct {
    Protocol    Protocol
    Source      string
    Destination string
    Headers     map[string]string
    Payload     []byte
    Timestamp   time.Time
}
```

#### 1.3 Decision Engine (for Half Mode)
```go
package networkmode

type DecisionEngine struct {
    rules      []DecisionRule
    whitelist  *DomainList
    blacklist  *DomainList
    ruleCache  *cache.Cache
}

func (de *DecisionEngine) Decide(req *Request) (*Decision, error) {
    // Apply rules in priority order
    // Check whitelist/blacklist
    // Evaluate conditions
    // Return decision with action
}

type Decision struct {
    Action   Action      // FORWARD, BLOCK, MODIFY, SIMULATE
    Reason   string
    Modifier *Modifier
    Metadata map[string]interface{}
}

type Action string

const (
    ActionForward  Action = "forward"
    ActionBlock    Action = "block"
    ActionModify   Action = "modify"
    ActionSimulate Action = "simulate"
)
```

### Phase 2: Half Mode Implementation (2-3 weeks)

#### 2.1 Transparent Proxy
```
Location: dynamic-analysis/internal/networkmode/proxy/
Files:
  - proxy.go           # Main proxy server
  - http_handler.go    # HTTP/HTTPS handling
  - dns_handler.go     # DNS interception
  - modifier.go        # Request/Response modification
```

#### 2.2 Traffic Modifier
```go
package proxy

type TrafficModifier struct {
    requestModifiers  []RequestModifier
    responseModifiers []ResponseModifier
}

type RequestModifier interface {
    ModifyRequest(req *http.Request) (*http.Request, error)
}

type ResponseModifier interface {
    ModifyResponse(resp *http.Response) (*http.Response, error)
}

// Built-in modifiers
type HeaderStripperModifier struct {
    headersToStrip []string
}

type ExecutableSandboxModifier struct {
    sandboxDir string
    saveOriginal bool
}

type ContentLoggerModifier struct {
    logDir       string
    stripPII     bool
    maxSize      int64
}
```

### Phase 3: Configuration & API (1-2 weeks)

#### 3.1 Configuration Management
```yaml
# config/network-mode.yaml
network_mode:
  # Global settings
  mode: "half"  # full or half
  
  # Full mode settings
  full_mode:
    services:
      dns: "inetsim"
      http: "fakenet-ng"
      https: "service-simulation"
    complete_isolation: true
  
  # Half mode settings
  half_mode:
    proxy:
      transparent: true
      ssl_interception: true
    
    decision_rules_file: "/config/decision-rules.yaml"
    
    traffic_modifiers:
      request: "/config/request-modifiers.yaml"
      response: "/config/response-modifiers.yaml"
    
    logging:
      pcap_file: "/logs/traffic.pcap"
      decisions_file: "/logs/decisions.log"
```

#### 3.2 REST API
```go
// API endpoints for managing network mode

// GET /api/network-mode/status
// Returns current mode and statistics

// POST /api/network-mode/switch
// Switch between full and half mode
// Body: {"mode": "full" | "half"}

// GET /api/network-mode/rules
// List all decision rules

// POST /api/network-mode/rules
// Add new decision rule

// GET /api/network-mode/traffic
// Get traffic logs (paginated)

// GET /api/network-mode/decisions
// Get decision logs (what was forwarded/blocked)
```

### Phase 4: Integration & Testing (2 weeks)

#### 4.1 Integration Points
- [ ] Integrate with existing `networksim.go`
- [ ] Update analysis workflow to support mode selection
- [ ] Connect to INetSim and FakeNet-NG
- [ ] Add mode selector to Web UI

#### 4.2 Testing Strategy
```
tests/
  unit/
    - controller_test.go
    - decision_engine_test.go
    - modifier_test.go
  
  integration/
    - full_mode_test.go
    - half_mode_test.go
    - mode_switching_test.go
  
  e2e/
    - malware_sample_full_test.go
    - malware_sample_half_test.go
    - c2_communication_test.go
```

---

## 📈 Use Cases & Examples

### Use Case 1: Full Mode - Analyzing Unknown Malware

**Scenario**: Security researcher nhận được một malware sample chưa biết từ email phishing.

**Configuration**:
```yaml
network_mode:
  mode: "full"
  isolation:
    block_all_external: true
```

**Expected Behavior**:
```
1. Malware attempts DNS lookup for "evil-c2-server.com"
   → INetSim returns 127.0.0.1
   
2. Malware connects to HTTP://evil-c2-server.com
   → FakeNet-NG intercepts and returns fake response
   
3. Malware attempts to download payload.exe
   → Service Simulation returns honeypot executable
   
4. All activity logged, NO external communication
```

**Benefits**:
- ✅ Zero risk of actual infection
- ✅ Observe malware behavior
- ✅ Identify communication patterns
- ✅ Safe for any environment

---

### Use Case 2: Half Mode - Tracking C2 Infrastructure

**Scenario**: Malware analyst muốn theo dõi C2 server thực để thu thập IOCs.

**Configuration**:
```yaml
network_mode:
  mode: "half"
  
  decision_engine:
    rules:
      - name: "forward_to_c2"
        condition:
          domain: "known-c2-domain.com"
        action: "forward"
        modifier:
          log_full_content: true
      
      - name: "block_executables"
        condition:
          content_type: "application/x-msdownload"
        action: "modify"
        modifier:
          sandbox_file: true
```

**Expected Behavior**:
```
1. Malware connects to known-c2-domain.com
   → Decision: FORWARD
   → Request sent to REAL server
   → Response logged completely
   
2. C2 server sends back executable payload
   → Decision: MODIFY
   → Save original exe to /logs/payloads/
   → Return honeypot exe to malware
   
3. Malware attempts to exfiltrate data
   → Decision: INSPECT_AND_FORWARD
   → Log data content (stripped PII)
   → Forward to real destination
```

**Benefits**:
- ✅ Get real C2 responses
- ✅ Download actual payloads safely
- ✅ Track infrastructure changes
- ✅ Controlled risk

---

### Use Case 3: Half Mode - Analyzing Phishing Campaign

**Scenario**: Phân tích phishing malware để hiểu attack chain.

**Configuration**:
```yaml
network_mode:
  mode: "half"
  
  decision_engine:
    rules:
      - name: "allow_initial_download"
        condition:
          domain_pattern: "*.legitimate-cdn.com"
        action: "forward"
      
      - name: "simulate_credential_theft"
        condition:
          path_pattern: "/login*"
          method: "POST"
        action: "simulate"
        service: "fake-auth-server"
      
      - name: "block_further_stages"
        condition:
          stage: "after_first_download"
        action: "simulate"
```

**Expected Behavior**:
```
1. Malware downloads initial payload from CDN
   → Decision: FORWARD to real CDN
   → Payload downloaded and logged
   
2. Malware POSTs credentials to phishing site
   → Decision: SIMULATE
   → Fake success response returned
   → Credentials logged (for analysis)
   
3. Malware attempts to download Stage 2
   → Decision: SIMULATE
   → Return fake payload
   → Prevent actual infection spread
```

---

## 🔐 Security Considerations

### Full Mode Security
- ✅ **Complete Isolation**: No data leaves analysis environment
- ✅ **No Reverse Shell Risk**: Attackers cannot connect back
- ✅ **Safe for Production**: Can run in enterprise networks
- ⚠️ **Limited Intelligence**: Cannot see real C2 infrastructure

### Half Mode Security
- ⚠️ **Potential Data Leaks**: Some data may reach real servers
- ⚠️ **C2 Communication Risk**: Malware may receive real commands
- ⚠️ **Requires Monitoring**: Must log and review all forwarded traffic
- ✅ **Controlled Exposure**: Decision engine limits risk
- ✅ **Content Sanitization**: Modifiers strip sensitive data

### Recommendations
1. **Default to Full Mode** for unknown samples
2. **Use Half Mode** only in dedicated analysis networks
3. **Implement network monitoring** for Half Mode
4. **Require approval** for Half Mode in production
5. **Regular audit** of decision rules and logs

---

## 🎨 Web UI Integration

### Mode Selector
```html
<!-- Analysis Configuration Screen -->
<div class="network-mode-selector">
  <h3>Network Analysis Mode</h3>
  
  <div class="mode-option">
    <input type="radio" name="mode" value="full" checked>
    <label>
      <strong>Full Mode (Isolated)</strong>
      <p>Complete isolation, no external communication. Safest option.</p>
      <span class="badge badge-success">Recommended</span>
    </label>
  </div>
  
  <div class="mode-option">
    <input type="radio" name="mode" value="half">
    <label>
      <strong>Half Mode (Transparent Proxy)</strong>
      <p>Selective forwarding with monitoring. Use for C2 tracking.</p>
      <span class="badge badge-warning">Advanced</span>
    </label>
  </div>
  
  <!-- Half Mode Configuration (shown when selected) -->
  <div id="half-mode-config" style="display: none;">
    <h4>Half Mode Settings</h4>
    
    <div class="form-group">
      <label>Default Action</label>
      <select name="default_action">
        <option value="simulate">Simulate (Safest)</option>
        <option value="forward">Forward (Research)</option>
        <option value="block">Block</option>
      </select>
    </div>
    
    <div class="form-group">
      <label>Decision Rules</label>
      <select name="rules_preset">
        <option value="conservative">Conservative (Block most)</option>
        <option value="balanced">Balanced (Default)</option>
        <option value="permissive">Permissive (Forward most)</option>
        <option value="custom">Custom Rules</option>
      </select>
    </div>
    
    <div class="form-group">
      <label>
        <input type="checkbox" name="sandbox_executables" checked>
        Sandbox executable downloads
      </label>
    </div>
    
    <div class="form-group">
      <label>
        <input type="checkbox" name="log_full_content">
        Log full request/response content
      </label>
    </div>
  </div>
</div>
```

### Traffic Dashboard
```html
<!-- Real-time traffic monitoring for Half Mode -->
<div class="traffic-dashboard">
  <h3>Network Traffic Monitor</h3>
  
  <div class="stats-row">
    <div class="stat-card">
      <h4>Total Requests</h4>
      <p class="stat-value">142</p>
    </div>
    
    <div class="stat-card">
      <h4>Forwarded</h4>
      <p class="stat-value text-warning">23</p>
    </div>
    
    <div class="stat-card">
      <h4>Blocked</h4>
      <p class="stat-value text-danger">45</p>
    </div>
    
    <div class="stat-card">
      <h4>Simulated</h4>
      <p class="stat-value text-success">74</p>
    </div>
  </div>
  
  <div class="traffic-log">
    <table>
      <thead>
        <tr>
          <th>Time</th>
          <th>Protocol</th>
          <th>Destination</th>
          <th>Action</th>
          <th>Reason</th>
        </tr>
      </thead>
      <tbody>
        <tr class="action-forward">
          <td>14:23:45</td>
          <td>HTTP</td>
          <td>evil-c2.com</td>
          <td><span class="badge badge-warning">FORWARD</span></td>
          <td>C2 tracking rule</td>
        </tr>
        <tr class="action-block">
          <td>14:23:50</td>
          <td>HTTP</td>
          <td>malware.exe</td>
          <td><span class="badge badge-danger">BLOCK</span></td>
          <td>Executable download</td>
        </tr>
        <tr class="action-simulate">
          <td>14:23:55</td>
          <td>DNS</td>
          <td>google.com</td>
          <td><span class="badge badge-success">SIMULATE</span></td>
          <td>Default action</td>
        </tr>
      </tbody>
    </table>
  </div>
</div>
```

---

## 📊 Performance & Scalability

### Full Mode Performance
- **Overhead**: Minimal (local simulation only)
- **Latency**: < 5ms (internal Docker network)
- **Throughput**: Limited by simulation services
- **Scalability**: Can run 100+ concurrent analyses

### Half Mode Performance
- **Overhead**: Moderate (proxy + inspection + decision)
- **Latency**: Variable (depends on external servers)
  - Decision making: ~ 1-2ms
  - External request: network dependent
- **Throughput**: Limited by proxy capacity
- **Scalability**: Recommended max 50 concurrent analyses

### Optimization Strategies
1. **Rule caching**: Cache decision results for repeated requests
2. **Async logging**: Non-blocking log writes
3. **Connection pooling**: Reuse HTTP connections
4. **Response caching**: Cache simulation responses
5. **Parallel processing**: Multi-threaded packet inspection

---

## 🚀 Migration Path

### For Existing Users

#### Step 1: Update Configuration
```bash
# Add network mode configuration
cp config/network-mode.example.yaml config/network-mode.yaml
```

#### Step 2: Backward Compatibility
```yaml
# Default configuration maintains current behavior
network_mode:
  mode: "full"  # Same as current isolated mode
  auto_detect: true  # Auto-switch based on analysis type
```

#### Step 3: Gradual Rollout
1. **Week 1-2**: Full mode only (default)
2. **Week 3-4**: Beta test Half mode with opt-in
3. **Week 5+**: General availability

### Breaking Changes
- ❌ None - Full mode behaves identically to current system
- ✅ All existing analyses will work unchanged
- ✅ New mode is opt-in

---

## 📚 Documentation Updates Required

1. **User Guide**: Add Network Mode Selection section
2. **API Documentation**: Document new endpoints
3. **Configuration Reference**: Document all new config options
4. **Security Best Practices**: When to use each mode
5. **Troubleshooting**: Common issues and solutions

---

## ✅ Success Metrics

### Full Mode
- ✅ 100% traffic isolation
- ✅ < 5ms average response time
- ✅ Support all major protocols
- ✅ Comprehensive logging

### Half Mode
- ✅ Decision accuracy > 99%
- ✅ < 10ms decision overhead
- ✅ Executable sandboxing rate 100%
- ✅ Zero false negatives on known threats
- ✅ Configurable false positive rate

---

## 🎯 Conclusion

Ý tưởng **Full Mode** và **Half Mode** là một bổ sung **rất có giá trị** cho Pack-A-Mal:

### ✅ Khả Thi
- Có sẵn 80% infrastructure (INetSim, FakeNet-NG, NetworkSim)
- Chỉ cần thêm Network Mode Controller và Decision Engine
- Không breaking changes cho users hiện tại

### ✅ Giá Trị
- **Full Mode**: An toàn tối đa, phù hợp production
- **Half Mode**: Nghiên cứu sâu, thu thập intelligence
- Linh hoạt cho nhiều use cases

### ✅ Thời Gian Triển Khai
- **Phase 1-2**: 4-6 tuần (core + half mode)
- **Phase 3**: 1-2 tuần (config + API)
- **Phase 4**: 2 tuần (testing)
- **Total**: ~8-10 tuần cho MVP

### 🚀 Khuyến Nghị
**Nên triển khai!** Đây là một tính năng quan trọng giúp Pack-A-Mal trở thành một platform phân tích malware toàn diện và chuyên nghiệp hơn.
