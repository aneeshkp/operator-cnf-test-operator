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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	testv1 "github.com/aneeshkp/operator-cnf-test-operator/api/v1"
	v1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmcli "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
)

// CnfoperatorsReconciler reconciles a Cnfoperators object
type CnfoperatorsReconciler struct {
	client.Client
	Config *rest.Config
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var csvInstance *v1alpha1.ClusterServiceVersion

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
	// reset all status
	initStatus(cnfOperator)
	//func (r *CnfoperatorsReconciler) getCSV(namespace string, name string, group string, version string, resource string) (e error, csv *v1alpha1.ClusterServiceVersion) {
	if err, csvInstance = r.getCSV(cnfOperator.Spec.CSVNamespace, cnfOperator.Spec.CSVName, v1alpha1.GroupName, v1alpha1.GroupVersion, "clusterserviceversions"); err != nil {
		//if err = loadCSV(r.Client, cnfOperator); err != nil {
		r.Log.Error(err, "Error Loading CSV files")
		r.Log.Info("Error Loading csv")
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
		r.Log.Info("Updtaing status ")
		err = r.Client.Status().Update(context.Background(), cnfOperator)
		if err != nil {
			r.Log.Error(err, "Updating status err")
		}
		return reconcile.Result{}, err
	}

	if len(csvInstance.Spec.CustomResourceDefinitions.Owned) > 0 {
		for _, owned := range csvInstance.Spec.CustomResourceDefinitions.Owned {
			parts := strings.SplitN(owned.Name, ".", 2)
			e, name := r.getCrdInstance(cnfOperator.Spec.CRNamespace, owned.Name, parts[1], owned.Version, parts[0])
			if e == nil {
				cnfOperator.Status.CRDS[parts[0]] = name
			}

		}
		err = r.Client.Status().Update(context.Background(), cnfOperator)
		if err != nil {
			r.Log.Error(err, "Updating status err")
		}

	}

	return ctrl.Result{}, nil
}

func (r *CnfoperatorsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&testv1.Cnfoperators{}).
		Complete(r)
}
func loadCSV(rclient client.Client, instance *testv1.Cnfoperators) (err error) {

	csv := &v1alpha1.ClusterServiceVersion{}
	key := client.ObjectKey{Namespace: instance.Spec.CSVNamespace, Name: instance.Spec.CSVName}

	if err = rclient.Get(context.TODO(), key, csv); err == nil {
		csvInstance = csv
	}

	return err

}
func initStatus(cnfOperator *testv1.Cnfoperators) {
	status := testv1.CnfoperatorsStatus{}
	status.CSV = testv1.TestResult{Name: "", Type: "CSV", Status: ""}
	status.CRDS = make(map[string]string)
	status.Deployment = testv1.TestResult{Name: "", Type: "deployment", Status: ""}
	status.Operators = testv1.TestResult{Name: "", Type: "Operators", Status: ""}
	status.Operands = []testv1.TestResult{}
	status.PodNames = []string{}
	cnfOperator.Status = status

}

func (r *CnfoperatorsReconciler) getCSV(namespace string, name string, group string, version string, resource string) (e error, csv *v1alpha1.ClusterServiceVersion) {

	//var olClientset olmcli.Clientset
	/*customGVR := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}*/
	//v1alpha1.SchemeBuilder.GroupVersion.WithResource(resource)

	cvsClient, errClient := olmcli.NewForConfig(r.Config)
	//dynClient, errClient := dynamic.NewForConfig(r.Config)
	if errClient != nil {
		r.Log.Info("Error received creating client ")
		return errClient, nil
	}
	csv, errCrd := cvsClient.OperatorsV1alpha1().ClusterServiceVersions(namespace).Get(context.Background(), name, metav1.GetOptions{})
	//crdClient := dynClient.Resource(customGVR)
	/*resp, err := dynClient.Resource(resourceScheme).
		Namespace(namespace).
		Get(context.Background(),name, metav1.GetOptions{})
	assertNoError(err)*/

	//unCsv, errCrd := crdClient.Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if errCrd != nil {
		r.Log.Error(errCrd, "Error getting CRD %v")
		return
	}
	r.Log.Info("Got CRD: ")
	r.Log.Info("Got CRD: %v", csv)
	//unstructured := unCsv.UnstructuredContent()
	//runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &csv)
	//	crd, errCrd := crdClient.Namespace("openshift-machine-api").List( metav1.ListOptions{})
	return
}

func (r *CnfoperatorsReconciler) getCrdInstance(namespace string, name string, group string, version string, resource string) (error, string) {

	customGVR := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
	dynClient, errClient := dynamic.NewForConfig(r.Config)
	if errClient != nil {
		r.Log.Info("Error received creating client ")
		return errClient, ""
	}
	crdClient := dynClient.Resource(customGVR)

	crd, errCrd := crdClient.Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	//	crd, errCrd := crdClient.Namespace("openshift-machine-api").List( metav1.ListOptions{})

	if errCrd != nil {

		r.Log.Error(errCrd, "Error getting CRD %v")
	}
	r.Log.Info("Got CRD: %v", crd)
	return errCrd, crd.GetName()
}
