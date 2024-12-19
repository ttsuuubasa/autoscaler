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

package nodeinfosprovider

import (
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	caerror "k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
)

const (
	NodeHasFabricDeviceKey = "composable.fabric.dra"
)

// FabricDRANodeInfoProvider is a wrapper for MixedTemplateNodeInfoProvider.
type FabricDRANodeInfoProvider struct {
	templateNodeInfoProvider TemplateNodeInfoProvider
}

// NewFabricNodeInfoProvider returns FabricDRANodeInfoProvider wrapping MixedTemplateNodeInfoProvider.
func NewFabricNodeInfoProvider(t *time.Duration, forceDaemonSets bool) *FabricDRANodeInfoProvider {
	return &FabricDRANodeInfoProvider{
		templateNodeInfoProvider: NewMixedTemplateNodeInfoProvider(t, forceDaemonSets),
	}
}

// Process returns the nodeInfos set for this cluster
func (p *FabricDRANodeInfoProvider) Process(ctx *context.AutoscalingContext, nodes []*apiv1.Node, daemonsets []*appsv1.DaemonSet, taintConfig taints.TaintConfig, currentTime time.Time) (map[string]*framework.NodeInfo, errors.AutoscalerError) {
	nodeInfos, err := p.templateNodeInfoProvider.Process(ctx, nodes, daemonsets, taintConfig, currentTime)
	if err != nil {
		return nil, err
	}

	// Remove Node local ResourceSlices of nodeInfos and labels nodeInfo to be referred from the fabric ResourceSlices.
	result := make(map[string]*framework.NodeInfo)
	for _, node := range nodes {
		nodeGroup, err := ctx.CloudProvider.NodeGroupForNode(node)
		if err != nil {
			return nil, caerror.ToAutoscalerError(caerror.CloudProviderError, err)
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			continue
		}
		id := nodeGroup.Id()
		if _, found := result[id]; !found {
			if _, isFabric := node.Labels[NodeHasFabricDeviceKey]; isFabric {
				nodeInfos[id].LocalResourceSlices = nil
				result[id] = nodeInfos[id]
			} else {
				result[id] = nodeInfos[id]
			}
		}
	}
	return result, nil
}

// CleanUp cleans up processor's internal structures.
func (p *FabricDRANodeInfoProvider) CleanUp() {
}
