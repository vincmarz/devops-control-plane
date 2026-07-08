package main

import (
	"testing"

	kubernetesadapter "github.com/vincmarz/devops-control-plane/internal/adapters/kubernetes"
	"github.com/vincmarz/devops-control-plane/internal/app"
)

// Compile-time assertion: *kubernetesadapter.Client must satisfy the
// app.KubernetesSecretGetter interface. If the Kubernetes adapter GetSecret
// signature drifts, this line will fail to compile and prevent a regression
// before the composition root wires the real Kubernetes secret value loader.
var _ app.KubernetesSecretGetter = (*kubernetesadapter.Client)(nil)

// TestKubernetesClientSatisfiesSecretGetterContract is an explicit runtime
// witness of the compile-time assertion above, so that go test surfaces the
// contract as a documented, executable expectation.
func TestKubernetesClientSatisfiesSecretGetterContract(t *testing.T) {
	var getter app.KubernetesSecretGetter = (*kubernetesadapter.Client)(nil)
	_ = getter
}
