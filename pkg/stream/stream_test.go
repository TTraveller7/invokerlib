package stream

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_streamBp_Map(t *testing.T) {
	bp := &streamBp[int]{}
	nextBp := Map[int, string](bp, func(i int) string {
		return fmt.Sprintf("%v", i)
	})
	t.Logf("%v, %v\n", reflect.TypeOf(bp).Kind(), reflect.TypeOf(nextBp.lbp).Kind())
}
