# containerlib

`containerlib` is a Go library that provides abstraction from different container management runtimes.
With one code, it is possible to work with containers running on different backends (for now, only Docker is supported).

[![Go Report Card](https://goreportcard.com/badge/github.com/samtaborsky/containerlib)](https://goreportcard.com/report/github.com/samtaborsky/containerlib)

[Go Packages](https://pkg.go.dev/github.com/samtaborsky/containerlib) |
[License](https://mit-license.org/)

---

## About The Project

`containerlib` was created to simplify and standardize interactions with container runtimes in Go applications. Instead of writing runtime-specific code, you can use `containerlib`'s unified API to manage the lifecycle of your containers.

This is especially useful for applications that need to be portable across different environments or for building higher-level tools on top of container technology.

### Key Features

*   **Unified API**: A single, clean interface for container and system operations.
*   **Container Lifecycle Management**: Create, Start, Stop, and Remove containers.
*   **System Information**: Inspect the host container runtime environment.
*   **Resource Management**: Define CPU, memory, ports, and volumes for containers.
*   **Extensibility**: Designed to support multiple container backends.
*   **Clear Error Handling**: Provides distinct error types for common issues.

---

## Getting Started

Follow these steps to integrate `containerlib` into your project.

### Prerequisites

*   Go 1.25+
*   A running container runtime (e.g., Docker)

### Installation

To add `containerlib` as a dependency to your project, run the following command:

```sh
go get github.com/samtaborsky/containerlib
```

---

## Usage

Here is a quick example of how to use `containerlib` to create, inspect, and clean up an `nginx` container.

```go

```

## API Documentation

TBD

## Error Handling

`containerlib` returns distinct error variables for known failures, allowing you to handle them programmatically. You can use `errors.Is()` to check for specific errors.

**Example:**
```go

```

Types of errors include:
*   `types.ErrContainerNotFound`
*   `types.ErrImageNotFound`
*   `types.ErrConflict`
*   `types.ErrInvalidInput`
*   `types.ErrConnectionFailed`

## Supported Runtimes

-   [x] Docker
-   [ ] Podman (Planned)
-   [ ] containerd (Planned)

## Contributing

Any contributions you make are greatly appreciated.

1.  Fork the Project
2.  Create your Feature Branch (`git checkout -b feature/NewFeature`)
3.  Commit your Changes (`git commit -m 'Add a feature'`)
4.  Push to the Branch (`git push origin feature/NewFeature`)
5.  Open a Pull Request

## License

This project is licensed under the [MIT License](LICENSE).

## Contact

Samuel Táborský - samuel.taborsky01@upol.cz

Project Link: https://github.com/samtaborsky/containerlib
