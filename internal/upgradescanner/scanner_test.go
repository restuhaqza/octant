/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package upgradescanner

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	clusterFake "github.com/vmware-tanzu/octant/internal/cluster/fake"
	"github.com/vmware-tanzu/octant/pkg/store"
	storeFake "github.com/vmware-tanzu/octant/pkg/store/fake"
)

func cronJobResourceList() *metav1.APIResourceList {
	return &metav1.APIResourceList{
		GroupVersion: "batch/v1beta1",
		APIResources: []metav1.APIResource{
			{Name: "cronjobs", Kind: "CronJob", Namespaced: true, Verbs: []string{"get", "list"}},
			{Name: "cronjobs/status", Kind: "CronJob", Namespaced: true, Verbs: []string{"get"}},
		},
	}
}

func cronJobList(namespaces ...string) *unstructured.UnstructuredList {
	list := &unstructured.UnstructuredList{}
	for _, namespace := range namespaces {
		list.Items = append(list.Items, unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "batch/v1beta1",
				"kind":       "CronJob",
				"metadata": map[string]interface{}{
					"name":      "cron-" + namespace,
					"namespace": namespace,
				},
			},
		})
	}
	return list
}

func TestScanReportsRemovedAPIs(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	discoveryClient := clusterFake.NewMockDiscoveryInterface(controller)
	discoveryClient.EXPECT().
		ServerGroupsAndResources().
		Return(nil, []*metav1.APIResourceList{cronJobResourceList()}, nil)

	objectStore := storeFake.NewMockStore(controller)
	objectStore.EXPECT().
		List(gomock.Any(), store.Key{APIVersion: "batch/v1beta1", Kind: "CronJob"}).
		Return(cronJobList("default", "kube-system"), false, nil)

	findings, err := Scan(context.Background(), discoveryClient, objectStore, "", "1.26")
	require.NoError(t, err)
	require.Len(t, findings, 2)

	finding := findings[0]
	assert.Equal(t, "batch", finding.Group)
	assert.Equal(t, "v1beta1", finding.Version)
	assert.Equal(t, "CronJob", finding.Kind)
	assert.Equal(t, "1.25", finding.RemovedIn)
	assert.Equal(t, "batch/v1 CronJob", finding.Replacement)
	assert.Equal(t, "batch/v1beta1", finding.GroupVersion())
	assert.True(t, finding.Removed)
	assert.Equal(t, "batch/v1", finding.ReplaceAPIVersion)
	assert.Equal(t, "CronJob", finding.ReplaceKind)
	assert.Equal(t, "default", finding.Namespace)
}

func TestScanMarksNotYetRemovedAsDeprecated(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	discoveryClient := clusterFake.NewMockDiscoveryInterface(controller)
	discoveryClient.EXPECT().
		ServerGroupsAndResources().
		Return(nil, []*metav1.APIResourceList{cronJobResourceList()}, nil)

	objectStore := storeFake.NewMockStore(controller)
	objectStore.EXPECT().
		List(gomock.Any(), gomock.Any()).
		Return(cronJobList("default"), false, nil)

	findings, err := Scan(context.Background(), discoveryClient, objectStore, "", "1.24")
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.False(t, findings[0].Removed)
}

func TestScanIgnoresStableAPIs(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	discoveryClient := clusterFake.NewMockDiscoveryInterface(controller)
	discoveryClient.EXPECT().
		ServerGroupsAndResources().
		Return(nil, []*metav1.APIResourceList{
			{
				GroupVersion: "apps/v1",
				APIResources: []metav1.APIResource{
					{Name: "deployments", Kind: "Deployment", Namespaced: true},
				},
			},
		}, nil)

	// No List expectation: the mock fails if a stable API is listed.
	objectStore := storeFake.NewMockStore(controller)

	findings, err := Scan(context.Background(), discoveryClient, objectStore, "", "1.36")
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestScanToleratesPartialDiscoveryFailure(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	discoveryClient := clusterFake.NewMockDiscoveryInterface(controller)
	discoveryClient.EXPECT().
		ServerGroupsAndResources().
		Return(nil, []*metav1.APIResourceList{cronJobResourceList()}, errors.New("some groups failed"))

	objectStore := storeFake.NewMockStore(controller)
	objectStore.EXPECT().
		List(gomock.Any(), gomock.Any()).
		Return(cronJobList("default"), false, nil)

	findings, err := Scan(context.Background(), discoveryClient, objectStore, "", "1.25")
	require.NoError(t, err)
	require.Len(t, findings, 1)
}

func TestScanScopesNamespacedResources(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	discoveryClient := clusterFake.NewMockDiscoveryInterface(controller)
	discoveryClient.EXPECT().
		ServerGroupsAndResources().
		Return(nil, []*metav1.APIResourceList{cronJobResourceList()}, nil)

	objectStore := storeFake.NewMockStore(controller)
	objectStore.EXPECT().
		List(gomock.Any(), store.Key{Namespace: "default", APIVersion: "batch/v1beta1", Kind: "CronJob"}).
		Return(cronJobList("default"), false, nil)

	findings, err := Scan(context.Background(), discoveryClient, objectStore, "default", "1.25")
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "default", findings[0].Namespace)
}

func TestScanIgnoresListErrors(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	discoveryClient := clusterFake.NewMockDiscoveryInterface(controller)
	discoveryClient.EXPECT().
		ServerGroupsAndResources().
		Return(nil, []*metav1.APIResourceList{cronJobResourceList()}, nil)

	objectStore := storeFake.NewMockStore(controller)
	objectStore.EXPECT().
		List(gomock.Any(), gomock.Any()).
		Return(nil, false, errors.New("no informer"))

	findings, err := Scan(context.Background(), discoveryClient, objectStore, "", "1.25")
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestScanRequiresClients(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	_, err := Scan(context.Background(), nil, storeFake.NewMockStore(controller), "", "1.25")
	require.Error(t, err)

	discoveryClient := clusterFake.NewMockDiscoveryInterface(controller)
	_, err = Scan(context.Background(), discoveryClient, nil, "", "1.25")
	require.Error(t, err)
}
