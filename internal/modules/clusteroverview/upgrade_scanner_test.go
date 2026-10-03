/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package clusteroverview

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	configFake "github.com/vmware-tanzu/octant/internal/config/fake"
	"github.com/vmware-tanzu/octant/pkg/config"
)

type stubCRDWatcher struct{}

var _ config.CRDWatcher = (*stubCRDWatcher)(nil)

func (stubCRDWatcher) Watch(context.Context) error { return nil }

func (stubCRDWatcher) AddConfig(*config.CRDWatchConfig) error { return nil }

func Test_upgradeScanner_pathsRegistered(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	dash := configFake.NewMockDash(controller)
	dash.EXPECT().CRDWatcher().Return(stubCRDWatcher{}).AnyTimes()

	co, err := New(context.Background(), Options{DashConfig: dash})
	require.NoError(t, err)

	paths := []string{
		"/upgrade-scanner",
		"/upgrade-scanner/1.30",
		"/upgrade-scanner/1.30/networking.k8s.io",
	}
	for _, path := range paths {
		pf, err := co.pathMatcher.Find(path)
		require.NoError(t, err, "expected %s to resolve", path)
		assert.Equal(t, upgradeScannerDescriber, pf.Describer)
	}
}

func Test_upgradeScanner_isNotRootDescriberChild(t *testing.T) {
	for _, pf := range rootDescriber.PathFilters() {
		assert.False(t, pf.Match("/upgrade-scanner/1.30"),
			"the scanner must not be a child of rootDescriber or it runs on the Cluster Overview landing page")
	}
}
