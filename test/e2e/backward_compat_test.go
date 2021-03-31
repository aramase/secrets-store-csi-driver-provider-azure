// +build e2e

package e2e

import (
	. "github.com/onsi/ginkgo"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/secrets-store-csi-driver/apis/v1alpha1"

	"github.com/Azure/secrets-store-csi-driver-provider-azure/test/e2e/framework/namespace"
)

var _ = Describe("When upgrading Secrets Store CSI Driver and AKV Provider", func() {
	var (
		specName = "backward-compat"
		spc      *v1alpha1.SecretProviderClass
		ns       *corev1.Namespace
		p        *corev1.Pod
	)

	BeforeEach(func() {
		ns = namespace.Create(namespace.CreateInput{
			Creator: kubeClient,
			Name:    specName,
		})
	})

	AfterEach(func() {
		Cleanup(CleanupInput{
			Namespace: ns,
			Getter:    kubeClient,
			Lister:    kubeClient,
			Deleter:   kubeClient,
		})
	})

	It("should be backward compatible with old and new version", func() {

	})
})
