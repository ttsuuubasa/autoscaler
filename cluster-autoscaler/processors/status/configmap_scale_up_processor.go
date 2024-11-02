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
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/context"
)

type ConfigmapScaleUpStatusProcessor struct{}

type ConfigmapScaleUpStatus struct {
	NodeGroups []NodeGroupStatus
}

type NodeGroupStatus struct {
	Name string

	ScaleUp NodeGroupScaleUpCondition
}

type NodeGroupScaleUpCondition struct {
	Result ScaleUpResult

	Reasons map[string]int
}

func (*ConfigmapScaleUpStatusProcessor) Process(context *context.AutoscalingContext, status *ScaleUpStatus) {

	var configmapStatus ConfigmapScaleUpStatus
	var nodeGroupStatus NodeGroupStatus

	for _, unschedulabePods := range status.PodsRemainUnschedulable {
		for nodeGroup, reasons := range unschedulabePods.SkippedNodeGroups {

			nodeGroupStatus.Name = nodeGroup

			nodeGroupStatus.ScaleUp.Reasons[reasons.ReasonsStatus()]++

			configmapStatus.NodeGroups = append(configmapStatus.NodeGroups, nodeGroupStatus)

		}
	}

	utils.WriteStatusConfigMap(context.ClientSet, context.ConfigNamespace,
		configmapStatus, context.LogRecorder, context.StatusConfigMapName, time.Now())

}

func (*ConfigmapScaleUpStatusProcessor) CleanUp() {

}
