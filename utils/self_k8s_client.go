// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package utils offer the some utils for certificate handling  and k8s config
package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"huawei.com/npu-exporter/hwlog"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const prefix = "/etc/mindx-dl/"

var (
	k8sClientOnce sync.Once
	kubeClientSet *kubernetes.Clientset
)

// K8sClient Get the internal k8s client of the cluster
func K8sClient(kubeconfig string) (*kubernetes.Clientset, error) {
	k8sClientOnce.Do(func() {
		if kubeconfig == "" {
			configPath := os.Getenv("KUBECONFIG")
			if len(configPath) > maxLen {
				hwlog.RunLog.Error("the path is too long")
			}
			kubeconfig = configPath
		}
		path, err := CheckPath(kubeconfig)
		if err != nil {
			hwlog.RunLog.Error(err)
		}
		config, err := BuildConfigFromFlags("", path)
		if err != nil {
			hwlog.RunLog.Error(err)
			return
		}
		// Create a new k8sClientSet based on the specified config using the current context
		kubeClientSet, err = kubernetes.NewForConfig(config)
		if err != nil {
			hwlog.RunLog.Error(err)
		}
	})
	if kubeClientSet == nil {
		return nil, errors.New("get k8s client failed")
	}

	return kubeClientSet, nil
}

// K8sClientFor  add a default path for each component of MindXDL
// component name is noded,task-manager,hccl-controller  etc.
func K8sClientFor(kubeConfig, component string) (*kubernetes.Clientset, error) {
	// if kubeConfig not set, check and use default path
	kubeConf := prefix + component + "/" + KubeCfgFile
	if kubeConfig == "" && component != "" && IsExists(kubeConf) {
		return K8sClient(kubeConf)
	}
	// use custom path
	return K8sClient(kubeConfig)
}

// SelfClientConfigLoadingRules  extend   clientcmd.ClientConfigLoadingRules
type SelfClientConfigLoadingRules struct {
	clientcmd.ClientConfigLoadingRules
}

// Load  override the clientcmd.ClientConfigLoadingRules Load method
func (rules *SelfClientConfigLoadingRules) Load() (*clientcmdapi.Config, error) {
	errlist := []error{}
	if len(rules.ExplicitPath) == 0 {
		return nil, errors.New("no ExplicitPath set")
	}
	config, err := LoadFromFile(rules.ExplicitPath)
	if err != nil {
		errlist = append(errlist, fmt.Errorf("error loading config file \"%s\": %v",
			MaskPrefix(rules.ExplicitPath), err))
	}
	return config, utilerrors.NewAggregate(errlist)
}

// LoadFromFile takes a filename and deserializes the contents into Config object
func LoadFromFile(filename string) (*clientcmdapi.Config, error) {
	kubeconfigBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(kubeconfigBytes, []byte("apiVersion:")) {
		KmcInit(Aes256gcm, "", "")
		hwlog.RunLog.Info("start to decrypt cfg")
		kubeconfigBytes, err = Decrypt(0, kubeconfigBytes)
		if err != nil {
			return nil, err
		}
	}
	cfg, err := clientcmd.Load(kubeconfigBytes)
	if err != nil {
		return nil, err
	}
	hwlog.RunLog.Infof("Config loaded from file: %s", MaskPrefix(filename))

	for key, val := range cfg.AuthInfos {
		val.LocationOfOrigin = filename
		cfg.AuthInfos[key] = val
	}
	for key, val := range cfg.Clusters {
		val.LocationOfOrigin = filename
		cfg.Clusters[key] = val
	}
	for key, val := range cfg.Contexts {
		val.LocationOfOrigin = filename
		cfg.Contexts[key] = val
	}

	if cfg.AuthInfos == nil {
		cfg.AuthInfos = map[string]*clientcmdapi.AuthInfo{}
	}
	if cfg.Clusters == nil {
		cfg.Clusters = map[string]*clientcmdapi.Cluster{}
	}
	if cfg.Contexts == nil {
		cfg.Contexts = map[string]*clientcmdapi.Context{}
	}
	return cfg, nil
}

// BuildConfigFromFlags local implement of k8s client buildConfig
func BuildConfigFromFlags(masterURL, confPath string) (*restclient.Config, error) {
	if confPath == "" && masterURL == "" {
		hwlog.RunLog.Warnf("Neither --kubeconfig nor --master was specified." +
			"Using the inClusterConfig.  This might not work.")
		kubeconf, err := restclient.InClusterConfig()
		if err == nil {
			return kubeconf, nil
		}
		hwlog.RunLog.Warn("error creating inClusterConfig")
	}
	cliRule := clientcmd.ClientConfigLoadingRules{ExplicitPath: confPath}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&SelfClientConfigLoadingRules{cliRule},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: masterURL}}).ClientConfig()
}
