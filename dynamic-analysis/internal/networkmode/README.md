# Network Mode Package

## Overview

The `networkmode` package implements a **dual-mode network controller** for the Pack-A-Mal malware analysis platform. It provides two distinct operating modes for handling network traffic during dynamic analysis:

- **Full Mode (Isolated)**: Complete network isolation with all traffic simulated
- **Half Mode (Transparent Proxy)**: Selective forwarding with deep packet inspection and decision-based routing

## Architecture

```
┌────────────────────────────────────────────────────────────┐
│                    Controller                              │
│  - Mode management                                         │
│  - Request orchestration                                   │
│  - Statistics tracking                                     │
└──────────────┬─────────────────────────────────────────────┘
               │
    ┌──────────┼──────────┐
    ↓          ↓          ↓
┌─────────┐ ┌──────────┐ ┌─────────┐
│Intercept│ │ Decision │ │ Router  │
│         │ │  Engine  │ │         │
└─────────┘ └──────────┘ └─────────┘
    │            │            │
    └────────────┼────────────┘
                 ↓
         ┌────────────┐
         │  Modifier  │
         └────────────┘
```

## Usage

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "log/slog"
    
    "github.com/ossf/package-analysis/internal/networkmode"
)

func main() {
    // Create default configuration (Full Mode)
    config := networkmode.DefaultConfig()
    
    // Create controller
    controller, err := networkmode.NewController(config, slog.Default())
    if err != nil {
        log.Fatal(err)
    }
    defer controller.Close()
    
    // Handle a request
    req := &networkmode.Request{
        ID:       "req-001",
        Protocol: string(networkmode.ProtocolHTTP),
        Domain:   "example.com",
        Path:     "/api/data",
        Method:   "GET",
    }
    
    ctx := context.Background()
    resp, err := controller.HandleRequest(ctx, req)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Response: %s (source: %s)", resp.Body, resp.Source)
}
```

### Full Mode Example

```go
config := networkmode.DefaultConfig()
config.Mode = networkmode.ModeFull
config.FullMode.CompleteIsolation = true

controller, _ := networkmode.NewController(config, slog.Default())
defer controller.Close()

// All requests will be simulated
resp, _ := controller.HandleRequest(ctx, request)
// resp.Source == "simulated"
```

### Half Mode Example

```go
config := networkmode.DefaultConfig()
config.Mode = networkmode.ModeHalf
config.HalfMode.Enabled = true
config.HalfMode.DefaultAction = networkmode.ActionSimulate

// Add custom rules
controller, _ := networkmode.NewController(config, slog.Default())
defer controller.Close()

// Add a custom rule
rule := networkmode.DecisionRule{
    Name:     "allow_cdn",
    Priority: 100,
    Enabled:  true,
    Action:   networkmode.ActionForward,
    Condition: &networkmode.RuleCondition{
        Type: networkmode.ConditionDomainWhitelist,
        Domains: []string{"*.cloudflare.com"},
    },
}
controller.AddDecisionRule(ctx, rule)

// Requests will be evaluated by decision engine
resp, _ := controller.HandleRequest(ctx, request)
```

### Switching Modes at Runtime

```go
// Start in Full Mode
controller, _ := networkmode.NewController(config, slog.Default())

// Switch to Half Mode
ctx := context.Background()
err := controller.SwitchMode(ctx, networkmode.ModeHalf)
if err != nil {
    log.Fatal(err)
}

