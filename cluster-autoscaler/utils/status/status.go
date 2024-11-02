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
	"errors"
	"fmt"
	"time"

	"gopkg.in/yaml.v2"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type ConfigMapStatus interface {
	Time(string)
}

// WriteStatusConfigMap writes updates status ConfigMap with a given message or creates a new
// ConfigMap if it doesn't exist. If logRecorder is passed and configmap update is successful
// logRecorder's internal reference will be updated.
func WriteStatusConfigMap(kubeClient kube_client.Interface, namespace string, status ConfigMapStatus, logRecorder *LogEventRecorder, statusConfigMapName string, currentTime time.Time) (*apiv1.ConfigMap, error) {
	statusUpdateTime := currentTime.Format(ConfigMapLastUpdateFormat)
	status.Time = statusUpdateTime
	var configMap *apiv1.ConfigMap
	var getStatusError, writeStatusError error
	var errMsg string
	maps := kubeClient.CoreV1().ConfigMaps(namespace)
	configMap, getStatusError = maps.Get(context.TODO(), statusConfigMapName, metav1.GetOptions{})
	statusYaml, err := yaml.Marshal(status)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal status configmap: %v", err)
	}
	statusMsg := string(statusYaml)
	if getStatusError == nil {
		if configMap.Data == nil {
			configMap.Data = make(map[string]string)
		}
		configMap.Data["status"] = statusMsg
		if configMap.ObjectMeta.Annotations == nil {
			configMap.ObjectMeta.Annotations = make(map[string]string)
		}
		configMap.ObjectMeta.Annotations[ConfigMapLastUpdatedKey] = statusUpdateTime
		configMap, writeStatusError = maps.Update(context.TODO(), configMap, metav1.UpdateOptions{})
	} else if kube_errors.IsNotFound(getStatusError) {
		configMap = &apiv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      statusConfigMapName,
				Annotations: map[string]string{
					ConfigMapLastUpdatedKey: statusUpdateTime,
				},
			},
			Data: map[string]string{
				"status": statusMsg,
			},
		}
		configMap, writeStatusError = maps.Create(context.TODO(), configMap, metav1.CreateOptions{})
	} else {
		errMsg = fmt.Sprintf("Failed to retrieve status configmap for update: %v", getStatusError)
	}
	if writeStatusError != nil {
		errMsg = fmt.Sprintf("Failed to write status configmap: %v", writeStatusError)
	}
	if errMsg != "" {
		klog.Error(errMsg)
		return nil, errors.New(errMsg)
	}
	klog.V(8).Infof("Successfully wrote status configmap with body \"%v\"", statusMsg)
	// Having this as a side-effect is somewhat ugly
	// But it makes error handling easier, as we get a free retry each loop
	if logRecorder != nil {
		logRecorder.statusObject = configMap
	}
	return configMap, nil
}
