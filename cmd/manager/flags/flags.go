/*
Copyright 2020 Redis Labs Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package flags

import (
	"flag"
)

const (
	DefaultNamespace = "default"
	DefaultTlsSecret = "gesher-tls"
	DefaultService   = "gesher"
	DefaultHttpsPort = 8443
)

var (
	Namespace = flag.String("namespace", DefaultNamespace, "kubernetes namespace of gesher pod")
	TlsSecret = flag.String("tls-secret", DefaultTlsSecret, "secret to fetch and store tls files from")
	Service   = flag.String("service-name", DefaultService, "service name to use for gesher")
	Port      = flag.Int("port", DefaultHttpsPort, "port https server should run on")
)
