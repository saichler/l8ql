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
package interpreter

// QuerySubsetDataCreator contains deprecated code for creating partial object clones
// with only the requested columns. This functionality has been superseded by
// the cloneOnlyWithColumns method in Query.

/*
func (query *Query) RequestedDataOnly(any interface{}) interface{} {
	if len(query.columns) == 0 {
		return any
	}
	value := reflect.ValueOf(any)
	result := query.newInstance(any)
	for _, attribute := range query.columns {
		attrValues := attribute.ValueOf(value)
		if attrValues != nil {
			for _, attrValue := range attrValues {
				attribute.SetValue(result, attrValue.Interface())
			}
		}
	}
	return result
}

func (query *Query) newInstance(any interface{}) interface{} {
	t := reflect.ValueOf(any).Elem().Type()
	return reflect.New(t).Interface()
}*/
