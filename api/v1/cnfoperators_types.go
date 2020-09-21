/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CnfoperatorsSpec defines the desired state of Cnfoperators
type CnfoperatorsSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Cnfoperators. Edit Cnfoperators_types.go to remove/update
	CSVName           string `json:"csvname,omitempty"`
	OperatorNameSpace string `json:"namespace,omitempty"`
}

type ObjectStatus string

// Const for ObjectStatus
const (
	ObjectStatusNotFound ObjectStatus = "NotFound"
	ObjectStatusRunning               = "Running"
	ObjectStatusError                 = "Error"
)

type TestResult struct {
	Type   string       `json:"type,omitempty"`
	Name   string       `json:"name,omitempty"`
	Status ObjectStatus `json:"status,omitempty"`
}

// CnfoperatorsStatus defines the observed state of Cnfoperators
type CnfoperatorsStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	CSV        TestResult        `json:"csv"`
	Deployment TestResult        `json:"deployment"`
	Operators  TestResult        `json:"operators"`
	Operands   []TestResult      `json:"operands"`
	CRDS       map[string]string `json:"crds"`
	PodNames   []string          `json:"pods"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Cnfoperators is the Schema for the cnfoperators API
type Cnfoperators struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CnfoperatorsSpec   `json:"spec,omitempty"`
	Status CnfoperatorsStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CnfoperatorsList contains a list of Cnfoperators
type CnfoperatorsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cnfoperators `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cnfoperators{}, &CnfoperatorsList{})
}
