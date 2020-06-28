package flags

import (
	"flag"
)

const (
	DefaultNamespace = "default"
	DefaultTlsSecret = "gesher-tls"
	DefaultService   = "gesher"
)

var (
	Namespace = flag.String("namespace", DefaultNamespace, "kubernetes namespace of gesher pod")
	TlsSecret = flag.String("tls-secret", DefaultTlsSecret, "secret to fetch and store tls files from")
	Service   = flag.String("service-name", DefaultService, "service name to use for gesher")
)
