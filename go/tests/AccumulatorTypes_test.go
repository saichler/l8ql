/*
© 2025 Sharon Aicler (saichler@gmail.com)

Layer 8 Ecosystem is licensed under the Apache License, Version 2.0.
You may obtain a copy of the License at:

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package tests

import (
	"testing"

	"github.com/saichler/l8ql/go/gsql/interpreter"
)

// TestAccumulatorUnsignedAndSmallInts checks that sum/avg/min/max count every
// integer kind: protobuf uint32/uint64 fields (e.g. a GPU count) used to be
// skipped, so their sum was 0.
func TestAccumulatorUnsignedAndSmallInts(t *testing.T) {
	cases := []struct {
		name   string
		values []interface{}
	}{
		{"uint32", []interface{}{uint32(8), uint32(8), uint32(8)}},
		{"uint64", []interface{}{uint64(8), uint64(8), uint64(8)}},
		{"uint", []interface{}{uint(8), uint(8), uint(8)}},
		{"int8", []interface{}{int8(8), int8(8), int8(8)}},
		{"int16", []interface{}{int16(8), int16(8), int16(8)}},
		{"uint8", []interface{}{uint8(8), uint8(8), uint8(8)}},
		{"uint16", []interface{}{uint16(8), uint16(8), uint16(8)}},
		{"mixed", []interface{}{uint32(4), int64(8), uint64(12)}},
	}
	for _, c := range cases {
		want := map[string]float64{"sum": 24, "avg": 8, "min": 8, "max": 8}
		if c.name == "mixed" {
			want["min"], want["max"] = 4, 12
		}
		for fn, w := range want {
			a := interpreter.NewAccumulator(fn)
			for _, v := range c.values {
				a.Add(v)
			}
			if got, _ := a.Result().(float64); got != w {
				t.Errorf("%s(%s) = %v, want %v", fn, c.name, a.Result(), w)
			}
		}
	}
}

// TestAccumulatorIgnoresNonNumbers keeps strings and nil out of the sum while
// still counting the records.
func TestAccumulatorIgnoresNonNumbers(t *testing.T) {
	sum := interpreter.NewAccumulator("sum")
	count := interpreter.NewAccumulator("count")
	for _, v := range []interface{}{uint32(5), "x", nil, uint64(7)} {
		sum.Add(v)
		count.Add(v)
	}
	if got := sum.Result(); got != float64(12) {
		t.Errorf("sum = %v, want 12", got)
	}
	if got := count.Result(); got != int64(4) {
		t.Errorf("count = %v, want 4", got)
	}
}
