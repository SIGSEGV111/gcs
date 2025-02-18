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

package ux

import (
	"reflect"

	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/slices"
)

const noAndOr = ""

var lastPrereqTypeUsed = model.TraitPrereqType

type prereqPanel struct {
	unison.Panel
	entity   *model.Entity
	root     **model.PrereqList
	andOrMap map[model.Prereq]*unison.Label
}

func newPrereqPanel(entity *model.Entity, root **model.PrereqList) *prereqPanel {
	p := &prereqPanel{
		entity:   entity,
		root:     root,
		andOrMap: make(map[model.Prereq]*unison.Label),
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{Columns: 1})
	p.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	p.SetBorder(unison.NewCompoundBorder(
		&TitledBorder{
			Title: i18n.Text("Prerequisites"),
			Font:  unison.LabelFont,
		},
		unison.NewEmptyBorder(unison.NewUniformInsets(2))))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.ContentColor.Paint(gc, rect, unison.Fill))
	}
	p.AddChild(p.createPrereqListPanel(0, *root))
	return p
}

func (p *prereqPanel) createPrereqListPanel(depth int, list *model.PrereqList) *unison.Panel {
	panel := unison.NewPanel()
	p.createButtonsPanel(panel, depth, list)
	inFront := andOrText(list) != noAndOr
	if inFront {
		p.addAndOr(panel, list)
	}
	addNumericCriteriaPanel(panel, nil, "", i18n.Text("When the Tech Level"), i18n.Text("When Tech Level"),
		&list.WhenTL, 0, fxp.Twelve, 1, true, true)
	popup := addBoolPopup(panel, i18n.Text("requires all of:"), i18n.Text("requires at least one of:"), &list.All)
	callback := popup.SelectionChangedCallback
	popup.SelectionChangedCallback = func(pop *unison.PopupMenu[string]) {
		callback(pop)
		p.adjustAndOrForList(list)
	}
	if !inFront {
		p.addAndOr(panel, list)
	}
	columns := len(panel.Children())
	panel.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HAlign:   unison.FillAlignment,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	for _, child := range list.Prereqs {
		p.addToList(panel, depth+1, -1, child)
	}
	return panel
}

func (p *prereqPanel) addToList(parent *unison.Panel, depth, index int, child model.Prereq) {
	var panel *unison.Panel
	switch one := child.(type) {
	case *model.PrereqList:
		panel = p.createPrereqListPanel(depth, one)
	case *model.TraitPrereq:
		panel = p.createTraitPrereqPanel(depth, one)
	case *model.AttributePrereq:
		panel = p.createAttributePrereqPanel(depth, one)
	case *model.ContainedQuantityPrereq:
		panel = p.createContainedQuantityPrereqPanel(depth, one)
	case *model.ContainedWeightPrereq:
		panel = p.createContainedWeightPrereqPanel(depth, one)
	case *model.EquippedEquipmentPrereq:
		panel = p.createEquippedEquipmentPrereqPanel(depth, one)
	case *model.SkillPrereq:
		panel = p.createSkillPrereqPanel(depth, one)
	case *model.SpellPrereq:
		panel = p.createSpellPrereqPanel(depth, one)
	default:
		jot.Warn(errs.Newf("unknown prerequisite type: %s", reflect.TypeOf(child).String()))
	}
	if panel != nil {
		columns := parent.Layout().(*unison.FlexLayout).Columns
		panel.SetLayoutData(&unison.FlexLayoutData{
			HSpan:  columns,
			HAlign: unison.FillAlignment,
			HGrab:  true,
		})
		if index < 0 {
			parent.AddChild(panel)
		} else {
			parent.AddChildAtIndex(panel, columns+index)
		}
	}
}

