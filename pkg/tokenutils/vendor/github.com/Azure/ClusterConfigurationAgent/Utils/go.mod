module github.com/Azure/ClusterConfigurationAgent/Utils

go 1.13

require (
	github.com/Azure/ClusterConfigurationAgent/LogHelper v0.0.0
	github.com/Azure/go-autorest/autorest/adal v0.9.3
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/bugsnag/bugsnag-go v1.7.0 // indirect
	github.com/bugsnag/panicwrap v1.2.0 // indirect
	github.com/containerd/containerd v1.3.2
	github.com/deislabs/oras v0.8.1
	github.com/docker/go-metrics v0.0.0-20181218153428-b84716841b82 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/garyburd/redigo v1.6.0 // indirect
	github.com/go-openapi/strfmt v0.19.3
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/gorilla/handlers v1.4.0 // indirect
	github.com/kardianos/osext v0.0.0-20170510131534-ae77be60afb1 // indirect
	github.com/miekg/dns v0.0.0-20181005163659-0d29b283ac0f // indirect
	github.com/opencontainers/image-spec v1.0.1
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/common v0.7.0
	github.com/spf13/pflag v1.0.5
	github.com/xenolf/lego v0.0.0-20160613233155-a9d8cec0e656 // indirect
	github.com/yvasiyarov/go-metrics v0.0.0-20150112132944-c25f46c4b940 // indirect
	github.com/yvasiyarov/gorelic v0.0.6 // indirect
	golang.org/x/net v0.0.0-20200520004742-59133d7f0dd7
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	gopkg.in/square/go-jose.v1 v1.1.2 // indirect
	gotest.tools v2.2.0+incompatible
	helm.sh/helm/v3 v3.2.0
	k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/cli-runtime v0.18.0
	k8s.io/client-go v0.18.6
	k8s.io/code-generator v0.18.6
	k8s.io/kubectl v0.18.0
	rsc.io/letsencrypt v0.0.1 // indirect
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/Azure/ClusterConfigurationAgent/LogHelper v0.0.0 => ../LogHelper
	github.com/Azure/ClusterConfigurationAgent/Utils v0.0.0 => ../Utils/
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	k8s.io/api => k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver v0.18.0 => k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.6
	k8s.io/client-go => k8s.io/client-go v0.18.6
	rsc.io/letsencrypt => github.com/dmcgowan/letsencrypt v0.0.0-20160928181947-1847a81d2087
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.6.3
	vbom.ml/util => github.com/fvbommel/util v0.0.2
)
