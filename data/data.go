package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
)

type UnmarshalFunc func([]byte, interface{}) error

type Data struct {
	Values map[string]interface{}
}

func NewData() *Data {
	return &Data{map[string]interface{}{}}
}

func DataFromObj(v interface{}) error {
	data := NewData()
	return data.UnmarshalJSON(v)
}

func UnmarshalJSONMap(m map[string]interface{}, v interface{}) error {
	// First marshal to JSON, then unmarshal into v
	byts, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("Unable to marshal existing values to JSON: %v", err)
	}
	if err = json.Unmarshal(byts, v); err != nil {
		return fmt.Errorf("Unable to unmarshal existing JSON to value: %v", err)
	}
	return nil
}

func (d *Data) UnmarshalJSON(v interface{}) error {
	return UnmarshalJSONMap(d.Values, v)
}

func (d *Data) ApplyTemplate(byts []byte) ([]byte, error) {
	return d.ApplyTemplateWithDelims(byts, "{{", "}}")
}

func (d *Data) ApplyTemplateWithDelims(byts []byte, leftDelim string, rightDelim string) ([]byte, error) {
	tmpl, err := template.New("data").Delims(leftDelim, rightDelim).Funcs(funcMap).Parse(string(byts))
	if err != nil {
		return nil, fmt.Errorf("Invalid template: %v", err)
	}
	tmplValues := map[string]interface{}{"prev": d.Values}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, tmplValues); err != nil {
		return nil, fmt.Errorf("Unable to execute template: %v", err)
	}
	return buf.Bytes(), nil
}

func (d *Data) ApplyTemplateAndJSONMerge(byts []byte) error {
	return d.ApplyTemplateAndMerge(byts, json.Unmarshal)
}

func (d *Data) ApplyTemplateAndMerge(byts []byte, unmarshal UnmarshalFunc) error {
	byts, err := d.ApplyTemplate(byts)
	if err != nil {
		return err
	}
	newValues := map[string]interface{}{}
	if err := unmarshal(byts, &newValues); err != nil {
		return fmt.Errorf("Unable to unmarshal resulting text: %v", err)
	}
	// Merge it
	applyMap(d.Values, newValues)
	return nil
}

func applyMap(existing map[string]interface{}, newValues map[string]interface{}) {
	for k, v := range newValues {
		if oldValue, ok := existing[k]; !ok {
			existing[k] = v
		} else {
			// Rules:
			//  arrays append to arrays
			//  maps append to maps
			//  everything else overrides
			switch oldTypedVal := oldValue.(type) {
			case map[string]interface{}:
				if newMap, ok := v.(map[string]interface{}); ok {
					applyMap(oldTypedVal, newMap)
				} else {
					existing[k] = v
				}
			case []interface{}:
				if newSlice, ok := v.([]interface{}); ok {
					existing[k] = append(oldTypedVal, newSlice)
				} else {
					existing[k] = v
				}
			default:
				existing[k] = v
			}
		}
	}
}

var funcMap = template.FuncMap{
	"hiddenPrompt": hiddenPrompt,
	"jsonString":   jsonString,
	"jsonVal":      jsonVal,
}

func ApplyTemplate(name string, byts []byte, v interface{}, funcs ...template.FuncMap) ([]byte, error) {
	newFuncMap := template.FuncMap{}
	for _, fmap := range append([]template.FuncMap{funcMap}, funcs...) {
		for key, val := range fmap {
			newFuncMap[key] = val
		}
	}
	tmpl, err := template.New(name).Funcs(newFuncMap).Parse(string(byts))
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
