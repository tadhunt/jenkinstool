package jenkinstool

import(
	"fmt"
	"reflect"
)

func String(v interface{}) string {
	if v == nil {
		return "<nil>"
	}

	s, isStringer := v.(fmt.Stringer)
	if isStringer {
		return s.String()
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		return String(rv.Elem().Interface())
	}

	return fmt.Sprintf("%v", v)
}
