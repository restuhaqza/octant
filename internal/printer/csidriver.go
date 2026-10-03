/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	storagev1 "k8s.io/api/storage/v1"

	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// CSIDriverListHandler is a printFunc that prints CSI drivers.
func CSIDriverListHandler(ctx context.Context, list *storagev1.CSIDriverList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("csi driver list is nil")
	}

	cols := component.NewTableCols("Name", "Attach Required", "Pod Info On Mount", "Age")
	ot := NewObjectTable("CSI Drivers", "We couldn't find any csi drivers!", cols, options.DashConfig.ObjectStore())
	ot.EnablePluginStatus(options.DashConfig.PluginManager())
	for _, driver := range list.Items {
		row := component.TableRow{}
		nameLink, err := options.Link.ForObject(&driver, driver.Name)
		if err != nil {
			return nil, err
		}

		row["Name"] = nameLink
		row["Attach Required"] = component.NewText(formatBoolPtr(driver.Spec.AttachRequired))
		row["Pod Info On Mount"] = component.NewText(formatBoolPtr(driver.Spec.PodInfoOnMount))
		row["Age"] = component.NewTimestamp(driver.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, &driver, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// CSIDriverHandler is a printFunc that prints a single CSI driver.
func CSIDriverHandler(ctx context.Context, driver *storagev1.CSIDriver, options Options) (component.Component, error) {
	o := NewObject(driver)
	o.EnableEvents()

	config, err := csiDriverConfig(driver)
	if err != nil {
		return nil, err
	}
	o.RegisterConfig(config)

	return o.ToComponent(ctx, options)
}

func csiDriverConfig(driver *storagev1.CSIDriver) (*component.Summary, error) {
	if driver == nil {
		return nil, errors.New("csi driver is nil")
	}

	modes := make([]string, len(driver.Spec.VolumeLifecycleModes))
	for i := range driver.Spec.VolumeLifecycleModes {
		modes[i] = string(driver.Spec.VolumeLifecycleModes[i])
	}

	sections := component.SummarySections{}
	sections.AddText("Attach Required", formatBoolPtr(driver.Spec.AttachRequired))
	sections.AddText("Pod Info On Mount", formatBoolPtr(driver.Spec.PodInfoOnMount))
	sections.AddText("Volume Lifecycle Modes", strings.Join(modes, ", "))
	sections.AddText("Storage Capacity", formatBoolPtr(driver.Spec.StorageCapacity))
	if driver.Spec.FSGroupPolicy != nil {
		sections.AddText("FS Group Policy", string(*driver.Spec.FSGroupPolicy))
	}
	sections.AddText("Requires Republish", formatBoolPtr(driver.Spec.RequiresRepublish))
	sections.AddText("SELinux Mount", formatBoolPtr(driver.Spec.SELinuxMount))

	return component.NewSummary("Configuration", sections...), nil
}
