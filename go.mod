module github.com/aneeshkp/operator-cnf-test-operator

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/operator-framework/api v0.3.12
	//github.com/operator-framework/operator-lifecycle-manager
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20200903182547-fddbf04ca175
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v0.18.2
	sigs.k8s.io/controller-runtime v0.6.0
)

replace vbom.ml/util => github.com/fvbommel/util v0.0.0-20180919145318-efcd4e0f9787
