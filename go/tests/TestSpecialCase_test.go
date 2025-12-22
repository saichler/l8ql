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

// TestSpecialCase_test.go contains tests for special edge cases and
// integration with external types like L8PTarget.

import (
	"github.com/saichler/l8pollaris/go/types/l8tpollaris"
	"github.com/saichler/l8ql/go/gsql/interpreter"
	. "github.com/saichler/l8test/go/infra/t_resources"
	"github.com/saichler/l8types/go/ifs"
	"testing"
)

// TestSpecialCase tests query creation with external types and primary key decorators.
func TestSpecialCase(t *testing.T) {
	r, _ := CreateResources(25000, 2, ifs.Trace_Level)
	r.Introspector().Decorators().AddPrimaryKeyDecorator(&l8tpollaris.L8PTarget{}, "TargetId")
	gsql := "select * from L8PTarget where InventoryType=1 and (State=0 or State=1)"
	_, e := interpreter.NewQuery(gsql, r)
	if e != nil {
		Log.Fail(t, e)
		return
	}
}

// TestSortBy tests query creation with sort-by clause on nested properties.
func TestSortBy(t *testing.T) {
	r, _ := CreateResources(25000, 2, ifs.Trace_Level)
	r.Introspector().Decorators().AddPrimaryKeyDecorator(&l8tpollaris.L8PTarget{}, "TargetId")
	gsql := "select * from L8PTarget where InventoryType=1 and (State=0 or State=1) sort-by hosts.hostid"
	_, e := interpreter.NewQuery(gsql, r)
	if e != nil {
		Log.Fail(t, e)
		return
	}
}