func (p *prereqPanel) createButtonsPanel(parent *unison.Panel, depth int, data model.Prereq) {
	buttons := unison.NewPanel()
	buttons.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: float32(depth * 20)}))
	parent.AddChild(buttons)
	if prereqList, ok := data.(*model.PrereqList); ok {
		addPrereqButton := unison.NewSVGButton(svg.CircledAdd)
		addPrereqButton.ClickCallback = func() {
			if created := p.createPrereqForType(lastPrereqTypeUsed, prereqList); created != nil {
				prereqList.Prereqs = slices.Insert(prereqList.Prereqs, 0, created)
				p.addToList(parent, depth+1, 0, created)
				p.adjustAndOrForList(prereqList)
				unison.Ancestor[*unison.DockContainer](p).MarkForLayoutRecursively()
				MarkModified(p)
			}
		}
		buttons.AddChild(addPrereqButton)

		addPrereqListButton := unison.NewSVGButton(svg.CircledVerticalEllipsis)
		addPrereqListButton.ClickCallback = func() {
			newList := model.NewPrereqList()
			newList.Parent = prereqList
			prereqList.Prereqs = slices.Insert(prereqList.Prereqs, 0, model.Prereq(newList))
			p.addToList(parent, depth+1, 0, newList)
			p.adjustAndOrForList(prereqList)
			unison.Ancestor[*unison.DockContainer](p).MarkForLayoutRecursively()
			MarkModified(p)
		}
		buttons.AddChild(addPrereqListButton)
	}
	parentList := data.ParentList()
	if parentList != nil {
		deleteButton := unison.NewSVGButton(svg.Trash)
		deleteButton.ClickCallback = func() {
			delete(p.andOrMap, data)
			if i := slices.IndexFunc(parentList.Prereqs, func(elem model.Prereq) bool { return elem == data }); i != -1 {
				parentList.Prereqs = slices.Delete(parentList.Prereqs, i, i+1)
			}
			parent.RemoveFromParent()
			p.adjustAndOrForList(parentList)
			unison.Ancestor[*unison.DockContainer](p).MarkForLayoutRecursively()
			MarkModified(p)
		}
		buttons.AddChild(deleteButton)
	}
	buttons.SetLayout(&unison.FlexLayout{
		Columns: len(buttons.Children()),
	})
}

func (p *prereqPanel) addAndOr(parent *unison.Panel, data model.Prereq) {
	label := NewFieldLeadingLabel(andOrText(data))
	parent.AddChild(label)
	p.andOrMap[data] = label
}

func (p *prereqPanel) adjustAndOrForList(list *model.PrereqList) {
	for _, one := range list.Prereqs {
		p.adjustAndOr(one)
	}
	p.MarkForLayoutRecursively()
}

func (p *prereqPanel) adjustAndOr(data model.Prereq) {
	if label, ok := p.andOrMap[data]; ok {
		if text := andOrText(data); text != label.Text {
			parent := label.Parent()
			label.RemoveFromParent()
			label.Text = text
			i := 1
			if text == noAndOr {
				i = parent.Layout().(*unison.FlexLayout).Columns - 1
			}
			parent.AddChildAtIndex(label, i)
		}
	}
}

func andOrText(pr model.Prereq) string {
	list := pr.ParentList()
	if list == nil || len(list.Prereqs) < 2 || list.Prereqs[0] == pr {
		return noAndOr
	}
	if list.All {
		return i18n.Text("and")
	}
	return i18n.Text("or")
}

func (p *prereqPanel) addPrereqTypeSwitcher(parent *unison.Panel, depth int, pr model.Prereq) {
	prereqType := pr.PrereqType()
	popup := addPopup(parent, model.AllPrereqType[1:], &prereqType)
	popup.SelectionChangedCallback = func(pop *unison.PopupMenu[model.PrereqType]) {
		if item, ok := pop.Selected(); ok {
			parentList := pr.ParentList()
			if newPrereq := p.createPrereqForType(item, parentList); newPrereq != nil {
				lastPrereqTypeUsed = item
				parentOfParent := parent.Parent()
				parent.RemoveFromParent()
				list := parentList.Prereqs
				i := slices.IndexFunc(list, func(one model.Prereq) bool { return one == pr })
				list[i] = newPrereq
				p.addToList(parentOfParent, depth, i, newPrereq)
				unison.Ancestor[*unison.DockContainer](p).MarkForLayoutRecursively()
				MarkModified(p)
			}
		}
	}
}

