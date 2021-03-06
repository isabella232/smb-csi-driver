package main_test

import (
	local_k8s_cluster "code.cloudfoundry.org/smb-volume-k8s-local-cluster"
	"github.com/onsi/gomega/gexec"
	"k8s.io/kubernetes/test/e2e/framework"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)
var smbCsiDriverPath string
var session *gexec.Session
var smbBrokerCompiledPath string
var kubeConfigPath string
var nodeName string

func init() {
	contextType := &framework.TestContext
	contextType.KubeConfig = "/tmp/csi-kubeconfig"
	contextType.KubectlPath = "/usr/local/bin/kubectl"
	framework.AfterReadingAllFlags(contextType)
}

func TestTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Suite")
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(10 * time.Minute)
	nodeName = "default-smb-csi-driver-test-node"
	kubeConfigPath = "/tmp/csi-kubeconfig"
	namespace := "cf-smb"

	local_k8s_cluster.CreateK8sCluster(nodeName, kubeConfigPath, os.Getenv("K8S_IMAGE"))

	useKustomize := os.Getenv("USE_KUSTOMIZE")
	if useKustomize == "true" {
		local_k8s_cluster.Kubectl("apply", "--kustomize", "./deploy/overlays/test")
	} else {
		local_k8s_cluster.Kubectl("create", "namespace", namespace)
		kubectlStdOut := local_k8s_cluster.YttStdout("-f", "./ytt/base", "-f", "ytt/test.yaml")
		local_k8s_cluster.KappWithStringAsStdIn("-y", "deploy", "-a", "smb-csi-driver", "-f")(kubectlStdOut)
	}

	Eventually(func()string{
		return local_k8s_cluster.Kubectl("get", "pod", "-l", "app=csi-nodeplugin-smbplugin", "-n", namespace)
	}, 10 * time.Minute, 1 * time.Second).Should(ContainSubstring("Running"))

	By("pulling the image into the docker daemon", func() {
		local_k8s_cluster.Docker("pull", "localhost:5000/cfpersi/smb-csi-driver:local-test")
	})

	var imageDestination string
	var found bool
	if imageDestination, found = os.LookupEnv("IMAGE_DESTINATION"); !found {
		imageDestination = "/tmp/smb-csi-driver.tgz"
	}

	local_k8s_cluster.Docker("save", "localhost:5000/cfpersi/smb-csi-driver:local-test", "-o", imageDestination)
})

var _ = AfterSuite(func() {
	local_k8s_cluster.DeleteK8sCluster(nodeName, kubeConfigPath)
})