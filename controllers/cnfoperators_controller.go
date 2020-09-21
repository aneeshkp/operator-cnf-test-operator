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

package controllers

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	testv1 "github.com/aneeshkp/operator-cnf-test-operator/api/v1"
	csv "github.com/aneeshkp/operator-cnf-test-operator/internal/csv/v1alpha1"
)

// CnfoperatorsReconciler reconciles a Cnfoperators object
type CnfoperatorsReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var csvInstance *csv.ClusterServiceVersion

// +kubebuilder:rbac:groups=test.cnf.operators.com,resources=cnfoperators,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=test.cnf.operators.com,resources=cnfoperators/status,verbs=get;update;patch

func (r *CnfoperatorsReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("cnfoperators", req.NamespacedName)

	// your logic here
	cnfOperator := &testv1.Cnfoperators{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, cnfOperator)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if err = loadCSV(r.Client, cnfOperator); err != nil {
		r.Log.Error(err, "Error Loading CSV files")
		cnfOperator.Status.CSV.Name = cnfOperator.Spec.CSVName
		csvResult := testv1.TestResult{}
		csvResult.Type = "CSV"
		if errors.IsNotFound(err) {
			csvResult.Status = testv1.ObjectStatusNotFound
		} else {
			csvResult.Status = testv1.ObjectStatusError
		}
		csvResult.Name = cnfOperator.Spec.CSVName
		cnfOperator.Status.CSV = csvResult
		r.Client.Status().Update(context.Background(), cnfOperator)
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CnfoperatorsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&testv1.Cnfoperators{}).
		Complete(r)
}
func loadCSV(rclient client.Client, instance *testv1.Cnfoperators) (err error) {

	csv := &csv.ClusterServiceVersion{}
	key := client.ObjectKey{Namespace: instance.Spec.OperatorNameSpace, Name: instance.Spec.CSVName}

	if err = rclient.Get(context.TODO(), key, csv); err == nil {
		csvInstance = csv
	}
	return err

}
