package daemon

import (
	"context"
	"testing"
	"time"

	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	//+kubebuilder:scaffold:imports
	configv1 "github.com/openshift/dpu-operator/api/v1"
	"github.com/openshift/dpu-operator/internal/testutils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg      *rest.Config
	testEnv  *envtest.Environment
	timeout  = 10 * time.Second
	interval = 1 * time.Second
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(g.Fail)

	g.RunSpecs(t, "e2e tests")
}

var _ = g.BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(g.GinkgoWriter), zap.UseDevMode(true)))
})

var _ = g.AfterSuite(func() {
	// Nothing needed
})

func getPod(c client.Client, name, namespace string) *corev1.Pod {
	podList := &corev1.PodList{}
	err := c.List(context.TODO(), podList, client.InNamespace(namespace))
	Expect(err).NotTo(HaveOccurred())
	for _, pod := range podList.Items {
		if pod.Name == name {
			return &pod
		}
	}
	return nil
}

var _ = g.Describe("Dpu side", g.Ordered, func() {
	var (
		err           error
		dpuSideClient client.Client
	)

	g.BeforeEach(func() {
		cluster := testutils.CdaCluster{
			Name:           "",
			HostConfigPath: "hack/cluster-configs/config-dpu-host.yaml",
			DpuConfigPath:  "hack/cluster-configs/config-dpu.yaml",
		}
		_, dpuConfig, _ := cluster.EnsureExists()
		Expect(err).NotTo(HaveOccurred())

		scheme := runtime.NewScheme()
		utilruntime.Must(configv1.AddToScheme(scheme))
		utilruntime.Must(clientgoscheme.AddToScheme(scheme))

		dpuSideClient, err = client.New(dpuConfig, client.Options{Scheme: scheme})
		Expect(err).NotTo(HaveOccurred())
	})

	g.AfterEach(func() {
	})

	g.It("Should create a pod when creating an SFC and there is no pod", func() {
		nfName := "example-nf"
		nfImage := "example-nf-image-url"
		ns := "openshift-dpu-operator"

		Eventually(func() bool {
			return getPod(dpuSideClient, nfName, ns) != nil
		}, timeout, interval).Should(BeFalse())

		sfc := &configv1.ServiceFunctionChain{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sfc-test",
				Namespace: ns,
			},
			Spec: configv1.ServiceFunctionChainSpec{
				NetworkFunctions: []configv1.NetworkFunction{
					{
						Name:  nfName,
						Image: nfImage,
					},
				},
			},
		}
		err := dpuSideClient.Create(context.TODO(), sfc)
		Expect(err).NotTo(HaveOccurred())

		podList := &corev1.PodList{}
		err = dpuSideClient.List(context.TODO(), podList, client.InNamespace(ns))
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() bool {
			pod := getPod(dpuSideClient, nfName, ns)
			if pod != nil {
				println("Pod got")
				return pod.Spec.Containers[0].Image == nfImage
			}
			println("pod not there")
			return false
		}, timeout, interval).Should(BeTrue())

		err = dpuSideClient.Delete(context.TODO(), sfc)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() bool {
			return getPod(dpuSideClient, nfName, ns) != nil
		}, timeout, interval).Should(BeFalse())
	})
})