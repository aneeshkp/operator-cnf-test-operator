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
// +kubebuilder:rbac:groups=operators.coreos.com,resources=clusterserviceversions,verbs=get;list;watch
// +kubebuilder:rbac:groups=operators.coreos.com,resources=clusterserviceversions/status,verbs=get
// +kubebuilder:rbac:groups="",resources=pods;services;endpoints;deployment;configmaps;daemonset;nodes,verbs=get;list
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list
// +kubebuilder:rbac:groups="*",resources="*",verbs=watch;get;list

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
	r.Log.Info("The CRD what we found is ", "CRD", cnfOperator)
	r.Log.Info("The Name of the CRD is ", "Name", cnfOperator.Name)
	r.Log.Info("The PCSV of the CRD is ", "Name", cnfOperator.Spec.CSVName)
	// reset all status
	initStatus(cnfOperator)
	//func (r *CnfoperatorsReconciler) getCSV(namespace string, name string, group string, version string, resource string) (e error, csv *v1alpha1.ClusterServiceVersion) {
	if err, csvInstance = r.getCSV(cnfOperator.Spec.CSVNamespace, cnfOperator.Spec.CSVName, v1alpha1.GroupName, v1alpha1.GroupVersion, "clusterserviceversions"); err != nil {
		r.Log.Error(err, "Error Loading CSV files")
		cnfOperator.Status.CSV.Name = cnfOperator.Spec.CSVName
		if errors.IsNotFound(err) {
			cnfOperator.Status.CSV.Status = "Not Found"
		} else {
			cnfOperator.Status.CSV.Status = "Error"
		}
		cnfOperator.Status.CSV.Name = cnfOperator.Spec.CSVName
		r.Log.Info("Updating status ")
		errStatusUpdate := r.Client.Status().Update(context.Background(), cnfOperator)
		if err != nil {
			r.Log.Error(errStatusUpdate, "Updating status err")
		}
		return reconcile.Result{}, err
	} else {
		cnfOperator.Status.CSV.Name = cnfOperator.Spec.CSVName
		cnfOperator.Status.CSV.Status = csvInstance.Status.Phase
		var rStatus []v1alpha1.RequirementStatus
		for _, value := range csvInstance.Status.RequirementStatus {
			if value.Status != v1alpha1.RequirementStatusReasonPresent &&
				value.Status != v1alpha1.DependentStatusReasonSatisfied {
				rStatus = append(rStatus, value)
			}
		}
		cnfOperator.Status.CSV.CSVRequirementStatus = rStatus
		r.Log.Info("Updating status ")
		errStatusUpdate := r.Client.Status().Update(context.Background(), cnfOperator)

		if err != nil {
			r.Log.Error(errStatusUpdate, "Updating status err")
			return reconcile.Result{}, errStatusUpdate
		}

	}
	// Check for owned CRDS
	if len(csvInstance.Spec.CustomResourceDefinitions.Owned) > 0 {
		for _, owned := range csvInstance.Spec.CustomResourceDefinitions.Owned {
			parts := strings.SplitN(owned.Name, ".", 2)
			e, name := r.getCrdInstance(cnfOperator.Spec.CRNamespace, owned.Name, parts[1], owned.Version, parts[0])
			if e == nil {
				cnfOperator.Status.CRDS[parts[0]] = name
			} else {
				cnfOperator.Status.CRDS[parts[0]] = e.Error()
			}
		}
		r.Log.Info("Updating CRD status")
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
	status.CSV = testv1.CSVTestResult{Name: "", Type: "CSV", Status: ""}
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
		r.Log.Error(errCrd, "Error getting CRD")
		return
	}
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
	r.Log.Info("Reading following CR ", "Name", name, "resource", resource, "Group", group, "version", version)
	crd, errCrd := crdClient.Namespace(namespace).List(context.Background(), metav1.ListOptions{})
	//crd, errCrd := crdClient.Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	//	crd, errCrd := crdClient.Namespace("openshift-machine-api").List( metav1.ListOptions{})
	if errCrd != nil || len(crd.Items) < 1 {
		r.Log.Error(errCrd, "Error getting CRD")
		return errCrd, ""
	}
	r.Log.Info("Got CRD ", "name", crd.Items[0].GetName())

	return errCrd, crd.Items[0].GetName()
}

//requirementStatus:
