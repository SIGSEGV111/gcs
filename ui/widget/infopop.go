/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package widget

import (
	"fmt"
	"strings"

	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// NewDefaultInfoPop creates a new InfoPop with the message about mouse wheel scaling.
func NewDefaultInfoPop() *unison.Button {
	button := NewInfoPop()
	AddScalingHelpToInfoPop(button)
	return button
}

// NewInfoPop creates a new InfoPop.
func NewInfoPop() *unison.Button {
	return unison.NewSVGButton(res.InfoSVG)
}

// ClearInfoPop clears the InfoPop data.
func ClearInfoPop(target unison.Paneler) {
	panel := target.AsPanel()
	panel.Tooltip = nil
	panel.TooltipImmediate = false
}

// AddHelpToInfoPop adds one or more lines of help text to an InfoPop.
func AddHelpToInfoPop(target unison.Paneler, text string) {
	tip := prepareInfoPop(target)
	for _, str := range strings.Split(text, "\n") {
		if str == "" && len(tip.Children()) == 0 {
			continue
		}
		label := unison.NewLabel()
		label.LabelTheme = unison.DefaultTooltipTheme.Label
		label.Text = str
		label.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
		tip.AddChild(label)
	}
}

// AddScalingHelpToInfoPop adds the help info about scaling to an InfoPop.
func AddScalingHelpToInfoPop(target unison.Paneler) {
	AddHelpToInfoPop(target, fmt.Sprintf(i18n.Text(`
Holding down the %s key while using
the mouse wheel will change the scale.`), unison.OptionModifier.String()))
}

// AddKeyBindingInfoToInfoPop adds information about a key binding to an InfoPop.
func AddKeyBindingInfoToInfoPop(target unison.Paneler, keyBinding unison.KeyBinding, text string) {
	keyLabel := unison.NewLabel()
	keyLabel.LabelTheme = unison.DefaultTooltipTheme.Label
	keyLabel.OnBackgroundInk = unison.DefaultTooltipTheme.BackgroundInk
	keyLabel.Font = unison.DefaultMenuItemTheme.KeyFont
	keyLabel.Text = keyBinding.String()
	keyLabel.HAlign = unison.MiddleAlignment
	keyLabel.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.FillAlignment})
	keyLabel.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.DefaultTooltipTheme.Label.OnBackgroundInk.Paint(gc, rect, unison.Fill))
		keyLabel.DefaultDraw(gc, rect)
	}
	keyLabel.SetBorder(unison.NewEmptyBorder(unison.NewHorizontalInsets(4)))
	tip := prepareInfoPop(target)
	tip.AddChild(keyLabel)

	descLabel := unison.NewLabel()
	descLabel.LabelTheme = unison.DefaultTooltipTheme.Label
	descLabel.Text = text
	tip.AddChild(descLabel)
}

func prepareInfoPop(target unison.Paneler) *unison.Panel {
	panel := target.AsPanel()
	panel.TooltipImmediate = true
	if panel.Tooltip == nil {
		panel.Tooltip = unison.NewTooltipBase()
		panel.Tooltip.SetLayout(&unison.FlexLayout{
			Columns:  2,
			HSpacing: unison.StdHSpacing,
			VSpacing: unison.StdVSpacing,
		})
	}
	return panel.Tooltip
}
