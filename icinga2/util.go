package icinga2

import (
	"encoding/json"
	"strconv"
)

func (s Service) MarshalJSON() ([]byte, error) {
	// Prevent json.Marshal() recursion
	type service Service

	svc := service(s)
	// Clear top-level Vars field, so it's not added to the marshalled JSON. We marshal all service variables into individual top-level `vars.<variable name>` fields below.

	svc.Vars = Vars{}

	serviceAsJson, err := json.Marshal(svc)
	if err != nil {
		return nil, err
	}

	var serviceAsMap map[string]interface{}
	if err := json.Unmarshal(serviceAsJson, &serviceAsMap); err != nil {
		return nil, err
	}

	// This loop flattens the json, so each var will be at the same level
	for k, v := range Flatten(s.Vars) {
		serviceAsMap["vars."+k] = v
	}

	return json.Marshal(serviceAsMap)
}

func Flatten(m map[string]interface{}) map[string]interface{} {
	flat := map[string]interface{}{}

	for k, v := range m {
		switch child := v.(type) {
		case map[string]interface{}:
			flat_child := Flatten(child)
			for ck, cv := range flat_child {
				flat[k+"."+ck] = cv
			}
		case []interface{}:
			for i := 0; i < len(child); i++ {
				flat[k+"["+strconv.Itoa(i)+"]"] = child[i]
			}
		default:
			flat[k] = v
		}
	}

	return flat
}
