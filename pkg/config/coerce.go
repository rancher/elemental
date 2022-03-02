/*
Copyright © 2022 SUSE LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"github.com/rancher/wrangler/pkg/data"
	"github.com/rancher/wrangler/pkg/data/convert"
	schemas2 "github.com/rancher/wrangler/pkg/schemas"
	"github.com/rancher/wrangler/pkg/schemas/mappers"
)

type Converter func(val interface{}) interface{}

type fieldConverter struct {
	mappers.DefaultMapper
	fieldName string
	converter Converter
}

func (f fieldConverter) ToInternal(data data.Object) error {
	val, ok := data[f.fieldName]
	if !ok {
		return nil
	}
	data[f.fieldName] = f.converter(val)
	return nil
}

type typeConverter struct {
	mappers.DefaultMapper
	converter Converter
	fieldType string
	mappers   schemas2.Mappers
}

func (t *typeConverter) ToInternal(data data.Object) error {
	return t.mappers.ToInternal(data)
}

func (t *typeConverter) ModifySchema(schema *schemas2.Schema, schemas *schemas2.Schemas) error {
	for name, field := range schema.ResourceFields {
		if field.Type == t.fieldType {
			t.mappers = append(t.mappers, fieldConverter{
				fieldName: name,
				converter: t.converter,
			})
		}
	}
	return nil
}

func NewTypeConverter(fieldType string, converter Converter) schemas2.Mapper {
	return &typeConverter{
		fieldType: fieldType,
		converter: converter,
	}
}

func NewToMap() schemas2.Mapper {
	return NewTypeConverter("map[string]", func(val interface{}) interface{} {
		if m, ok := val.(map[string]interface{}); ok {
			obj := make(map[string]string, len(m))
			for k, v := range m {
				obj[k] = convert.ToString(v)
			}
			return obj
		}
		return val
	})
}

func NewToSlice() schemas2.Mapper {
	return NewTypeConverter("array[string]", func(val interface{}) interface{} {
		if str, ok := val.(string); ok {
			return []string{str}
		}
		return val
	})
}

func NewToBool() schemas2.Mapper {
	return NewTypeConverter("boolean", func(val interface{}) interface{} {
		if str, ok := val.(string); ok {
			return str == "true" //nolint:goconst
		}
		return val
	})
}
