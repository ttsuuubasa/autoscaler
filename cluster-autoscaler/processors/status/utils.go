/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package status

import (
	"context"

	"gopkg.in/yaml.v2"
	apiv1 "k8s.io/api/core/v1"
	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "k8s.io/client-go/kubernetes"
	klog "k8s.io/klog/v2"
)

func WriteStatusConfigMap(kubeClient kube_client.Interface, namespace string, status interface{}, caConfigMapName string) {
	var configMap *apiv1.ConfigMap
	var getStatusError, writeStatusError error

	scaleUpConfigMapName := caConfigMapName + "-scaleup"
	maps := kubeClient.CoreV1().ConfigMaps(namespace)
	configMap, getStatusError = maps.Get(context.TODO(), scaleUpConfigMapName, metav1.GetOptions{})
	statusYaml, err := yaml.Marshal(status)
	if err != nil {
		klog.Error(err)
	}
	statusMsg := string(statusYaml)
	if getStatusError == nil {
		if configMap.Data == nil {
			configMap.Data = make(map[string]string)
		}
		configMap.Data["status"] = statusMsg
		_, writeStatusError = maps.Update(context.TODO(), configMap, metav1.UpdateOptions{})
	} else if kube_errors.IsNotFound(getStatusError) {
		configMap = &apiv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      scaleUpConfigMapName,
			},
			Data: map[string]string{
				"status": statusMsg,
			},
		}
		_, writeStatusError = maps.Create(context.TODO(), configMap, metav1.CreateOptions{})
	} else {
		klog.Errorf("Failed to retriece status configmap for update: %v", getStatusError)
	}
	if writeStatusError != nil {
		klog.Errorf("Failed to write status configmap: %v", writeStatusError)
	}
	klog.V(4).Infof("Successfully wrote status configmap with name \"%v\"", scaleUpConfigMapName)
}