// Switch back to Full Mode
err = controller.SwitchMode(ctx, networkmode.ModeFull)
```

## Components

### Controller

The main orchestrator that manages all components and handles requests.

**Key Methods:**
- `HandleRequest(ctx, req) (*Response, error)` - Process a network request
- `SwitchMode(ctx, mode) error` - Switch between Full and Half modes
- `GetStats() *Stats` - Get controller statistics
- `Health(ctx) error` - Check controller health

### Decision Engine

Makes decisions about how to handle traffic in Half Mode based on rules.

**Key Methods:**
- `AddRule(rule) error` - Add a decision rule
- `Decide(ctx, req) (*Decision, error)` - Make a decision for a request
- `GetRules() []DecisionRule` - Get all rules

**Decision Actions:**
- `ActionForward` - Forward to real destination
- `ActionBlock` - Block the request
- `ActionModify` - Modify and forward
- `ActionSimulate` - Use simulation

### Traffic Modifier

Modifies requests and responses based on rules.

**Features:**
- Strip authentication headers
- Inject tracking headers
- Sandbox executable downloads
- Strip PII from responses
- Limit response sizes

### Router

Routes traffic based on mode and decisions.

**Full Mode Routing:**
- DNS → INetSim (172.20.0.2:53)
- HTTP → FakeNet-NG (172.20.0.3:80)
- HTTPS → FakeNet-NG (172.20.0.3:443)
- SMTP → INetSim (172.20.0.2:25)
- FTP → INetSim (172.20.0.2:21)

**Half Mode Routing:**
- Based on Decision Engine actions
- Supports real internet forwarding
- Request blocking
- Traffic modification

## Configuration

### Configuration File

Create `config/network-mode.yaml`:

```yaml
network_mode:
  mode: "full"  # or "half"
  
  full_mode:
    complete_isolation: true
    services:
      dns: "inetsim"
      dns_address: "172.20.0.2:53"
      # ... other services
  
  half_mode:
    enabled: false  # Must be explicitly enabled
    default_action: "simulate"
    # ... other settings
  
  logging:
    level: "info"
    capture_pcap: true
```

### Decision Rules

Create `config/decision-rules.yaml`:

```yaml
rules:
  - name: "block_malware"
    priority: 100
    enabled: true
    condition:
      type: "domain_blacklist"
      domains: ["*.evil.com"]
    action: "block"
  
  - name: "allow_cdn"
    priority: 90
    enabled: true
    condition:
      type: "domain_whitelist"
      domains: ["*.cloudflare.com"]
    action: "forward"
```

## Testing

Run unit tests:

```bash
cd internal/networkmode
go test -v
```

Run specific tests:

```bash
go test -v -run TestController_HandleRequest
go test -v -run TestDecisionEngine
```

## Security Considerations

### Full Mode
✅ **Complete isolation** - No external communication
✅ **Safe by default** - Default mode
✅ **No data leaks** - All traffic simulated
⚠️ **Limited intelligence** - Cannot observe real C2 infrastructure

### Half Mode
⚠️ **External communication** - Can forward to real internet
⚠️ **Potential data leaks** - If rules are misconfigured
⚠️ **Requires monitoring** - Must audit all forwarded traffic
✅ **Controlled exposure** - Decision engine limits risk
✅ **Content sanitization** - Strips sensitive data

### Best Practices

1. **Default to Full Mode** for unknown/untrusted samples
2. **Enable Half Mode** only in isolated analysis networks
3. **Review decision rules** regularly
4. **Monitor all forwarded traffic** 
5. **Use sandbox for executables** always
6. **Audit logs frequently**

## Statistics

Get real-time statistics:

```go
stats := controller.GetStats()
fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
fmt.Printf("Forwarded: %d\n", stats.ForwardedRequests)
fmt.Printf("Blocked: %d\n", stats.BlockedRequests)
fmt.Printf("Simulated: %d\n", stats.SimulatedRequests)
```

## Logging

The package provides comprehensive logging:

- **Traffic logs** - All requests and responses
- **Decision logs** - Decision engine actions
- **Modification logs** - Traffic modifications
- **PCAP capture** - Raw packet capture

Logs are written to:
- `/logs/traffic.log` - Traffic events
- `/logs/decisions.log` - Decision events
- `/logs/traffic.pcap` - Packet capture
- `/logs/executables/` - Sandboxed files

## Error Handling

The controller implements fail-safe behavior:

- **Invalid mode** → Falls back to Full Mode
- **Decision engine error** → Falls back to simulation
- **Modification error** → Uses original request/response
- **Routing error** → Returns error response

```go
resp, err := controller.HandleRequest(ctx, req)
if err != nil {
    // Handle error
    // Controller has already logged the error
    return
}
```

## Performance

### Full Mode
- **Overhead**: Minimal (local simulation)
- **Latency**: < 5ms
- **Throughput**: Limited by simulation services
- **Scalability**: 100+ concurrent analyses

### Half Mode
- **Overhead**: Moderate (inspection + decision + modification)
- **Latency**: Variable (depends on external servers)
  - Decision making: ~1-2ms
  - External requests: network dependent
- **Throughput**: Limited by proxy capacity
- **Scalability**: ~50 concurrent analyses

## License

Apache 2.0

## Author

Pack-A-Mal Development Team
