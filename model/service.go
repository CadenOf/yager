package model

type Service struct {
	SvcMeta ServiceMeta `json:"svcMeta"`
	SvcSpec ServiceSpec `json:""svcSpec`
}

type ServiceMeta struct {
	AppMeta     AppMetaInfo       `json:"appMeta"`
	Name        string            `json:""name` // Name must be unique within a namespace.
	Namespace   string            `json:"namespace"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type ServiceSpec struct {
	Ports    []ServicePort     `json:"ports,omitempty"`
	Selector map[string]string `json:"selector,omitempty"`

	// type determines how the Service is exposed. Defaults to ClusterIP. Valid
	// options are ExternalName, ClusterIP, NodePort, and LoadBalancer.
	Type string `json:"type,omitempty"`
}

type ServicePort struct {
	// The IP protocol for this port. Supports "TCP", "UDP", and "SCTP".
	Protocol string `json:"protocol,omitempty"`

	// The port that will be exposed by this service.
	Port int32 `json:"port"`

	// Number or name of the port to access on the pods targeted by the service.
	// Number must be in the range 1 to 65535.
	TargetPort int32 `json:"targetPort,omitempty"`

	// The port on each node on which this service is exposed when type=NodePort or LoadBalancer.
	NodePort int32 `json:"nodePort,omitempty"`
}
