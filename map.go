package grout

import (
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
