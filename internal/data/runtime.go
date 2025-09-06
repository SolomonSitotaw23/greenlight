package data

import (
	"fmt"
	"strconv"
)

type Runtime int32

// implement custom MarshalJSON method on the Runtime so it satisfies the json.Marshaler interface
func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)
	quotedJSONValue := strconv.Quote(jsonValue) //wrap in a quote before returning it

	return []byte(quotedJSONValue), nil
}
