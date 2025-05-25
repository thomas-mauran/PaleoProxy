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
- Access logging for each request
- Docker Compose setup for a quick local demo
- **Dynamic mode support using Docker events**

---

## Prerequisites

- [Go 1.18+](https://go.dev/doc/install)
- [Docker](https://www.docker.com/) (optional, for running the demo)

---

## Getting Started 

### Config mode

You can spin up a demo using Docker Compose:

```
# Start 5 'whoami' containers and an 'echo' service
docker compose up -d --scale whoami=5

# Navigate to the source directory
cd src/

# Build the Paleo Proxy binary
go mod tidy
go build

# Run the proxy with the example configuration
./paleo-proxy ./config.example.yaml
```

Once running, you can visit:

- [http://whoami.localhost:8080](http://whoami.localhost:8080) — shows randomized responses from 5 whoami services
- [http://echo.localhost:8080](http://echo.localhost:8080) — a simple echo service

---

### Dynamic Mode

In addition to the static YAML-based configuration, **Paleo Proxy** supports a **dynamic mode** where services are automatically registered based on Docker container events.

### How It Works

- The proxy listens for Docker container lifecycle events.
- If a container has the `paleo-subdomain` label, the proxy will automatically route traffic to it upon container start.
- Services are assumed to be exposed on port `8080` (this can be extended in future versions).

### Running in Dynamic Mode

To launch Paleo Proxy in dynamic mode:

```
# Build the Paleo Proxy binary
go mod tidy
go build

# Run Paleo Proxy with the config file and enable dynamic mode
./paleo-proxy ./config.example.yaml dynamic
```

> **Note:** The config file path is still required, but not used in dynamic mode. It's a placeholder to satisfy argument parsing and needs to be refactored in the future.

### Labeling Containers

Make sure your Docker containers include the `paleo-subdomain` label:

```
services:
  whoami:
    image: traefik/whoami
    labels:
      - paleo-subdomain=whoami-dynamic
    networks:
      - default
```

Then visit:

- [http://whoami-dynamic.localhost:8080](http://whoami-dynamic.localhost:8080)

> The proxy will route requests to any container labeled with a `paleo-subdomain`, using that value as the subdomain.

---

## Configuration Example

```
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
