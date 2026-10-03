/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package upgradescanner

import (
	"context"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/version"

	clusterFake "github.com/vmware-tanzu/octant/internal/cluster/fake"
	configFake "github.com/vmware-tanzu/octant/internal/config/fake"
	"github.com/vmware-tanzu/octant/internal/describer"
	"github.com/vmware-tanzu/octant/internal/link"
	linkFake "github.com/vmware-tanzu/octant/internal/link/fake"
	"github.com/vmware-tanzu/octant/pkg/store"
	storeFake "github.com/vmware-tanzu/octant/pkg/store/fake"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

func TestDescriber_PathFilters(t *testing.T) {
	d := NewDescriber()
	filters := d.PathFilters()
	require.Len(t, filters, 3)

	assert.True(t, filters[0].Match("/upgrade-scanner"))
	assert.False(t, filters[0].Match("/upgrade-scanner/1.30"))

	assert.True(t, filters[1].Match("/upgrade-scanner/1.30"))
	assert.False(t, filters[1].Match("/upgrade-scanner/1.30/networking.k8s.io"))

	assert.True(t, filters[2].Match("/upgrade-scanner/1.30/networking.k8s.io"))
	assert.False(t, filters[2].Match("/upgrade-scanner/1.30"))

	fields := filters[1].Fields("/upgrade-scanner/1.30")
	assert.Equal(t, "1.30", fields["target"])

	groupFields := filters[2].Fields("/upgrade-scanner/1.30/networking.k8s.io")
	assert.Equal(t, "1.30", groupFields["target"])
	assert.Equal(t, "networking.k8s.io", groupFields["group"])

	require.NoError(t, d.Reset(context.Background()))
}

func TestResolveTarget(t *testing.T) {
	tests := []struct {
		name           string
		fields         map[string]string
		clusterVersion string
		expected       string
	}{
		{name: "explicit target", fields: map[string]string{"target": "1.30"}, clusterVersion: "1.27", expected: "1.30"},
		{name: "cluster next minor", fields: nil, clusterVersion: "1.27", expected: "1.28"},
		{name: "default", fields: nil, expected: defaultTargetVersion},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, resolveTarget(test.fields, test.clusterVersion))
		})
	}
}

func TestClusterMinorVersion(t *testing.T) {
	got, err := clusterMinorVersion(&version.Info{GitVersion: "v1.27.3"})
	require.NoError(t, err)
	assert.Equal(t, "1.27", got)

	got, err = clusterMinorVersion(&version.Info{Major: "1", Minor: "28+"})
	require.NoError(t, err)
	assert.Equal(t, "1.28", got)

	_, err = clusterMinorVersion(nil)
	require.Error(t, err)
}

func TestTargetVersions(t *testing.T) {
	versions := targetVersions("1.30")
	require.NotEmpty(t, versions)
	assert.Equal(t, "1.30", versions[0])
	// The range must extend past the newest known removal (1.32).
	assert.Contains(t, versions, "1.32")
}

func TestDistinctGroups(t *testing.T) {
	groups := distinctGroups([]Finding{
		{Group: "batch"},
		{Group: "networking.k8s.io"},
		{Group: "batch"},
		{Group: ""},
	})
	assert.Equal(t, []string{"batch", "networking.k8s.io"}, groups)
}

func TestDescriber_Describe(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	discoveryClient := clusterFake.NewMockDiscoveryInterface(controller)
	discoveryClient.EXPECT().ServerVersion().Return(&version.Info{GitVersion: "v1.27.3"}, nil).AnyTimes()
	discoveryClient.EXPECT().
		ServerGroupsAndResources().
		Return(nil, []*metav1.APIResourceList{cronJobResourceList()}, nil).
		AnyTimes()

	clusterClient := clusterFake.NewMockClientInterface(controller)
	clusterClient.EXPECT().DiscoveryClient().Return(discoveryClient, nil).AnyTimes()

	objectStore := storeFake.NewMockStore(controller)
	objectStore.EXPECT().List(gomock.Any(), gomock.Any()).Return(cronJobList("default"), false, nil).AnyTimes()

	dash := configFake.NewMockDash(controller)
	dash.EXPECT().ClusterClient().Return(clusterClient).AnyTimes()
	dash.EXPECT().ObjectStore().Return(objectStore).AnyTimes()

	l := linkFake.NewMockInterface(controller)
	l.EXPECT().ForGVK("default", "batch/v1", "CronJob", "cron-default", "cron-default").
		Return(component.NewLink("", "cron-default", "/cluster-overview"), nil).
		AnyTimes()

	options := describer.Options{
		Dash:   dash,
		Link:   l,
		Fields: map[string]string{"target": "1.25"},
	}

	response, err := NewDescriber().Describe(context.Background(), "", options)
	require.NoError(t, err)
	require.Len(t, response.Components, 1)

	list, ok := response.Components[0].(*component.List)
	require.True(t, ok)
	require.Len(t, list.Config.Items, 4)

	summary, ok := list.Config.Items[0].(*component.Summary)
	require.True(t, ok)
	require.Len(t, summary.Config.Sections, 4)
	assert.Equal(t, "1.27", summary.Config.Sections[0].Content.(*component.Text).Config.Text)
	assert.Equal(t, "1.25", summary.Config.Sections[1].Content.(*component.Text).Config.Text)

	targetPicker, ok := list.Config.Items[1].(*component.FlexLayout)
	require.True(t, ok)
	require.NotEmpty(t, targetPicker.Config.Sections)

	groupPicker, ok := list.Config.Items[2].(*component.FlexLayout)
	require.True(t, ok)
	require.NotEmpty(t, groupPicker.Config.Sections)

	table, ok := list.Config.Items[3].(*component.Table)
	require.True(t, ok)
	require.Len(t, table.Config.Rows, 1)
	row := table.Config.Rows[0]
	assert.Equal(t, "batch/v1beta1", row["API Version"].(*component.Text).Config.Text)
	assert.IsType(t, &component.Link{}, row["Object"])
}

