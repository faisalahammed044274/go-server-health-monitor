Gotcha — you basically want a clean, structured doc for your Go **Server Health Checker** app so it reads like a legit project README/manual.
Here’s a solid draft:

---

# **Go Server Health Checker**

## **Overview**

The **Server Health Checker** is a Go-based CLI tool that monitors the availability and responsiveness of servers via TCP, HTTP, and HTTPS protocols.
It can:

* Run a one-time check
* Perform continuous monitoring
* Generate JSON health reports
* Work with a configuration file defining multiple servers

Ideal for sysadmins, DevOps engineers, and developers who want a simple, lightweight health monitoring solution without bloated dependencies.

---

## **Features**

* ✅ **Supports multiple protocols:** TCP, HTTP, HTTPS
* ⏱ **Response time measurement** (in milliseconds)
* 🔄 **Continuous monitoring** at configurable intervals
* 📄 **JSON report generation** for logs or integrations
* 📂 **Config file-based setup** for multiple server entries
* 🛠 **Sample config generator** for quick start
* ⚡ **Concurrent checks** using goroutines for speed

---

## **Installation**

You need **Go 1.18+** installed.

```bash
# Clone the repository
git clone https://github.com/yourusername/server-health-checker.git
cd server-health-checker

# Run directly
go run main.go -sample
```

Or build it:

```bash
go build -o health-checker main.go
./health-checker -sample
```

---

## **Configuration File**

The configuration file is a JSON file (default: `servers.json`).

**Example:**

```json
{
  "servers": [
    { "name": "Google DNS", "host": "8.8.8.8", "port": 53, "protocol": "tcp", "timeout": 5 },
    { "name": "Google", "host": "google.com", "port": 80, "protocol": "http", "timeout": 10 },
    { "name": "GitHub", "host": "github.com", "port": 443, "protocol": "https", "timeout": 10 },
    { "name": "Local SSH", "host": "localhost", "port": 22, "protocol": "tcp", "timeout": 3 },
    { "name": "Local Web", "host": "localhost", "port": 8080, "protocol": "http", "timeout": 5 }
  ]
}
```

**Fields:**

| Field      | Type   | Description                 |
| ---------- | ------ | --------------------------- |
| `name`     | string | Display name for the server |
| `host`     | string | Hostname or IP address      |
| `port`     | int    | Port number                 |
| `protocol` | string | `tcp`, `http`, or `https`   |
| `timeout`  | int    | Timeout in seconds          |

---

## **Command-line Options**

| Flag              | Description                                        |
| ----------------- | -------------------------------------------------- |
| `-config <file>`  | Path to config file (default: `servers.json`)      |
| `-once`           | Run a single check and exit                        |
| `-interval <dur>` | Continuous monitoring interval (e.g., `30s`, `1m`) |
| `-report <file>`  | Generate JSON report to file                       |
| `-sample`         | Create a sample `servers.json` config file         |
| `-help`           | Show help and usage examples                       |

---

## **Usage Examples**

**Generate a sample config file:**

```bash
go run main.go -sample
```

**Run a one-time health check:**

```bash
go run main.go -once
```

**Run continuous monitoring every 60 seconds:**

```bash
go run main.go -interval 60s
```

**Generate a health report in JSON:**

```bash
go run main.go -report report.json
```

---

## **Example Output**

**One-time check example:**

```
Loaded 5 servers from servers.json
Go version: go1.22.1, OS: linux, Arch: amd64
Checking 5 servers...
✓ [UP] google.com:80 - Google (53ms)
✓ [UP] github.com:443 - GitHub (120ms)
✗ [DOWN] localhost:8080 - Local Web (Error: connection refused)
...
Summary: 4 UP, 1 DOWN
```

---

## **JSON Report Example**

```json
{
  "timestamp": "2025-08-13T14:52:10Z",
  "results": [
    {
      "server": {
        "name": "Google DNS",
        "host": "8.8.8.8",
        "port": 53,
        "protocol": "tcp",
        "timeout": 5
      },
      "status": "UP",
      "response_time": 42,
      "timestamp": "2025-08-13T14:52:10Z"
    }
  ],
  "summary": {
    "total": 5,
    "up": 4,
    "down": 1
  }
}
```

---

## **How It Works**

1. **Load Config** → Reads the `servers.json` file.
2. **Concurrent Checks** → Each server is tested in its own goroutine.
3. **Protocol Handling** →

   * TCP: Uses `net.DialTimeout`
   * HTTP/HTTPS: Uses `http.Client` with status code validation
4. **Collect Results** → Aggregates status, response times, and errors.
5. **Output / Report** → Prints results to console or saves to JSON.

---

## **Future Improvements**

* Email/SMS/Slack alerts for downtime
* Prometheus metrics export
* Support for ICMP ping
* Custom health check endpoints
* Colored CLI output