func (p *prereqPanel) createPrereqForType(prereqType model.PrereqType, parentList *model.PrereqList) model.Prereq {
	switch prereqType {
	case model.ListPrereqType:
		one := model.NewPrereqList()
		one.Parent = parentList
		return one
	case model.TraitPrereqType:
		one := model.NewTraitPrereq()
		one.Parent = parentList
		return one
	case model.AttributePrereqType:
		one := model.NewAttributePrereq(p.entity)
		one.Parent = parentList
		return one
	case model.ContainedQuantityPrereqType:
		one := model.NewContainedQuantityPrereq()
		one.Parent = parentList
		return one
	case model.ContainedWeightPrereqType:
		one := model.NewContainedWeightPrereq(p.entity)
		one.Parent = parentList
		return one
	case model.EquippedEquipmentPrereqType:
		one := model.NewEquippedEquipmentPrereq()
		one.Parent = parentList
		return one
	case model.SkillPrereqType:
		one := model.NewSkillPrereq()
		one.Parent = parentList
		return one
	case model.SpellPrereqType:
		one := model.NewSpellPrereq()
		one.Parent = parentList
		return one
	default:
		jot.Warn(errs.Newf("unknown prerequisite type: %s", prereqType.Key()))
		return nil
	}
}

func (p *prereqPanel) createTraitPrereqPanel(depth int, pr *model.TraitPrereq) *unison.Panel {
	panel := unison.NewPanel()
	p.createButtonsPanel(panel, depth, pr)
	inFront := andOrText(pr) != noAndOr
	if inFront {
		p.addAndOr(panel, pr)
	}
	addHasPopup(panel, &pr.Has)
	p.addPrereqTypeSwitcher(panel, depth, pr)
	if !inFront {
		p.addAndOr(panel, pr)
	}
	columns := len(panel.Children())
	panel.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	addNameCriteriaPanel(panel, &pr.NameCriteria, columns-1, true)
	addNotesCriteriaPanel(panel, &pr.NotesCriteria, columns-1, true)
	addLevelCriteriaPanel(panel, nil, "", &pr.LevelCriteria, columns-1, true)
	return panel
}

