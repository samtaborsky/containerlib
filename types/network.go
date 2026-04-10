// SPDX-License-Identifier: MIT

package types

import "net/netip"

// PortBinding represents a bind between a port on the host and a port inside the container.
// It is used when there is need to access network resources running inside the container from the host.
type PortBinding struct {
	// HostIP specifies the IP address on the host to bind to (e.g. "192.168.0.1").
	// If left empty, a default "0.0.0.0" address will be used, which binds the port to all
	// available IP addresses on the host.
	HostIP netip.Addr `json:"hostIP"`

	// HostPort specifies the port number on the host (e.g. 8080).
	HostPort int `json:"hostPort"`
	// ContainerPort specifies the port number inside the container (e.g. 80).
	ContainerPort int `json:"containerPort"`

	// Protocol specifies the communication protocol to use (e.g. "tcp", "udp").
	Protocol string `json:"protocol"`
}
