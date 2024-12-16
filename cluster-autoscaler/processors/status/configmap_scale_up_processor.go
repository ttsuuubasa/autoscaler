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
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
)

type ConfigmapScaleUpStatusProcessor struct{}

// ConfigmapScaleUpStatus represents ScaleUpStatus written to Configmap
type ConfigmapScaleUpStatus struct {
	// Result of ScaleUpStatus
	Result string `json:"result" yaml:"result"`
	// NodeGroups contains reasons why scaleup is not executed with each node groups
	NodeGroups []NodeGroupStatus `json:"nodeGroups" yaml:"nodeGroups"`
}

// NodeGroupStatus contains status of every node group's scaleup status
type NodeGroupStatus struct {
	// Name of the node group
	Name string `json:"name" yaml:"name"`
	// ScaleUp contains information of scaleup status
	ScaleUp NodeGroupScaleUpCondition `json:"scaleUp" yaml:"scaleUp"`
}

// NodeGroupScaleUpCondition conains scaleup status
type NodeGroupScaleUpCondition struct {
	// Reasons contains why scaleup is not executed
	Reasons map[string]int `json:"reasons yaml:"reasons"`
}

func NewConfigmapScaleUpStatusProcessor() *ConfigmapScaleUpStatusProcessor {
	return &ConfigmapScaleUpStatusProcessor{}
}

func (*ConfigmapScaleUpStatusProcessor) Process(context *ca_context.AutoscalingContext, status *ScaleUpStatus) {
	var configmapStatus ConfigmapScaleUpStatus
	var nodeGroupStatus NodeGroupStatus

	configmapStatus.Result = scaleUpResult(status.Result)
	nodeGroupStatus.ScaleUp.Reasons = make(map[string]int)

	for _, unschedulabePods := range status.PodsRemainUnschedulable {
		for nodeGroup, reasons := range unschedulabePods.SkippedNodeGroups {
			nodeGroupStatus.Name = nodeGroup
			//nodeGroupStatus.ScaleUp.Result = scaleUpResult(status.Result)
			nodeGroupStatus.ScaleUp.Reasons[reasons.ReasonsStatus()]++
			configmapStatus.NodeGroups = append(configmapStatus.NodeGroups, nodeGroupStatus)
		}
	}
	WriteStatusConfigMap(context.ClientSet, context.ConfigNamespace, configmapStatus, context.StatusConfigMapName)
}

func (*ConfigmapScaleUpStatusProcessor) CleanUp() {

}

func scaleUpResult(result ScaleUpResult) string {
	switch result {
	case ScaleUpSuccessful:
		return "ScaleUpSuccessfule"
	case ScaleUpError:
		return "ScaleUpError"
	case ScaleUpNoOptionsAvailable:
		return "ScaleUpNoOptionsAvailable"
	case ScaleUpNotNeeded:
		return "ScaleUpNotNeeded"
	case ScaleUpNotTried:
		return "ScaleUpNoTried"
	case ScaleUpInCooldown:
		return "ScaleUpInCooldown"
	default:
		return "NoScaleUpExecuted"
	}
}