func TestDescriber_GroupFilter(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	storageClassResourceList := &metav1.APIResourceList{
		GroupVersion: "storage.k8s.io/v1beta1",
		APIResources: []metav1.APIResource{
			{Name: "storageclasses", Kind: "StorageClass", Namespaced: false},
		},
	}

	discoveryClient := clusterFake.NewMockDiscoveryInterface(controller)
	discoveryClient.EXPECT().ServerVersion().Return(&version.Info{GitVersion: "v1.27.3"}, nil).AnyTimes()
	discoveryClient.EXPECT().
		ServerGroupsAndResources().
		Return(nil, []*metav1.APIResourceList{cronJobResourceList(), storageClassResourceList}, nil).
		AnyTimes()

	clusterClient := clusterFake.NewMockClientInterface(controller)
	clusterClient.EXPECT().DiscoveryClient().Return(discoveryClient, nil).AnyTimes()

	storageClassList := &unstructured.UnstructuredList{Items: []unstructured.Unstructured{
		{Object: map[string]interface{}{
			"apiVersion": "storage.k8s.io/v1beta1",
			"kind":       "StorageClass",
			"metadata":   map[string]interface{}{"name": "standard"},
		}},
	}}

	objectStore := storeFake.NewMockStore(controller)
	// Register the more specific expectation first so it wins for StorageClass.
	objectStore.EXPECT().
		List(gomock.Any(), store.Key{APIVersion: "storage.k8s.io/v1beta1", Kind: "StorageClass"}).
		Return(storageClassList, false, nil).
		AnyTimes()
	objectStore.EXPECT().List(gomock.Any(), gomock.Any()).Return(cronJobList("default"), false, nil).AnyTimes()

	dash := configFake.NewMockDash(controller)
	dash.EXPECT().ClusterClient().Return(clusterClient).AnyTimes()
	dash.EXPECT().ObjectStore().Return(objectStore).AnyTimes()

	// No link generator: buildTable falls back to plain text.
	options := describer.Options{
		Dash:   dash,
		Fields: map[string]string{"target": "1.25", "group": "batch"},
	}

	response, err := NewDescriber().Describe(context.Background(), "", options)
	require.NoError(t, err)

	list, ok := response.Components[0].(*component.List)
	require.True(t, ok)

	summary := list.Config.Items[0].(*component.Summary)
	assert.Equal(t, "batch/v1beta1: 1", summary.Config.Sections[3].Content.(*component.Text).Config.Text)

	table := list.Config.Items[3].(*component.Table)
	require.Len(t, table.Config.Rows, 1)
	assert.Equal(t, "batch/v1beta1", table.Config.Rows[0]["API Version"].(*component.Text).Config.Text)
}

// fakeLinkConfig mimics the real object path resolution: only the replacement
// GVK (batch/v1 CronJob) is known, everything else resolves to an empty path.
type fakeLinkConfig struct{}

func (fakeLinkConfig) ObjectPath(namespace, apiVersion, kind, name string) (string, error) {
	if apiVersion == "batch/v1" && kind == "CronJob" {
		return path.Join("/overview/namespace", namespace, "workloads/cron-jobs", name), nil
	}
	return "", nil
}

func TestBuildTable_ReplacementLink(t *testing.T) {
	l, err := link.NewFromDashConfig(fakeLinkConfig{})
	require.NoError(t, err)

	options := describer.Options{Link: l}
	findings := []Finding{
		{
			Group: "batch", Version: "v1beta1", Kind: "CronJob",
			RemovedIn: "1.25", Replacement: "batch/v1 CronJob",
			Namespace: "default", Name: "cron",
			ReplaceAPIVersion: "batch/v1", ReplaceKind: "CronJob",
		},
		{
			// Replacement is set but the path cannot be resolved.
			Group: "policy", Version: "v1beta1", Kind: "PodDisruptionBudget",
			RemovedIn: "1.25", Replacement: "policy/v1 PodDisruptionBudget",
			Namespace: "default", Name: "pdb",
			ReplaceAPIVersion: "policy/v1", ReplaceKind: "PodDisruptionBudget",
		},
		{
			// No replacement at all.
			Group: "policy", Version: "v1beta1", Kind: "PodSecurityPolicy",
			RemovedIn: "1.25", Replacement: "none (removed)",
			Name: "psp",
		},
	}

	table := buildTable(options, "1.25", findings)
	require.Len(t, table.Config.Rows, 3)

	linked, ok := table.Config.Rows[0]["Object"].(*component.Link)
	require.True(t, ok, "expected a link for a resolvable replacement")
	assert.Equal(t, "/overview/namespace/default/workloads/cron-jobs/cron", linked.Config.Ref)

	assert.IsType(t, &component.Text{}, table.Config.Rows[1]["Object"], "expected a text fallback when the path is unresolved")
	assert.IsType(t, &component.Text{}, table.Config.Rows[2]["Object"], "expected a text fallback when there is no replacement")
}