func (p *prereqPanel) createAttributePrereqPanel(depth int, pr *model.AttributePrereq) *unison.Panel {
	panel := unison.NewPanel()
	p.createButtonsPanel(panel, depth, pr)
	inFront := andOrText(pr) != noAndOr
	if inFront {
		p.addAndOr(panel, pr)
	}
	addHasPopup(panel, &pr.Has)
	p.addPrereqTypeSwitcher(panel, depth, pr)
	if !inFront {
		p.addAndOr(panel, pr)
	}
	columns := len(panel.Children())
	panel.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	second := unison.NewPanel()
	second.SetLayoutData(&unison.FlexLayoutData{HSpan: columns - 1})
	extra := model.SizeFlag | model.DodgeFlag | model.ParryFlag | model.BlockFlag
	addAttributeChoicePopup(second, p.entity, noAndOr, &pr.Which, extra)
	addAttributeChoicePopup(second, p.entity, i18n.Text("combined with"), &pr.CombinedWith, extra|model.BlankFlag)
	addNumericCriteriaPanel(second, nil, "", i18n.Text("which"), i18n.Text("Attribute Qualifier"),
		&pr.QualifierCriteria, fxp.Min, fxp.Max, 1, false, false)
	second.SetLayout(&unison.FlexLayout{
		Columns:  len(second.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.AddChild(unison.NewPanel())
	panel.AddChild(second)
	return panel
}

func (p *prereqPanel) createContainedQuantityPrereqPanel(depth int, pr *model.ContainedQuantityPrereq) *unison.Panel {
	panel := unison.NewPanel()
	p.createButtonsPanel(panel, depth, pr)
	inFront := andOrText(pr) != noAndOr
	if inFront {
		p.addAndOr(panel, pr)
	}
	addHasPopup(panel, &pr.Has)
	p.addPrereqTypeSwitcher(panel, depth, pr)
	addQuantityCriteriaPanel(panel, nil, "", &pr.QualifierCriteria)
	if !inFront {
		p.addAndOr(panel, pr)
	}
	columns := len(panel.Children())
	panel.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	return panel
}

func (p *prereqPanel) createContainedWeightPrereqPanel(depth int, pr *model.ContainedWeightPrereq) *unison.Panel {
	panel := unison.NewPanel()
	p.createButtonsPanel(panel, depth, pr)
	inFront := andOrText(pr) != noAndOr
	if inFront {
		p.addAndOr(panel, pr)
	}
	addHasPopup(panel, &pr.Has)
	p.addPrereqTypeSwitcher(panel, depth, pr)
	if !inFront {
		p.addAndOr(panel, pr)
	}
	columns := len(panel.Children())
	panel.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	second := unison.NewPanel()
	second.SetLayoutData(&unison.FlexLayoutData{HSpan: columns - 1})
	addWeightCriteriaPanel(second, nil, "", p.entity, &pr.WeightCriteria)
	second.SetLayout(&unison.FlexLayout{
		Columns:  len(second.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.AddChild(unison.NewPanel())
	panel.AddChild(second)
	return panel
}

func (p *prereqPanel) createEquippedEquipmentPrereqPanel(depth int, pr *model.EquippedEquipmentPrereq) *unison.Panel {
	panel := unison.NewPanel()
	p.createButtonsPanel(panel, depth, pr)
	inFront := andOrText(pr) != noAndOr
	if inFront {
		p.addAndOr(panel, pr)
	}
	// addHasPopup(panel, &pr.Has)
	p.addPrereqTypeSwitcher(panel, depth, pr)
	if !inFront {
		p.addAndOr(panel, pr)
	}
	columns := len(panel.Children())
	panel.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	addNameCriteriaPanel(panel, &pr.NameCriteria, columns-1, true)
	return panel
}

func (p *prereqPanel) createSkillPrereqPanel(depth int, pr *model.SkillPrereq) *unison.Panel {
	panel := unison.NewPanel()
	p.createButtonsPanel(panel, depth, pr)
	inFront := andOrText(pr) != noAndOr
	if inFront {
		p.addAndOr(panel, pr)
	}
	addHasPopup(panel, &pr.Has)
	p.addPrereqTypeSwitcher(panel, depth, pr)
	if !inFront {
		p.addAndOr(panel, pr)
	}
	columns := len(panel.Children())
	panel.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	addNameCriteriaPanel(panel, &pr.NameCriteria, columns-1, true)
	addSpecializationCriteriaPanel(panel, &pr.SpecializationCriteria, columns-1, true)
	addLevelCriteriaPanel(panel, nil, "", &pr.LevelCriteria, columns-1, true)
	return panel
}

func (p *prereqPanel) createSpellPrereqPanel(depth int, pr *model.SpellPrereq) *unison.Panel {
	panel := unison.NewPanel()
	p.createButtonsPanel(panel, depth, pr)
	inFront := andOrText(pr) != noAndOr
	if inFront {
		p.addAndOr(panel, pr)
	}
	addHasPopup(panel, &pr.Has)
	addQuantityCriteriaPanel(panel, nil, "", &pr.QuantityCriteria)
	p.addPrereqTypeSwitcher(panel, depth, pr)
	if !inFront {
		p.addAndOr(panel, pr)
	}
	columns := len(panel.Children())
	panel.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	second := unison.NewPanel()
	second.SetLayoutData(&unison.FlexLayoutData{HSpan: columns - 1})
	subTypePopup := addPopup[model.SpellComparisonType](second, model.AllSpellComparisonType, &pr.SubType)
	popup, field := addStringCriteriaPanel(second, "", "", i18n.Text("Spell Qualifier"), &pr.QualifierCriteria, 1, false)
	savedCallback := subTypePopup.SelectionChangedCallback
	subTypePopup.SelectionChangedCallback = func(pop *unison.PopupMenu[model.SpellComparisonType]) {
		savedCallback(pop)
		blank := pr.SubType == model.AnySpellComparisonType || pr.SubType == model.CollegeCountSpellComparisonType
		adjustPopupBlank(popup, blank)
		adjustFieldBlank(field, blank)
	}
	adjustPopupBlank(popup, pr.SubType == model.AnySpellComparisonType || pr.SubType == model.CollegeCountSpellComparisonType)
	adjustFieldBlank(field, pr.SubType == model.AnySpellComparisonType || pr.SubType == model.CollegeCountSpellComparisonType)
	second.SetLayout(&unison.FlexLayout{
		Columns:  len(second.Children()),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.AddChild(unison.NewPanel())
	panel.AddChild(second)
	return panel
}
