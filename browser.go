package jsonbrowser

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
)

var NullBrowser = &JSONBrowser{node: nil}

type JSONBrowser struct {
	node any
}

func NewJSONBrowser(node any) *JSONBrowser {
	if node == nil {
		return NullBrowser
	}
	return &JSONBrowser{node: node}
}

func (jb *JSONBrowser) IsList() bool {
	_ , ok := jb.node.([]any)
	return  ok
}

func (jb *JSONBrowser) IsMap() bool {
	_ , ok := jb.node.(map[string]any)
	return  ok
}

func (jb *JSONBrowser) Index(i int) *JSONBrowser {
	if arr , ok := jb.node.([]any); ok {
		if i >= 0 && i < len(arr) {
			return NewJSONBrowser(arr[i])
		}
	}
	return NullBrowser
}

func (jb *JSONBrowser) Get(key string) *JSONBrowser {
	if m , ok := jb.node.(map[string]any); ok {
		return NewJSONBrowser(m[key])
	}
	return NullBrowser
}

func (jb *JSONBrowser) Put(key string , value any) error {
	m , ok := jb.node.(map[string]any)
	if !ok {
		return fmt.Errorf("put only works on maps")	
	}
	
	if browser , ok := value.(*JSONBrowser); ok {
		m[key] = browser.node
	} else {
		m[key] = value
	}
	return nil
}

func (jb *JSONBrowser) Remove(key string) error {
	m , ok := jb.node.(map[string]any)
	if !ok {
		return fmt.Errorf("remove only works on maps")	
	}
	delete(m , key)
	return nil
}

func (jb *JSONBrowser) Add(value any) error {
	arr , ok := jb.node.([]any)
	if !ok {
		return fmt.Errorf("add only works on lists")	
	}
	
	if browser , ok := value.(*JSONBrowser); ok {
		jb.node = append(arr , browser.node)
	} else {
		jb.node = append(arr , value)
	}
	return nil
}

func (jb *JSONBrowser) Set(index int , value any) error {
	arr , ok := jb.node.([]any)
	if !ok {
		return fmt.Errorf("set only works on lists")	
	}
	if index < 0 || index >= len(arr) {
		return fmt.Errorf("index out of bounds")	
	}
	
	if browser , ok := value.(*JSONBrowser); ok {
		arr[index] = browser.node
	} else {
		arr[index] = value
	}
	return nil
}

func (jb *JSONBrowser) RemoveAt(index int) error {
	arr , ok := jb.node.([]any)
	if !ok {
		return fmt.Errorf("removeAt only works on lists")
	} 
	
	if index < 0 || index >= len(arr) {
		return fmt.Errorf("index out of bounds")	
	}
	
	jb.node = append(arr[:index] , arr[index+1:]...)
	return nil
}

func (jb *JSONBrowser) Values() []*JSONBrowser {
	values := make([]*JSONBrowser , 0)
	
	if arr , ok := jb.node.([]any) ; ok {
		for _ , v := range arr {
			values = append(values , NewJSONBrowser(v))
		}
	} else if m , ok := jb.node.(map[string]any) ; ok {
		for _ , v := range m {
			values = append(values , NewJSONBrowser(v))
		}	
	}
	
	return values
}

func (jb *JSONBrowser) Keys() []string {
	if !jb.IsMap() {
		return []string{}
	}
	
	m , _ := jb.node.(map[string]any)
	return slices.Collect(maps.Keys(m))
}

func (jb *JSONBrowser) As(target any) error {
	data , err := json.Marshal(jb.node)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(data , target)
}

func (jb *JSONBrowser) Text() *string {
	if jb.node == nil {
		return nil
	}
	
	switch v := jb.node.(type) {
	case string:
		return &v
	case bool:
		s := strconv.FormatBool(v)
		return &s
	case float64:
		s := strconv.FormatFloat(v, 'f', -1, 64)
		return &s
	case int:
		s := strconv.Itoa(v)
		return &s
	case int64:
		s := strconv.FormatInt(v, 10)
		return &s
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		s := string(bytes)
		return &s
	}
}

func (jb *JSONBrowser) IsNull() bool {
	return jb.node == nil || jb == NullBrowser
}

func (jb *JSONBrowser) TextOrDefault(deafult string) string {
	text := jb.Text()
	if text == nil {
		return deafult
	}
	return *text
}

func (jb *JSONBrowser) AsBoolean(defaultValue bool) bool {
	if jb.node == nil {
		return defaultValue
	}
	
	switch v := jb.node.(type) {
	case bool:
		return v
	case string:
	    b , err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	
	return defaultValue
}

func (jb *JSONBrowser) AsLong(defaultValue int64) int64 {
	if jb.node == nil {
		return defaultValue
	}
	
	switch v := jb.node.(type) {
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case int64:
		return v
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
	}
	
	return defaultValue
}

func (jb *JSONBrowser) AsInt(defaultValue int) int {
	if jb.node == nil {
		return defaultValue
	}
	
	switch v := jb.node.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	
	return defaultValue
}

func (jb *JSONBrowser) SafeText() string {
	return jb.TextOrDefault("")
}


func (jb *JSONBrowser) JsonString() *string {
	if jb.node == nil {
		return nil
	}
	
	bytes , err := json.Marshal(jb.node)
	if err != nil {
		return nil
	}
	
	s := string(bytes)
	return &s
}


func ParseJsonString(jsonStr string) (*JSONBrowser , error) {
	var node any
	if err := json.Unmarshal([]byte(jsonStr) , &node); err != nil {
		return nil , err
	}
	return NewJSONBrowser(node) , nil
}

func ParseJsonBytes(jsonBytes []byte) (*JSONBrowser , error) {
	var node any
	if err := json.Unmarshal(jsonBytes , &node); err != nil {
		return nil , err
	}
	return NewJSONBrowser(node) , nil
}

func ParseReader(reader io.Reader) (*JSONBrowser , error) {
	var node any
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	if err := decoder.Decode(&node); err != nil {
		return nil , err
	}
	return NewJSONBrowser(node) , nil
}

func NewEmptyMap() *JSONBrowser {
	return NewJSONBrowser(make(map[string]any))
}

func NewEmptyList() *JSONBrowser {
	return NewJSONBrowser(make([]any , 0))
}