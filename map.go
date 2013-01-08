package grout

import (
	"fmt"
	"strings"
)

type M map[string]interface{}

func (m M) Map(path string) M {
	val, ok := m.get(path).(M)
	if !ok {
		return nil
	}
	return val
}

func (m M) String(path string, def string) string {
	val, ok := m.get(path).(string)
	if !ok {
		return def
	}
	return val
}

func (m M) Int(path string, def int) int {
	val, ok := m.get(path).(int)
	if !ok {
		return def
	}
	return val
}

func (m M) get(path string) interface{} {
	parts := strings.Split(path, "/")
	for i, p := range parts {
		val, ok := m[p]
		if !ok {
			break
		}

		if i < len(parts)-1 {
			m, ok = val.(M)
			if !ok {
				break
			}
		} else {
			return val
		}
	}
	return nil
}

func (m M) sanitize() {
	for k, v := range m {
		vmap, ok := v.(map[interface{}]interface{})
		if !ok {
			continue
		}

		newv := make(M, len(vmap))
		for vk, vv := range vmap {
			s := fmt.Sprintf("%v", vk)
			newv[s] = vv
		}
		newv.sanitize()
		m[k] = newv
	}
}
