/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	storagev1 "k8s.io/api/storage/v1"

	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// VolumeAttachmentListHandler is a printFunc that prints volume attachments.
func VolumeAttachmentListHandler(ctx context.Context, list *storagev1.VolumeAttachmentList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("volume attachment list is nil")
	}

	cols := component.NewTableCols("Name", "Attacher", "Node", "Attached", "Age")
	ot := NewObjectTable("Volume Attachments", "We couldn't find any volume attachments!", cols, options.DashConfig.ObjectStore())
	ot.EnablePluginStatus(options.DashConfig.PluginManager())
	for _, attachment := range list.Items {
		row := component.TableRow{}
		nameLink, err := options.Link.ForObject(&attachment, attachment.Name)
		if err != nil {
			return nil, err
		}

		row["Name"] = nameLink
		row["Attacher"] = component.NewText(attachment.Spec.Attacher)
		row["Node"] = component.NewText(attachment.Spec.NodeName)
		row["Attached"] = component.NewText(fmt.Sprintf("%v", attachment.Status.Attached))
		row["Age"] = component.NewTimestamp(attachment.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, &attachment, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// VolumeAttachmentHandler is a printFunc that prints a single volume attachment.
func VolumeAttachmentHandler(ctx context.Context, attachment *storagev1.VolumeAttachment, options Options) (component.Component, error) {
	o := NewObject(attachment)
	o.EnableEvents()

	config, err := volumeAttachmentConfig(attachment)
	if err != nil {
		return nil, err
	}
	o.RegisterConfig(config)

	return o.ToComponent(ctx, options)
}

func volumeAttachmentConfig(attachment *storagev1.VolumeAttachment) (*component.Summary, error) {
	if attachment == nil {
		return nil, errors.New("volume attachment is nil")
	}

	sections := component.SummarySections{}
	sections.AddText("Attacher", attachment.Spec.Attacher)
	sections.AddText("Node", attachment.Spec.NodeName)
	if attachment.Spec.Source.PersistentVolumeName != nil {
		sections.AddText("Persistent Volume", *attachment.Spec.Source.PersistentVolumeName)
	}
	sections.AddText("Attached", fmt.Sprintf("%v", attachment.Status.Attached))
	if attachment.Status.AttachError != nil {
		sections.AddText("Attach Error", attachment.Status.AttachError.Message)
	}
	if attachment.Status.DetachError != nil {
		sections.AddText("Detach Error", attachment.Status.DetachError.Message)
	}

	return component.NewSummary("Configuration", sections...), nil
}
