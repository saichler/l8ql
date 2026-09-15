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

	. "github.com/saichler/l8test/go/infra/t_resources"
)

// TestHashIncludesAAAId verifies that Query.Hash() folds AAAId into the
// same hash as the query text, so two otherwise-identical queries issued
// by different callers (or one with no AAAId at all) never produce the
// same Hash(). Regression test for a bug where a cache keyed only on
// Hash() (l8utils/go/utils/cache/internalCache.go) combined it with a
// separately-computed hash of AAAId that could sign-extend and clobber
// Hash()'s own bits, collapsing distinct queries under the same AAAId onto
// the same cache bucket.
func TestHashIncludesAAAId(t *testing.T) {
	q1, _, e := createQuery("select * from testproto")
	if e != nil {
		Log.Fail(t, "Error creating query: ", e.Error())
		return
	}
	q1.SetAAAId("11111111-1111-1111-1111-111111111111")
	h1 := q1.Hash()

	q2, _, e := createQuery("select * from testproto")
	if e != nil {
		Log.Fail(t, "Error creating query: ", e.Error())
		return
	}
	q2.SetAAAId("22222222-2222-2222-2222-222222222222")
	h2 := q2.Hash()

	q3, _, e := createQuery("select * from testproto")
	if e != nil {
		Log.Fail(t, "Error creating query: ", e.Error())
		return
	}
	h3 := q3.Hash() // no AAAId at all

	if h1 == h2 {
		Log.Fail(t, "expected different AAAId to change Hash(), got the same value for both")
		return
	}
	if h1 == h3 || h2 == h3 {
		Log.Fail(t, "expected a query with an AAAId to hash differently than one without")
	}
}
