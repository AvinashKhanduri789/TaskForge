
# TaskForge — Event-Driven Distributed Execution Platform (v1)



TaskForge is an **event-driven, distributed execution platform** designed to explore and validate core systems concepts such as **scheduling, concurrency, and service-to-service communication**.

The current version (**v1**) focuses on **architectural correctness, controlled concurrency, and gRPC-based coordination**, intentionally keeping the execution layer simple while establishing a strong foundation for **future runtime isolation and multi-language execution**.


This project is **not a finished product** — it is a **foundational scaffold / architectural skeleton** designed to evolve into a full-fledged serverless runtime.

---

## 🚀 What TaskForge Solves

- Accepts function execution requests from users
- Schedules work across multiple workers
- Executes tasks concurrently
- Streams results back reliably
- Scales horizontally by adding workers

All of this is done using **Go, gRPC, goroutines, and event-driven design**.

---

## 🧱 High-Level Architecture (v1)
 
![TaskForge v1 Architecture](./v1 architecture digram.png)

### Core Components

#### 1. API Gateway (Node.js)
- Exposes HTTP APIs for users
- Translates HTTP → gRPC
- Decodes binary execution output into JSON
- Acts as the system’s external interface

#### 2. Scheduler Service (Go + gRPC)
- Central coordination layer
- Accepts execution requests from Gateway
- Tracks connected workers and their load
- Dispatches work using **least-loaded scheduling**
- Maintains execution state
- Communicates with workers via **bidirectional gRPC streams**

#### 3. Worker Service (Go)
- Connects to Scheduler as a **client**
- Maintains a **worker pool (goroutines)**
- Executes jobs concurrently
- Streams execution results back to Scheduler
- Periodically sends **heartbeats** to report load

---

## 🔄 Execution Flow (v1)

1. User registers a function via Gateway
2. User triggers execution
3. Scheduler:
   - Creates an execution record
   - Selects the least-loaded worker
   - Streams execution request
4. Worker:
   - Queues job
   - Executes inside goroutine pool
   - Sends result back
5. Scheduler updates execution state
6. Gateway exposes execution status & output

---

## ⚙️ Concurrency Model

TaskForge uses **multi-layer concurrency**:

### System-Level Concurrency
- Achieved by running **multiple worker processes**
- Scheduler distributes work across workers

### Worker-Level Concurrency
- Each worker runs a **goroutine pool**
- Configurable pool size
- Non-blocking execution

**Total concurrency = (number of workers) × (worker pool size)**

---

## 📦 Data Storage (v1)

Current storage is intentionally simple:

- Functions → JSON files
- Executions → JSON files
- Stored on local filesystem

This choice allows:
- Easy debugging
- Zero external dependencies
- Focus on architecture, not persistence

---

## 🧠 Why This Design (v1 Goals)

The goal of v1 is **NOT** to:
- Run arbitrary untrusted code safely
- Support multiple languages
- Be production-ready

The goal **IS** to:
- Prove the event-driven architecture
- Validate gRPC streaming
- Handle real concurrency
- Build a scalable control plane
- Create a solid base for future execution engines

---

## 🚧 Known Limitations (v1)

- No sandboxing
- No isolation between executions
- JSON-based persistence
- No retries or failure recovery
- No execution timeouts
- No autoscaling

These are **intentional trade-offs**.

---

## 🔮 Roadmap (Future Versions)

### v2 — Containerized Execution
- Use **Docker** to execute functions
- Each execution runs in an isolated container
- Support multiple languages (Go, Python, JS, etc.)
- Workers become container orchestrators

### v3 — Queue-Based Scheduling
- Introduce **Redis / message queues**
- Decouple scheduling from execution
- Enable buffering, retries, and backpressure
- Improve fault tolerance

### v4 — Persistent Storage
- Replace JSON files with **MongoDB**
- Durable execution history
- Scalable metadata storage

### v5 — Production Features
- Execution timeouts
- Cancellation
- Retry policies
- Priority queues
- Autoscaling workers
- Metrics & observability

---

## 🧪 Current State of the Project

> TaskForge is currently an **architectural skeleton / foundational implementation**.

It is intentionally minimal, but:
- Architecturally sound
- Horizontally scalable
- Concurrency-safe
- Ready to evolve

---

## 🛠 Tech Stack

- **Go** — Scheduler & Worker services
- **Node.js** — API Gateway
- **gRPC** — Service-to-service communication
- **Protocol Buffers** — Contracts
- **Goroutines & Channels** — Concurrency
- **JSON (filesystem)** — Temporary persistence

---

## 📌 Key Takeaway

TaskForge v1 proves that:
- Serverless systems are fundamentally **event-driven**
- gRPC streaming is powerful for control planes
- Concurrency ≠ request count
- Horizontal scaling starts with architecture, not infrastructure

---



