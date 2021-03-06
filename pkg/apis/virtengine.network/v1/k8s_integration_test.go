// +build k8s_integration

package v1_test

import (
	"context"
	"fmt"
	"math/rand"
	"path"
	"testing"

	virtenginev1 "github.com/virtengine/virtengine/pkg/apis/virtengine.network/v1"
	virtengineclient "github.com/virtengine/virtengine/pkg/client/clientset/versioned"
	"github.com/virtengine/virtengine/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func TestWriteRead(t *testing.T) {
	ctx := context.Background()

	withNamespace(ctx, t, func(kcfg *rest.Config, ns string) {
		client, err := virtengineclient.NewForConfig(kcfg)
		require.NoError(t, err)

		for i, spec := range testutil.ManifestGenerators {

			// ensure decode(encode(obj)) == obj

			var (
				lid   = testutil.LeaseID(t)
				group = spec.Generator.Group(t)
				name  = fmt.Sprintf("foo-%v", i)
			)

			// create local k8s manifest object
			kmani, err := virtenginev1.NewManifest(name, lid, &group)

			require.NoError(t, err, spec.Name)

			// save to k8s
			obj, err := client.VirtEngineV1().Manifests(ns).Create(ctx, kmani, metav1.CreateOptions{})
			require.NoError(t, err, spec.Name)

			// ensure created CRD has correct name
			assert.Equal(t, name, obj.GetName(), spec.Name)

			// convert to virtengine-native objects and ensure no data corruption
			deployment, err := obj.Deployment()
			require.NoError(t, err, spec.Name)

			assert.Equal(t, lid, deployment.LeaseID(), spec.Name)
			assert.Equal(t, group, deployment.ManifestGroup(), spec.Name)
		}
	})
}

func withNamespace(ctx context.Context, t testing.TB, fn func(*rest.Config, string)) {
	kcfg := kubeConfig(t)

	kc, err := kubernetes.NewForConfig(kcfg)
	require.NoError(t, err)

	// create a namespace with a random name and a searchable label.
	nsname := fmt.Sprintf("virtengine-test-ns-%v", rand.Uint64())
	_, err = kc.CoreV1().Namespaces().Create(ctx, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nsname,
			Labels: map[string]string{
				"virtengine.network/integration-test": "true",
			},
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	defer func() {
		// delete namespace
		err = kc.CoreV1().Namespaces().Delete(ctx, nsname, metav1.DeleteOptions{})
		require.NoError(t, err)
	}()

	// invoke callback
	fn(kcfg, nsname)

}

func kubeConfig(t testing.TB) *rest.Config {
	t.Helper()
	cfgpath := path.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", cfgpath)
	require.NoError(t, err)
	return config
}
