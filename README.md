# Go Load Balancer

## What is this?

This is a basic learning project demonstrating how to build an HTTP load balancer using Go's standard libraries. It listens for incoming HTTP requests and distributes them across a list of predefined backend servers.

## Features

* Listens on a configurable port (e.g., `:8080`).
* Manages a list of backend HTTP servers.
* Distributes incoming requests using a **Round Robin** algorithm.
* Performs periodic **health checks** (simple TCP connection test) on backend servers.
* Marks backends as UP or DOWN based on health checks.
* Only forwards requests to currently healthy (UP) backend servers.
* Returns an HTTP 503 error if no backend servers are available.
* Uses Go's standard `net/http` and `net/http/httputil` packages.
* Logs basic events like startup, backend status changes, and request routing.

## How it Works

1. The load balancer starts (`main.go`) and reads a list of backend server addresses.
2. It begins periodically checking if each backend server is reachable (health checks).
3. When a client sends an HTTP request to the load balancer (e.g., `http://localhost:8080`):
    * The load balancer selects the next *healthy* backend server from its list using the Round Robin method.
    * It acts as a reverse proxy, forwarding the client's request to the selected backend.
    * The backend server processes the request and sends its response back to the load balancer.
    * The load balancer sends that response back to the original client.
4. If the load balancer can't find any healthy backends, it sends a "503 Service Unavailable" response directly to the client.

## How to Run

1. **Run Backend Servers:**
    * You need at least two instances of the simple backend server running.
    * Open two separate terminals.
    * In each terminal, navigate to the `loadbalancer/backend/` directory.
    * Run the following commands (one in each terminal):

        ```bash
        # Terminal 1
        go run backend.go 9001 "Response from Backend 1"

        # Terminal 2
        go run backend.go 9002 "Response from Backend 2"
        ```

2. **Run Load Balancer:**
    * Open a third terminal.
    * Navigate to the main `loadbalancer/` directory.
    * Run the load balancer:

        ```bash
        go run main.go
        ```

    * Observe the logs indicating startup, backend detection, and health status changes.

3. **Test:**
    * Open a web browser and navigate to `http://localhost:8080`.
    * You should see a response from one of the backend servers.
    * Refresh the page several times. You should see the response alternating between "Response from Backend 1" and "Response from Backend 2".
