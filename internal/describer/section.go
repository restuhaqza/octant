/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package describer

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// Section is a wrapper to combine content from multiple describers.
type Section struct {
	path       string
	title      string
	describers []Describer
}

var _ Describer = (*Section)(nil)

// NewSection creates a Section.
func NewSection(p, title string, describers ...Describer) *Section {
	return &Section{
		path:       p,
		title:      title,
		describers: describers,
	}
}

// nestedSection is implemented by Section. It lets a parent section render a
// nested section without that section adding its own empty-state placeholder,
// which would otherwise show up out of context on combined pages like the
// namespace overview.
type nestedSection interface {
	component(ctx context.Context, namespace string, options Options, nested bool) (*component.List, error)
}

// Describe generates content.
func (d *Section) Describe(ctx context.Context, namespace string, options Options) (component.ContentResponse, error) {
	list, err := d.component(ctx, namespace, options, false)
	if err != nil {
		return component.EmptyContentResponse, err
	}

	cr := component.ContentResponse{
		Components: []component.Component{list},
		Title:      component.Title(component.NewText(d.title)),
	}

	return cr, nil
}

// Component returns the list for a section. It is used when a section is
// embedded within another view (for example the namespace overview inside an
// application page), so it is treated as nested and does not add the
// page-level empty-state placeholder.
func (d *Section) Component(ctx context.Context, namespace string, options Options) (*component.List, error) {
	return d.component(ctx, namespace, options, true)
}

func (d *Section) component(ctx context.Context, namespace string, options Options, nested bool) (*component.List, error) {
	title := component.Title(component.NewText(d.title))
	list := component.NewList(title, nil)

	for describerIndex := range d.describers {
		child := d.describers[describerIndex]

		var nestedLists []*component.List
		if section, ok := child.(nestedSection); ok {
			nestedList, err := section.component(ctx, namespace, options, true)
			if err != nil {
				return nil, err
			}
			nestedLists = append(nestedLists, nestedList)
		} else {
			cResponse, err := child.Describe(ctx, namespace, options)
			if err != nil {
				return nil, err
			}

			for componentIndex := range cResponse.Components {
				if nestedList, ok := cResponse.Components[componentIndex].(*component.List); ok {
					nestedLists = append(nestedLists, nestedList)
				}
			}
		}

		for _, nestedList := range nestedLists {
			for itemIndex := range nestedList.Config.Items {
				item := nestedList.Config.Items[itemIndex]
				if !item.IsEmpty() {
					list.Add(item)
				}
			}
		}
	}

	// A page-level section with no content would otherwise render blank. Nested
	// sections are left empty on purpose, so combined pages (for example the
	// namespace overview) don't show a placeholder for each empty subsection.
	if !nested && len(list.Config.Items) == 0 {
		list.Add(component.NewText(emptySectionMessage(d.title, namespace)))
	}

	return list, nil
}

// emptySectionMessage is the placeholder shown when a page-level section has
// nothing to display, so the page is never left blank.
func emptySectionMessage(title, namespace string) string {
	if namespace != "" {
		return fmt.Sprintf("No %s resources were found in namespace %q.", title, namespace)
	}

	return fmt.Sprintf("No %s resources were found.", title)
}

// PathFilters returns path filters for the section.
func (d *Section) PathFilters() []PathFilter {
	PathFilters := []PathFilter{
		*NewPathFilter(d.path, d),
	}

	for _, child := range d.describers {
		PathFilters = append(PathFilters, child.PathFilters()...)
	}

	return PathFilters
}

func (d *Section) Reset(ctx context.Context) error {
	for i := range d.describers {
		if err := d.describers[i].Reset(ctx); err != nil {
			return errors.Wrapf(err, "reset describer in section %s", d.path)
		}
	}

	return nil
}
