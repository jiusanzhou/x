package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
)

type k8sConfigMapProvider struct {
	namespace  string
	configMap  string
	key        string
	kubeconfig string
	token      string
	apiServer  string
	client     *http.Client
	mu         sync.RWMutex
}

type K8sConfigMapConfig struct {
	Namespace  string
	ConfigMap  string
	Key        string
	Kubeconfig string
	Token      string
	APIServer  string
}

func (k *k8sConfigMapProvider) Read(name string, typs ...string) ([]byte, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	cm, err := k.getConfigMap()
	if err != nil {
		return nil, fmt.Errorf("get configmap: %w", err)
	}

	key := k.key
	if key == "" {
		key = name
		if len(typs) > 0 && !strings.HasSuffix(key, "."+typs[0]) {
			key = key + "." + typs[0]
		}
	}

	data, ok := cm.Data[key]
	if !ok {
		if binaryData, ok := cm.BinaryData[key]; ok {
			return binaryData, nil
		}
		return nil, fmt.Errorf("key '%s' not found in configmap", key)
	}

	return []byte(data), nil
}

func (k *k8sConfigMapProvider) Write(name string, data []byte, typs ...string) error {
	return errors.New("kubernetes configmap provider is read-only")
}

func (k *k8sConfigMapProvider) Watch(name string, typs ...string) (Watcher, error) {
	key := k.key
	if key == "" {
		key = name
		if len(typs) > 0 && !strings.HasSuffix(key, "."+typs[0]) {
			key = key + "." + typs[0]
		}
	}

	return &k8sWatcher{
		provider: k,
		name:     name,
		key:      key,
		types:    typs,
		exit:     make(chan bool),
	}, nil
}

func (k *k8sConfigMapProvider) String() string {
	return fmt.Sprintf("k8s://%s/%s", k.namespace, k.configMap)
}

func (k *k8sConfigMapProvider) getConfigMap() (*corev1.ConfigMap, error) {
	apiServer := k.apiServer
	if apiServer == "" {
		apiServer = os.Getenv("KUBERNETES_SERVICE_HOST")
		port := os.Getenv("KUBERNETES_SERVICE_PORT")
		if apiServer != "" && port != "" {
			apiServer = fmt.Sprintf("https://%s:%s", apiServer, port)
		}
	}

	if apiServer == "" {
		return nil, errors.New("kubernetes API server not configured")
	}

	token := k.token
	if token == "" {
		tokenPath := "/var/run/secrets/kubernetes.io/serviceaccount/token"
		if data, err := os.ReadFile(tokenPath); err == nil {
			token = string(data)
		}
	}

	url := fmt.Sprintf("%s/api/v1/namespaces/%s/configmaps/%s",
		strings.TrimSuffix(apiServer, "/"),
		k.namespace,
		k.configMap,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := k.client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var cm corev1.ConfigMap
	if err := json.NewDecoder(resp.Body).Decode(&cm); err != nil {
		return nil, err
	}

	return &cm, nil
}

type k8sWatcher struct {
	provider *k8sConfigMapProvider
	name     string
	key      string
	types    []string
	exit     chan bool
	lastData []byte
}

func (w *k8sWatcher) Next() (*ChangeSet, error) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.exit:
			return nil, errors.New("watcher closed")
		case <-ticker.C:
			data, err := w.provider.Read(w.name, w.types...)
			if err != nil {
				continue
			}

			if string(data) != string(w.lastData) {
				w.lastData = data
				cs := &ChangeSet{
					Data:      data,
					Timestamp: time.Now(),
					Source:    w.provider.String(),
				}
				cs.Checksum = cs.Sum()
				return cs, nil
			}
		}
	}
}

func (w *k8sWatcher) Stop() error {
	select {
	case <-w.exit:
	default:
		close(w.exit)
	}
	return nil
}

func NewK8sConfigMapProvider(cfg K8sConfigMapConfig) (Provider, error) {
	if cfg.Namespace == "" {
		nsPath := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
		if data, err := os.ReadFile(nsPath); err == nil {
			cfg.Namespace = string(data)
		} else {
			cfg.Namespace = "default"
		}
	}

	if cfg.ConfigMap == "" {
		return nil, errors.New("configmap name is required")
	}

	if cfg.Kubeconfig != "" {
		expanded := cfg.Kubeconfig
		if strings.HasPrefix(expanded, "~") {
			home, _ := os.UserHomeDir()
			expanded = filepath.Join(home, expanded[1:])
		}
		if _, err := os.Stat(expanded); err == nil {
			cfg.Kubeconfig = expanded
		}
	}

	return &k8sConfigMapProvider{
		namespace:  cfg.Namespace,
		configMap:  cfg.ConfigMap,
		key:        cfg.Key,
		kubeconfig: cfg.Kubeconfig,
		token:      cfg.Token,
		apiServer:  cfg.APIServer,
	}, nil
}

func WithConfigMap(namespace, name string, opts ...func(*K8sConfigMapConfig)) Option {
	return func(c *Options) {
		cfg := K8sConfigMapConfig{
			Namespace: namespace,
			ConfigMap: name,
		}
		for _, opt := range opts {
			opt(&cfg)
		}
		provider, err := NewK8sConfigMapProvider(cfg)
		if err != nil {
			return
		}
		c.providers = append(c.providers, provider)
	}
}

func WithConfigMapKey(key string) func(*K8sConfigMapConfig) {
	return func(c *K8sConfigMapConfig) {
		c.Key = key
	}
}

func WithK8sAPIServer(server string) func(*K8sConfigMapConfig) {
	return func(c *K8sConfigMapConfig) {
		c.APIServer = server
	}
}

func WithK8sToken(token string) func(*K8sConfigMapConfig) {
	return func(c *K8sConfigMapConfig) {
		c.Token = token
	}
}

func init() {
	RegisterProviderCreator("k8s", func(c interface{}) (Provider, error) {
		switch cfg := c.(type) {
		case K8sConfigMapConfig:
			return NewK8sConfigMapProvider(cfg)
		case string:
			parts := strings.Split(cfg, "/")
			if len(parts) < 2 {
				return nil, errors.New("k8s provider config must be namespace/configmap or namespace/configmap/key")
			}
			k8sCfg := K8sConfigMapConfig{
				Namespace: parts[0],
				ConfigMap: parts[1],
			}
			if len(parts) > 2 {
				k8sCfg.Key = parts[2]
			}
			return NewK8sConfigMapProvider(k8sCfg)
		default:
			return nil, errors.New("invalid k8s provider config")
		}
	})

	RegisterProviderCreator("configmap", func(c interface{}) (Provider, error) {
		return NewProvider("k8s", c)
	})
}
