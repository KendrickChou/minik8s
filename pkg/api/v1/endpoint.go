package v1

type Endpoint struct {
	TypeMeta
	ObjectMeta `json:"metadata"`

	ServiceIp string `json:"serviceip,omitempty"`

	// The set of all endpoints is the union of all subsets. Addresses are placed into
	// subsets according to the IPs they share. A single address with multiple ports,
	// some of which are ready and some of which are not (because they come from
	// different containers) will result in the address being displayed in different
	// subsets for the different ports. No address will appear in both Addresses and
	// NotReadyAddresses in the same subset.
	// Sets of addresses and ports that comprise a service.
	// +optional
	Subset EndpointSubset `json:"subsets,omitempty"`
}

// EndpointSubset is a group of addresses with a common set of ports. The
// expanded set of endpoints is the Cartesian product of Addresses x Ports.
// For example, given:
//   {
//     Addresses: [{"ip": "10.10.1.1"}, {"ip": "10.10.2.2"}],
//     Ports:     [{"name": "a", "port": 8675}, {"name": "b", "port": 309}]
//   }
// The resulting set of endpoints can be viewed as:
//     a: [ 10.10.1.1:8675, 10.10.2.2:8675 ],
//     b: [ 10.10.1.1:309, 10.10.2.2:309 ]
type EndpointSubset struct {
	// IP addresses which offer the related ports that are marked as ready. These endpoints
	// should be considered safe for load balancers and clients to utilize.
	// +optional
	Addresses []EndpointAddress `json:"addresses,omitempty"`

	// Port numbers available on the related IP addresses.
	// +optional
	Ports []EndpointPort `json:"ports,omitempty"`
}

type EndpointAddress struct {
	// The IP of this endpoint.
	// May not be loopback (127.0.0.0/8), link-local (169.254.0.0/16),
	// or link-local multicast ((224.0.0.0/24).
	// IPv6 is also accepted but not fully supported on all platforms. Also, certain
	// kubernetes components, like kube-proxy, are not IPv6 ready.
	IP string `json:"ip"`
	// The Hostname of this endpoint
	// +optional
	Hostname string `json:"hostname,omitempty"`
	// Optional: Node hosting this endpoint. This can be used to determine endpoints local to a node.
	// +optional
	NodeName *string `json:"nodeName,omitempty"`
}

// EndpointPort is a tuple that describes a single port.
type EndpointPort struct {
	// The name of this port.  This must match the 'name' field in the
	// corresponding ServicePort.
	// Must be a DNS_LABEL.
	// Optional only if one port is defined.
	// +optional
	Name string `json:"name,omitempty"`

	// The port number of the endpoint.
	Port int32 `json:"port"`

	ServicePort int32 `json:"serviceport"`

	// The IP protocol for this port.
	// Must be UDP, TCP, or SCTP.
	// Default is TCP.
	// +optional
	Protocol string `json:"protocol,omitempty"`
}
