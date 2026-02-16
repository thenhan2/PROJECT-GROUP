# 🧠 AI Coding Prompt – Network Mode Controller (Full Mode & Half Mode)

## 📌 Context

You are a **senior Go backend engineer and malware analysis platform developer**.

Your task is to implement a **Network Mode Controller** for a malware dynamic analysis system called **Pack-A-Mal**.

This document is the **final and authoritative specification**.  
Do **NOT** simplify the design.  
Do **NOT** remove security controls.  
Prefer **safety and correctness over performance**.

---

## 🎯 Goal

Implement a **dual network mode architecture**:

### 1. Full Mode (Isolated Mode)
- 100% simulated network
- No external internet access
- All traffic routed to internal simulation services

### 2. Half Mode (Transparent Proxy Mode)
- Intercept all outbound traffic
- Inspect, log, and decide per request
- Forward, block, modify, or simulate traffic based on rules

---

## 🏗️ Required Components (MUST IMPLEMENT)

### 1. Network Mode Controller

**Language**: Go  
**Location**:
```
dynamic-analysis/internal/networkmode/
```

**Files**:
```
controller.go
mode.go
interceptor.go
router.go
```

```go
type Controller struct {
    mode           Mode
    config         *Config
    interceptor    *TrafficInterceptor
    decisionEngine *DecisionEngine
    modifier       *TrafficModifier
    logger         *Logger
}

func (c *Controller) HandleRequest(req *Request) (*Response, error)
```

---

## 🔒 Full Mode Logic

- No external traffic allowed
- All protocols must be simulated
- Log everything
- Capture PCAP

---

## 🌐 Half Mode Logic

```
Request
 → Deep Packet Inspection
 → Decision Engine
 → Action (FORWARD | BLOCK | MODIFY | SIMULATE)
 → Traffic Modifier
 → Logging
```

---

## 🧠 Decision Engine

- Rule-based
- Priority-driven
- Deterministic

```go
type Decision struct {
    Action   Action
    Reason   string
    Modifier *Modifier
}
```

---

## 🔐 Security Rules

- Default mode = FULL
- Half Mode requires explicit enable
- Crash or error → fallback to FULL
- Never trust malware input

---

## 🧪 Testing

```
tests/
  unit/
  integration/
  e2e/
```

---

## 🚀 Final Instruction

Implement step-by-step.  
If ambiguous, **choose the safest option**.
