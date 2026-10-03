/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package upgradescanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseKubeVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		major   int
		minor   int
		isErr   bool
	}{
		{name: "minor", version: "1.25", major: 1, minor: 25},
		{name: "v prefix", version: "v1.29", major: 1, minor: 29},
		{name: "patch", version: "v1.25.3", major: 1, minor: 25},
		{name: "build suffix", version: "1.28.0+", major: 1, minor: 28},
		{name: "minor suffix", version: "1.25+", major: 1, minor: 25},
		{name: "empty", version: "", isErr: true},
		{name: "invalid", version: "not-a-version", isErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			major, minor, err := parseKubeVersion(test.version)
			if test.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.major, major)
			assert.Equal(t, test.minor, minor)
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{name: "older", a: "1.25", b: "1.29", expected: -1},
		{name: "newer", a: "1.32", b: "1.26", expected: 1},
		{name: "equal", a: "1.22", b: "v1.22.0", expected: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, compareVersions(test.a, test.b))
		})
	}
}

func TestIsRemovedInOrBefore(t *testing.T) {
	assert.True(t, isRemovedInOrBefore("1.25", "1.25"))
	assert.True(t, isRemovedInOrBefore("1.22", "1.29"))
	assert.False(t, isRemovedInOrBefore("1.32", "1.29"))
}

func TestNextMinorVersion(t *testing.T) {
	got, err := nextMinorVersion("1.27")
	require.NoError(t, err)
	assert.Equal(t, "1.28", got)

	_, err = nextMinorVersion("nope")
	require.Error(t, err)
}

func TestLatestRemovalVersion(t *testing.T) {
	assert.Equal(t, "1.32", latestRemovalVersion())
}

func TestRemovalsWellFormed(t *testing.T) {
	require.NotEmpty(t, Removals)

	seen := make(map[removalKey]bool)
	for _, removal := range Removals {
		assert.NotEmpty(t, removal.Version, "version must be set")
		assert.NotEmpty(t, removal.Kind, "kind must be set")
		assert.NotEmpty(t, removal.RemovedIn, "removedIn must be set")
		assert.NotEmpty(t, removal.Replacement, "replacement must be set")

		_, _, err := parseKubeVersion(removal.RemovedIn)
		require.NoError(t, err, "removedIn %q should parse", removal.RemovedIn)

		if removal.Replacement == "none (removed)" {
			assert.Empty(t, removal.ReplaceAPIVersion, "replacement API version should be empty when there is no replacement")
			assert.Empty(t, removal.ReplaceKind, "replacement kind should be empty when there is no replacement")
		} else {
			assert.NotEmpty(t, removal.ReplaceAPIVersion, "replacement API version should be set")
			assert.NotEmpty(t, removal.ReplaceKind, "replacement kind should be set")
		}

		key := removalKey{group: removal.Group, version: removal.Version, kind: removal.Kind}
		require.False(t, seen[key], "duplicate removal for %s/%s %s", removal.Group, removal.Version, removal.Kind)
		seen[key] = true
	}
}

func TestRemovalsPinnedVersions(t *testing.T) {
	tests := []struct {
		group     string
		version   string
		kind      string
		removedIn string
	}{
		{group: "storage.k8s.io", version: "v1beta1", kind: "CSIDriver", removedIn: "1.22"},
		{group: "storage.k8s.io", version: "v1beta1", kind: "CSINode", removedIn: "1.22"},
		{group: "storage.k8s.io", version: "v1beta1", kind: "StorageClass", removedIn: "1.22"},
		{group: "storage.k8s.io", version: "v1beta1", kind: "VolumeAttachment", removedIn: "1.22"},
		{group: "storage.k8s.io", version: "v1beta1", kind: "CSIStorageCapacity", removedIn: "1.27"},
		{group: "extensions", version: "v1beta1", kind: "Ingress", removedIn: "1.22"},
		{group: "extensions", version: "v1beta1", kind: "Deployment", removedIn: "1.16"},
		{group: "extensions", version: "v1beta1", kind: "DaemonSet", removedIn: "1.16"},
		{group: "extensions", version: "v1beta1", kind: "ReplicaSet", removedIn: "1.16"},
		{group: "extensions", version: "v1beta1", kind: "NetworkPolicy", removedIn: "1.16"},
		{group: "extensions", version: "v1beta1", kind: "PodSecurityPolicy", removedIn: "1.16"},
	}

	for _, test := range tests {
		t.Run(test.group+"/"+test.version+" "+test.kind, func(t *testing.T) {
			removal, ok := removalLookup[removalKey{group: test.group, version: test.version, kind: test.kind}]
			require.True(t, ok, "expected %s/%s %s in the database", test.group, test.version, test.kind)
			assert.Equal(t, test.removedIn, removal.RemovedIn)
		})
	}
}

func TestRemovalsExcludeFabricatedPairs(t *testing.T) {
	_, ok := removalLookup[removalKey{group: "extensions", version: "v1beta1", kind: "StatefulSet"}]
	assert.False(t, ok, "extensions/v1beta1 StatefulSet never existed")

	_, ok = removalLookup[removalKey{group: "apps", version: "v1beta1", kind: "DaemonSet"}]
	assert.False(t, ok, "apps/v1beta1 DaemonSet never existed")

	_, ok = removalLookup[removalKey{group: "storage.k8s.io", version: "v1beta1", kind: "StorageClass"}]
	require.True(t, ok)
}

func TestRemovalLookupCoversKnownRemovals(t *testing.T) {
	removal, ok := removalLookup[removalKey{group: "batch", version: "v1beta1", kind: "CronJob"}]
	require.True(t, ok)
	assert.Equal(t, "1.25", removal.RemovedIn)
	assert.Equal(t, "batch/v1 CronJob", removal.Replacement)

	_, ok = removalLookup[removalKey{group: "apps", version: "v1", kind: "Deployment"}]
	assert.False(t, ok)
}
