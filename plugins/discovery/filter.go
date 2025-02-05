// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"fmt"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"

	"github.com/gobwas/glob"
)

type resourceFilter struct {
	kind       string
	images     glob.Glob
	namespaces glob.Glob
	labels     map[string]glob.Glob
}

func newResourceFilter(selectors discovery.Selectors) (*resourceFilter, error) {
	rf := &resourceFilter{
		images:     filter.Compile(selectors.Images),
		labels:     filter.MultiCompile(selectors.Labels),
		namespaces: filter.Compile(selectors.Namespaces),
	}

	kind, err := resourceType(selectors.ResourceType)
	if err != nil {
		return nil, err
	}

	rf.kind = kind
	if rf.kind != discovery.NodeType.String() && rf.images == nil && rf.labels == nil && rf.namespaces == nil {
		return nil, fmt.Errorf("no selectors specified")
	}
	return rf, nil
}

func resourceType(kind string) (string, error) {
	if kind == "" {
		return discovery.PodType.String(), nil
	}
	switch kind {
	case discovery.PodType.String(), discovery.ServiceType.String(), discovery.NodeType.String():
		return kind, nil
	default:
		return "", fmt.Errorf("invalid resource type: %s", kind)
	}
}

func (r *resourceFilter) matches(resource discovery.Resource) bool {
	if r == nil {
		return false
	}
	if r.kind != resource.Kind {
		return false
	}
	if r.labels != nil && !matchesTags(r.labels, resource.Meta.Labels) {
		return false
	}
	if r.namespaces != nil && !r.namespaces.Match(resource.Meta.Namespace) {
		return false
	}
	if r.images != nil {
		for _, container := range resource.Containers {
			if r.images.Match(container.Image) {
				return true
			}
		}
		return false
	}
	return true
}

func matchesTags(matchers map[string]glob.Glob, tags map[string]string) bool {
	if tags == nil || len(tags) == 0 {
		return false
	}
	for k, matcher := range matchers {
		val, ok := tags[k]
		if !ok || !matcher.Match(val) {
			return false
		}
	}
	return true
}
