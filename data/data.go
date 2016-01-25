package data
import (
	"html/template"
	"fmt"
	"bytes"
	"encoding/json"
)

type UnmarshalFunc func([]byte, interface{}) error

type Data struct {
	Values map[string]interface{}
}

func NewData() *Data {
	return &Data{map[string]interface{}{}}
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

func (d *Data) ApplyTemplateAndJSONMerge(byts []byte) error {
	return d.ApplyTemplateAndMerge(byts, json.Unmarshal)
}

func (d *Data) ApplyTemplateAndMerge(byts []byte, unmarshal UnmarshalFunc) error {
	// First run template
	tmpl, err := template.New("data").Parse(string(byts))
	if err != nil {
		return fmt.Errorf("Invalid template: %v", err)
	}
	tmplValues := map[string]interface{}{"prev": d.Values}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, tmplValues); err != nil {
		return fmt.Errorf("Unable to execute template: %v", err)
	}
	newValues := map[string]interface{}{}
	if err := unmarshal(buf.Bytes(), &newValues); err != nil {
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
