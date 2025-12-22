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

// Package tests contains unit tests for the L8QL parser and interpreter.
// It provides test utilities for creating queries, validating parsing,
// and checking query matching against test data structures.
package tests

import (
	"github.com/saichler/l8ql/go/gsql/interpreter"
	. "github.com/saichler/l8test/go/infra/t_resources"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/testtypes"
	"testing"
)

// createQuery is a test helper that creates an interpreted Query from a query string.
// It sets up the necessary resources and introspects the TestProto type.
func createQuery(query string) (*interpreter.Query, ifs.IResources, error) {
	r, _ := CreateResources(25000, 2, ifs.Trace_Level)
	r.Introspector().Inspect(&testtypes.TestProto{})
	q, e := interpreter.NewQuery(query, r)
	return q, r, e
}

// checkQuery is a test helper that validates query parsing.
// It creates a query and verifies whether an error occurred as expected.
func checkQuery(query string, expErr bool, t *testing.T) bool {
	q, _, e := createQuery(query)
	if e != nil && !expErr {
		Log.Fail(t, "Error creating query: ", e.Error())
		return false
	}
	if e == nil && expErr {
		Log.Fail(t, "Expected an error when creating a query")
		return false
	}
	if q == nil && e == nil {
		Log.Fail(t, "Query is nil")
		return false
	}
	return true
}

// checkMatch is a test helper that validates query matching.
// It creates a query and checks if it matches the given TestProto instance as expected.
func checkMatch(query string, pb *testtypes.TestProto, expectMatch bool, t *testing.T) bool {
	q, _, e := createQuery(query)
	if e != nil {
		Log.Fail(t, e)
		return false
	}
	if !q.Match(pb) && expectMatch {
		Log.Fail(t, "Expected a match")
		return false
	}
	if q.Match(pb) && !expectMatch {
		Log.Fail(t, "Expected no match")
		return false
	}
	return true
}
