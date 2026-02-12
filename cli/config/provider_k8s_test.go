package config

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestK8sConfigMapProviderRead(t *testing.T) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
		},
		Data: map[string]string{
			"config.yaml": "key: value",
			"app.json":    `{"port": 8080}`,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/namespaces/default/configmaps/test-config" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cm)
	}))
	defer server.Close()

	provider, err := NewK8sConfigMapProvider(K8sConfigMapConfig{
		Namespace: "default",
		ConfigMap: "test-config",
		APIServer: server.URL,
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := provider.Read("config", "yaml")
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "key: value" {
		t.Errorf("got %q, want %q", string(data), "key: value")
	}
}

func TestK8sConfigMapProviderReadWithKey(t *testing.T) {
	cm := &corev1.ConfigMap{
		Data: map[string]string{
			"mykey": "myvalue",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cm)
	}))
	defer server.Close()

	provider, err := NewK8sConfigMapProvider(K8sConfigMapConfig{
		Namespace: "default",
		ConfigMap: "test-config",
		Key:       "mykey",
		APIServer: server.URL,
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := provider.Read("ignored")
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "myvalue" {
		t.Errorf("got %q, want %q", string(data), "myvalue")
	}
}

func TestK8sConfigMapProviderReadBinaryData(t *testing.T) {
	cm := &corev1.ConfigMap{
		BinaryData: map[string][]byte{
			"cert.pem": []byte("binary-data"),
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cm)
	}))
	defer server.Close()

	provider, err := NewK8sConfigMapProvider(K8sConfigMapConfig{
		Namespace: "default",
		ConfigMap: "test-config",
		Key:       "cert.pem",
		APIServer: server.URL,
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := provider.Read("ignored")
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "binary-data" {
		t.Errorf("got %q, want %q", string(data), "binary-data")
	}
}

func TestK8sConfigMapProviderReadKeyNotFound(t *testing.T) {
	cm := &corev1.ConfigMap{
		Data: map[string]string{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cm)
	}))
	defer server.Close()

	provider, err := NewK8sConfigMapProvider(K8sConfigMapConfig{
		Namespace: "default",
		ConfigMap: "test-config",
		Key:       "missing",
		APIServer: server.URL,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = provider.Read("ignored")
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestK8sConfigMapProviderWriteReadOnly(t *testing.T) {
	provider, err := NewK8sConfigMapProvider(K8sConfigMapConfig{
		Namespace: "default",
		ConfigMap: "test-config",
		APIServer: "http://localhost:8080",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = provider.Write("config", []byte("data"))
	if err == nil {
		t.Error("expected error for read-only provider")
	}
}

func TestK8sConfigMapProviderString(t *testing.T) {
	provider, err := NewK8sConfigMapProvider(K8sConfigMapConfig{
		Namespace: "mynamespace",
		ConfigMap: "myconfig",
		APIServer: "http://localhost:8080",
	})
	if err != nil {
		t.Fatal(err)
	}
	s := provider.String()
	if s != "k8s://mynamespace/myconfig" {
		t.Errorf("got %q, want %q", s, "k8s://mynamespace/myconfig")
	}
}

func TestNewK8sConfigMapProviderRequiresConfigMap(t *testing.T) {
	_, err := NewK8sConfigMapProvider(K8sConfigMapConfig{
		Namespace: "default",
	})
	if err == nil {
		t.Error("expected error for missing configmap name")
	}
}

func TestK8sProviderCreator(t *testing.T) {
	creator := providerCreators["k8s"]
	if creator == nil {
		t.Fatal("k8s provider creator not registered")
	}

	_, err := creator("default/myconfig")
	if err != nil {
		t.Fatal(err)
	}

	_, err = creator("default/myconfig/mykey")
	if err != nil {
		t.Fatal(err)
	}

	_, err = creator("invalid")
	if err == nil {
		t.Error("expected error for invalid config")
	}
}
