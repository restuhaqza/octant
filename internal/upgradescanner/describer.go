/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package upgradescanner

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/version"

	"github.com/vmware-tanzu/octant/internal/describer"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

const (
	// defaultTargetVersion is used when neither a target field nor the cluster
	// server version can be determined.
	defaultTargetVersion = "1.32"

	// scannerContentPath is the content path (including the module prefix) of
	// this view and is used to build internal links.
	scannerContentPath = "/cluster-overview/upgrade-scanner"
)

// Describer describes the Upgrade & Deprecation Scanner view.
type Describer struct{}

var _ describer.Describer = (*Describer)(nil)

// NewDescriber creates a new Upgrade & Deprecation Scanner describer.
func NewDescriber() *Describer {
	return &Describer{}
}

// Describe builds the scanner view for the requested target version.
func (d *Describer) Describe(ctx context.Context, namespace string, options describer.Options) (component.ContentResponse, error) {
	clusterVersion := ""
	if options.Dash != nil {
		if client := options.ClusterClient(); client != nil {
			if dc, err := client.DiscoveryClient(); err == nil {
				if info, err := dc.ServerVersion(); err == nil {
					if minor, err := clusterMinorVersion(info); err == nil {
						clusterVersion = minor
					}
				}
			}
		}
	}

	target := resolveTarget(options.Fields, clusterVersion)
	group := ""
	if options.Fields != nil {
		group = strings.TrimSpace(options.Fields["group"])
	}

	var findings []Finding
	if options.Dash != nil {
		if client := options.ClusterClient(); client != nil {
			if dc, err := client.DiscoveryClient(); err == nil && dc != nil {
				if objectStore := options.ObjectStore(); objectStore != nil {
					// A scan failure should not make the whole view fail.
					findings, _ = Scan(ctx, dc, objectStore, namespace, target)
				}
			}
		}
	}

	presentGroups := distinctGroups(findings)
	displayFindings := findings
	if group != "" {
		displayFindings = filterByGroup(findings, group)
	}

	list := component.NewList(component.Title(component.NewText("Upgrade & Deprecation Scanner")), nil)
	list.Add(buildSummary(clusterVersion, target, displayFindings))
	list.Add(buildTargetPicker(clusterVersion, target))
	list.Add(buildGroupPicker(target, group, presentGroups))
	list.Add(buildTable(options, target, displayFindings))

	return component.ContentResponse{
		Components: []component.Component{list},
	}, nil
}

// PathFilters returns the path filters for this describer.
func (d *Describer) PathFilters() []describer.PathFilter {
	return []describer.PathFilter{
		*describer.NewPathFilter("/upgrade-scanner", d),
		*describer.NewPathFilter(`/upgrade-scanner/(?P<target>[0-9]+\.[0-9]+)`, d),
		*describer.NewPathFilter(`/upgrade-scanner/(?P<target>[0-9]+\.[0-9]+)/(?P<group>[A-Za-z0-9._-]+)`, d),
	}
}

// Reset resets the describer.
func (d *Describer) Reset(ctx context.Context) error {
	return nil
}

// resolveTarget picks the target Kubernetes version. An explicit target field
// wins; otherwise the cluster minor version + 1 is used; otherwise the default.
func resolveTarget(fields map[string]string, clusterVersion string) string {
	if fields != nil {
		if target := strings.TrimSpace(fields["target"]); target != "" {
			return target
		}
	}

	if clusterVersion != "" {
		if next, err := nextMinorVersion(clusterVersion); err == nil {
			return next
		}
	}

	return defaultTargetVersion
}

// clusterMinorVersion returns the cluster version formatted as "major.minor".
func clusterMinorVersion(info *version.Info) (string, error) {
	if info == nil {
		return "", fmt.Errorf("version info is nil")
	}

	if gitVersion := info.GitVersion; gitVersion != "" {
		if major, minor, err := parseKubeVersion(gitVersion); err == nil {
			return fmt.Sprintf("%d.%d", major, minor), nil
		}
	}

	major := strings.TrimSuffix(strings.TrimSpace(info.Major), "+")
	minor := strings.TrimSuffix(strings.TrimSpace(info.Minor), "+")
	if major == "" || minor == "" {
		return "", fmt.Errorf("incomplete server version")
	}

	return fmt.Sprintf("%s.%s", major, minor), nil
}

// distinctGroups returns the sorted, unique API groups present in findings.
func distinctGroups(findings []Finding) []string {
	seen := make(map[string]bool)
	var groups []string
	for _, finding := range findings {
		if finding.Group == "" || seen[finding.Group] {
			continue
		}
		seen[finding.Group] = true
		groups = append(groups, finding.Group)
	}
	sort.Strings(groups)
	return groups
}

func filterByGroup(findings []Finding, group string) []Finding {
	var filtered []Finding
	for _, finding := range findings {
		if finding.Group == group {
			filtered = append(filtered, finding)
		}
	}
	return filtered
}

