package k8s

import (
	"fmt"
	"sync"
	"time"

	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"yager/pkg/logger"
)

type _client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

var (
	clients = make(map[string]*_client)
	mutex   = &sync.Mutex{}
)

// GetClientByAzCode fetch k8s client by zone code
func GetClientByAzCode(zoneName string) (*kubernetes.Clientset, error) {
	log := logger.RuntimeLog

	host := viper.GetString(fmt.Sprintf("k8s.%s.host", zoneName))
	token := viper.GetString(fmt.Sprintf("k8s.%s.token", zoneName))

	config := &rest.Config{
		Host:        host,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
		Timeout: time.Second * time.Duration(viper.GetInt(fmt.Sprintf("k8s.%s.timeout", zoneName))),
	}

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.WithError(err).Errorf("Failed to get k8s client with host:%s, token:%s", host, token)
		return nil, err
	}

	mutex.Lock()
	defer mutex.Unlock()

	clients[host] = &_client{
		clientset: cs,
		config:    config,
	}

	log.Infof("Use k8s cluster: %s", config.Host)
	return clients[host].clientset, nil
}

func int32Ptr(i int32) *int32 { return &i }

func int64Ptr(i int64) *int64 { return &i }
