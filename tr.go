// Copyright 2013 ChaiShushan <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gettext

import (
	"encoding/json"
	"github.com/ContextLogic/goi18n/mo"
	"github.com/ContextLogic/goi18n/plural"
	"github.com/ContextLogic/goi18n/po"
)

var nilTranslator = &translator{
	MessageMap:    make(map[string]mo.Message),
	PluralFormula: plural.FormulaByLang("??"),
}

type translator struct {
	MessageMap    map[string]mo.Message
	PluralFormula func(n uint32) int
}

func newMoTranslator(name string, data []byte) (*translator, error) {
	var (
		f   *mo.File
		err error
	)
	if len(data) != 0 {
		f, err = mo.Load(data)
	} else {
		f, err = mo.LoadFile(name)
	}
	if err != nil {
		return nil, err
	}
	var tr = &translator{
		MessageMap: make(map[string]mo.Message),
	}
	for _, v := range f.Messages {
		tr.MessageMap[tr.makeMapKey(v.MsgContext, v.MsgId)] = v
	}
	if pluralFormHeader := f.MimeHeader.PluralForms; pluralFormHeader != "" {
		tr.PluralFormula, err = plural.Formula(pluralFormHeader)
		if err != nil {
			return nil, err
		}
	} else if lang := f.MimeHeader.Language; lang != "" {
		tr.PluralFormula = plural.FormulaByLang(lang)
	} else {
		tr.PluralFormula = plural.FormulaByLang("??")
	}
	return tr, nil
}

func newPoTranslator(name string, data []byte) (*translator, error) {
	var (
		f   *po.File
		err error
	)
	if len(data) != 0 {
		f, err = po.Load(data)
	} else {
		f, err = po.LoadFile(name)
	}
	if err != nil {
		return nil, err
	}
	var tr = &translator{
		MessageMap: make(map[string]mo.Message),
	}
	for _, v := range f.Messages {
		tr.MessageMap[tr.makeMapKey(v.MsgContext, v.MsgId)] = mo.Message{
			MsgContext:   v.MsgContext,
			MsgId:        v.MsgId,
			MsgIdPlural:  v.MsgIdPlural,
			MsgStr:       v.MsgStr,
			MsgStrPlural: v.MsgStrPlural,
		}
	}
	if pluralFormHeader := f.MimeHeader.PluralForms; pluralFormHeader != "" {
		tr.PluralFormula, err = plural.Formula(pluralFormHeader)
		if err != nil {
			return nil, err
		}
	} else if lang := f.MimeHeader.Language; lang != "" {
		tr.PluralFormula = plural.FormulaByLang(lang)
	} else {
		tr.PluralFormula = plural.FormulaByLang("??")
	}
	return tr, nil
}

func newJsonTranslator(lang, name string, jsonData []byte) (*translator, error) {
	var msgList []struct {
		MsgContext  string   `json:"msgctxt"`      // msgctxt context
		MsgId       string   `json:"msgid"`        // msgid untranslated-string
		MsgIdPlural string   `json:"msgid_plural"` // msgid_plural untranslated-string-plural
		MsgStr      []string `json:"msgstr"`       // msgstr translated-string
	}
	if err := json.Unmarshal(jsonData, &msgList); err != nil {
		return nil, err
	}

	var tr = &translator{
		MessageMap:    make(map[string]mo.Message),
		PluralFormula: plural.FormulaByLang(lang),
	}

	for _, v := range msgList {
		var v_MsgStr string
		var v_MsgStrPlural = v.MsgStr

		if len(v.MsgStr) != 0 {
			v_MsgStr = v.MsgStr[0]
		}

		tr.MessageMap[tr.makeMapKey(v.MsgContext, v.MsgId)] = mo.Message{
			MsgContext:   v.MsgContext,
			MsgId:        v.MsgId,
			MsgIdPlural:  v.MsgIdPlural,
			MsgStr:       v_MsgStr,
			MsgStrPlural: v_MsgStrPlural,
		}
	}
	return tr, nil
}

func (p *translator) PGettext(msgctxt, msgid string) string {
	return p.findMsgStr(msgctxt, msgid)
}

func (p *translator) PNGettext(msgctxt, msgid, msgidPlural string, n int) string {
	pluralFormN := p.PluralFormula(uint32(n))
	if ss := p.findMsgStrPlural(msgctxt, msgid); len(ss) != 0 {
		if pluralFormN >= len(ss) {
			pluralFormN = len(ss) - 1
		}
		if ss[pluralFormN] != "" {
			return ss[pluralFormN]
		}
	}
	pluralFormN = plural.FormulaByLang("en")(uint32(n))
	if msgidPlural != "" && pluralFormN > 0 {
		return msgidPlural
	}
	return msgid
}

func (p *translator) findMsgStr(msgctxt, msgid string) string {
	key := p.makeMapKey(msgctxt, msgid)
	if v, ok := p.MessageMap[key]; ok {
		if v.MsgStr != "" {
			return v.MsgStr
		}
	}
	return p.PNGettext(msgctxt, msgid, "", 1)
}

func (p *translator) findMsgStrPlural(msgctxt, msgid string) []string {
	key := p.makeMapKey(msgctxt, msgid)
	if v, ok := p.MessageMap[key]; ok {
		if len(v.MsgIdPlural) != 0 {
			if len(v.MsgStrPlural) != 0 {
				return v.MsgStrPlural
			} else {
				return nil
			}
		} else {
			if len(v.MsgStr) != 0 {
				return []string{v.MsgStr}
			} else {
				return nil
			}
		}
	}
	return nil
}

func (p *translator) makeMapKey(msgctxt, msgid string) string {
	if msgctxt != "" {
		return msgctxt + mo.EotSeparator + msgid
	}
	return msgid
}