// targetVersions returns the range of target versions offered by the picker.
func targetVersions(clusterVersion string) []string {
	start := 0
	if _, minor, err := parseKubeVersion(clusterVersion); err == nil {
		start = minor
	} else if _, minor, err := parseKubeVersion(defaultTargetVersion); err == nil {
		start = minor
	}

	end := start + 4
	if latest := latestRemovalVersion(); latest != "" {
		if _, minor, err := parseKubeVersion(latest); err == nil && minor+2 > end {
			end = minor + 2
		}
	}

	var versions []string
	for minor := start; minor <= end; minor++ {
		versions = append(versions, fmt.Sprintf("1.%d", minor))
	}

	return versions
}

func buildSummary(clusterVersion, target string, findings []Finding) *component.Summary {
	if clusterVersion == "" {
		clusterVersion = "unknown"
	}

	removed := 0
	groupCounts := make(map[string]int)
	for _, finding := range findings {
		if finding.Removed {
			removed++
		}
		groupCounts[finding.GroupVersion()]++
	}

	summary := component.NewSummary("Upgrade & Deprecation Scanner")
	summary.AddSection("Cluster Version", component.NewText(clusterVersion))
	summary.AddSection("Target Version", component.NewText(target))
	summary.AddSection("Findings", component.NewTextf(
		"%d live object(s) use deprecated or removed APIs; %d are removed by Kubernetes %s.",
		len(findings), removed, target))
	summary.AddSection("By API Group", component.NewText(formatGroupCounts(groupCounts)))

	return summary
}

func formatGroupCounts(groupCounts map[string]int) string {
	if len(groupCounts) == 0 {
		return "none"
	}

	keys := make([]string, 0, len(groupCounts))
	for key := range groupCounts {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s: %d", key, groupCounts[key]))
	}

	return strings.Join(parts, ", ")
}

func buildTargetPicker(clusterVersion, target string) *component.FlexLayout {
	picker := component.NewFlexLayout("Target Kubernetes Version")

	var items component.FlexLayoutSection
	for _, version := range targetVersions(clusterVersion) {
		text := version
		if version == target {
			text = fmt.Sprintf("%s (selected)", version)
		}
		ref := fmt.Sprintf("%s/%s", scannerContentPath, version)
		items = append(items, component.FlexLayoutItem{
			Width: component.WidthQuarter,
			View:  component.NewLink("", text, ref),
		})
	}

	picker.AddSections(items)

	return picker
}

func buildGroupPicker(target, selectedGroup string, groups []string) *component.FlexLayout {
	picker := component.NewFlexLayout("API Group")

	allText := "All groups"
	if selectedGroup == "" {
		allText = "All groups (selected)"
	}

	items := component.FlexLayoutSection{
		{
			Width: component.WidthQuarter,
			View:  component.NewLink("", allText, fmt.Sprintf("%s/%s", scannerContentPath, target)),
		},
	}

	for _, group := range groups {
		text := group
		if group == selectedGroup {
			text = fmt.Sprintf("%s (selected)", group)
		}
		items = append(items, component.FlexLayoutItem{
			Width: component.WidthQuarter,
			View:  component.NewLink("", text, fmt.Sprintf("%s/%s/%s", scannerContentPath, target, group)),
		})
	}

	picker.AddSections(items)

	return picker
}

func buildTable(options describer.Options, target string, findings []Finding) *component.Table {
	columns := component.NewTableCols("Object", "Namespace", "API Version", "Kind", "Removed In", "Replacement")
	placeholder := fmt.Sprintf("There are no live objects using APIs removed in Kubernetes %s.", target)
	table := component.NewTable("Deprecated & Removed API Usage", placeholder, columns)

	for _, finding := range findings {
		var object component.Component = component.NewText(finding.Name)
		// Link to the replacement GVK when the object path can actually be
		// resolved (a link with an empty ref is a dead link).
		if options.Link != nil && finding.ReplaceAPIVersion != "" && finding.ReplaceKind != "" {
			if l, err := options.Link.ForGVK(finding.Namespace, finding.ReplaceAPIVersion, finding.ReplaceKind, finding.Name, finding.Name); err == nil && l != nil && l.Config.Ref != "" {
				object = l
			}
		}

		namespace := finding.Namespace
		if namespace == "" {
			namespace = "<cluster>"
		}

		removedIn := fmt.Sprintf("%s (deprecated)", finding.RemovedIn)
		if finding.Removed {
			removedIn = fmt.Sprintf("%s (removed)", finding.RemovedIn)
		}

		table.Add(component.TableRow{
			"Object":      object,
			"Namespace":   component.NewText(namespace),
			"API Version": component.NewText(finding.GroupVersion()),
			"Kind":        component.NewText(finding.Kind),
			"Removed In":  component.NewText(removedIn),
			"Replacement": component.NewText(finding.Replacement),
		})
	}

	return table
}
