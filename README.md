<div align="center">

<h1>Paleo Proxy</h1>

<img src="./prehistoric-gopher.webp" alt="prehistoric-gopher" width="400"/>

**A naive reverse proxy implementation in golang**

</div>

---

## Overview

**Paleo Proxy** is a minimal reverse proxy implementation in Go that lets you define services through a simple YAML configuration file. It supports basic load balancing by randomly routing requests to multiple upstream servers.


---

## Features

- Simple and readable YAML-based service configuration
- Subdomain-based request routing
- Randomized load balancing across multiple endpoints
- Docker Compose setup for a quick local demo

---

## Prerequisites

- [Go 1.18+](https://go.dev/doc/install)
- [Docker](https://www.docker.com/) (optional, for running the demo)

---

## Getting Started (Demo)

You can spin up a demo using Docker Compose:

```bash
# Start 5 'whoami' containers and an 'echo' service
docker compose up -d --scale whoami=5

# Navigate to the source directory
cd src/

# Build the Paleo Proxy binary
go mod tidy
go build

# Run the proxy with the example configuration
./paleo-proxy -config ./config.example.yaml
```

Once running, you can visit:

- [http://whoami.localhost:8080](http://whoami.localhost:8080) — shows randomized responses from 5 whoami services
- [http://echo.localhost:8080](http://echo.localhost:8080) — a simple echo service

---

## Configuration Example

```yaml
services:
  - name: "service1"
    description: "This is the first service configuration"
    enabled: true
    subdomain: "demo"
    endpoints:
      - ip: "10.5.0.2"
      - ip: "10.5.0.3"
      - ip: "10.5.0.4"
      - ip: "10.5.0.5"
      - ip: "10.5.0.6"
    port: 8080

  - name: "service2"
    description: "This is the second service configuration"
    enabled: true
    subdomain: "second-service"
    endpoints:
      - ip: "10.6.0.2"
    port: 8080
```

---

## Goal

This project is more of a simple example than a production-ready solution. However, contributions are welcome! If you have ideas for improvements or features, feel free to open an issue or submit a pull request.

---

## License

MIT License — see the [LICENSE](./LICENSE) file for details.
