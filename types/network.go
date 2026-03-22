// SPDX-License-Identifier: MIT

package types

// PortBinding represents a bind between a port on the host and a port inside the container.
// It is used when there is need to access network resources running inside the container from the host.
type PortBinding struct {
	// HostIP specifies the IP address on the host to bind to (e.g. "192.168.0.1").
	HostIP string

	// HostPort specifies the port number on the host (e.g. 8080).
	HostPort int
	// ContainerPort specifies the port number inside the container (e.g. 80).
	ContainerPort int

	// Protocol specifies the communication protocol to use (e.g. "tcp", "udp").
	Protocol string
}
