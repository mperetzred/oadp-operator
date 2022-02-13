package lib

import (
	"context"
	"os"
	"path/filepath"
	"runtime"

	"github.com/apenella/go-ansible/pkg/options"
	"github.com/apenella/go-ansible/pkg/playbook"
)

const PLAYBOOKS_PATH = "/sample-applications/ansible"

var (
	_, b, _, _ = runtime.Caller(0)

	// Root folder of this project
	Root = filepath.Join(filepath.Dir(b), "../")
	_    = os.Setenv("ANSIBLE_CONFIG", Root+PLAYBOOKS_PATH)
)

type App interface {
	//Init(string, string)
	Cleanup() error
	Deploy() error
	Validate() error
}

var ansiblePlaybookConnectionOptions = &options.AnsibleConnectionOptions{
	Connection: "local",
}

type GenericApp struct {
	Name      string
	Namespace string
	ExtraVars map[string]interface{}
}

// func (a *GenericApp) Init(name string, namespace string) {
// 	a.Name = name
// 	a.Namespace = namespace
// 	a.ExtraVars = make(map[string]interface{})

// }

func (a *GenericApp) Deploy() error {
	return a.execAppPlaybook("with_deploy")
}

func (a *GenericApp) Cleanup() error {
	return a.execAppPlaybook("with_cleanup")
}

func (a *GenericApp) Validate() error {
	return a.execAppPlaybook("with_validate")
}

func (a GenericApp) execAppPlaybook(role string) error {

	m := map[string]interface{}{
		"use_role":  a.Name,
		"namespace": a.Namespace,
		role:        true,
	}

	if a.ExtraVars == nil {
		a.ExtraVars = make(map[string]interface{})
	}

	for k, v := range a.ExtraVars {
		m[k] = v
	}

	ansiblePlaybookOptions := &playbook.AnsiblePlaybookOptions{
		ExtraVars: m,
		Verbose:   true,
	}

	playbook := &playbook.AnsiblePlaybookCmd{
		Playbooks:         []string{Root + PLAYBOOKS_PATH + "/deploy-app.yml"},
		ConnectionOptions: ansiblePlaybookConnectionOptions,
		Options:           ansiblePlaybookOptions,
	}

	err := playbook.Run(context.TODO())
	return err

}

type AccessUrlApp struct {
	GenericApp
	ExpectedNumVisits int
}

func (au *AccessUrlApp) Validate() error {
	au.ExpectedNumVisits++
	if au.ExtraVars == nil {
		au.ExtraVars = make(map[string]interface{})
	}
	au.ExtraVars["expected_num_visits"] = au.ExpectedNumVisits
	return au.GenericApp.Validate()
}

// package lib

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"os"

// 	"github.com/onsi/ginkgo/v2"
// 	ocpappsv1 "github.com/openshift/api/apps/v1"
// 	security "github.com/openshift/api/security/v1"
// 	appsv1 "k8s.io/api/apps/v1"
// 	corev1 "k8s.io/api/core/v1"
// 	apierrors "k8s.io/apimachinery/pkg/api/errors"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
// 	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
// 	"k8s.io/apimachinery/pkg/util/wait"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// )

// func InstallApplication(ocClient client.Client, file string) error {
// 	template, err := os.ReadFile(file)
// 	if err != nil {
// 		return err
// 	}
// 	obj := &unstructured.UnstructuredList{}

// 	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
// 	_, _, err = dec.Decode([]byte(template), nil, obj)
// 	if err != nil {
// 		return err
// 	}
// 	for _, resource := range obj.Items {
// 		err = ocClient.Create(context.Background(), &resource)
// 		if apierrors.IsAlreadyExists(err) {
// 			continue
// 		} else if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func DoesSCCExist(ocClient client.Client, sccName string) (bool, error) {
// 	scc := security.SecurityContextConstraints{}
// 	err := ocClient.Get(context.Background(), client.ObjectKey{
// 		Name: sccName,
// 	}, &scc)
// 	if err != nil {
// 		return false, err
// 	}
// 	return true, nil

// }

// func UninstallApplication(ocClient client.Client, file string) error {
// 	template, err := os.ReadFile(file)
// 	if err != nil {
// 		return err
// 	}
// 	obj := &unstructured.UnstructuredList{}

// 	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
// 	_, _, err = dec.Decode([]byte(template), nil, obj)
// 	if err != nil {
// 		return err
// 	}
// 	for _, resource := range obj.Items {
// 		err = ocClient.Delete(context.Background(), &resource)
// 		if apierrors.IsNotFound(err) {
// 			continue
// 		} else if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func AreApplicationPodsRunning(namespace string) wait.ConditionFunc {
// 	return func() (bool, error) {
// 		clientset, err := setUpClient()
// 		if err != nil {
// 			return false, err
// 		}
// 		// select Velero pod with this label
// 		veleroOptions := metav1.ListOptions{
// 			LabelSelector: "e2e-app=true",
// 		}
// 		// get pods in test namespace with labelSelector
// 		podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), veleroOptions)
// 		if err != nil {
// 			return false, nil
// 		}
// 		if len(podList.Items) == 0 {
// 			return false, nil
// 		}
// 		// get pod name and status with specified label selector
// 		for _, podInfo := range podList.Items {
// 			phase := podInfo.Status.Phase
// 			if phase != corev1.PodRunning && phase != corev1.PodSucceeded {
// 				ginkgo.GinkgoWriter.Write([]byte(fmt.Sprintf("Pod %v not yet succeeded", podInfo.Name)))
// 				ginkgo.GinkgoWriter.Write([]byte(fmt.Sprintf("status: %v", podInfo.Status)))
// 				return false, nil
// 			}
// 		}
// 		return true, err
// 	}
// }

// func IsDCReady(ocClient client.Client, namespace, dcName string) wait.ConditionFunc {
// 	return func() (bool, error) {
// 		dc := ocpappsv1.DeploymentConfig{}
// 		err := ocClient.Get(context.Background(), client.ObjectKey{
// 			Namespace: namespace,
// 			Name:      dcName,
// 		}, &dc)
// 		if err != nil {
// 			return false, err
// 		}
// 		if dc.Status.AvailableReplicas != dc.Status.Replicas || dc.Status.Replicas == 0 {
// 			for _, condition := range dc.Status.Conditions {
// 				if len(condition.Message) > 0 {
// 					ginkgo.GinkgoWriter.Write([]byte(fmt.Sprintf("DC not available with condition: %s\n", condition.Message)))
// 				}
// 			}
// 			return false, errors.New("DC is not in a ready state")
// 		}
// 		return true, nil
// 	}
// }

// func IsDeploymentReady(ocClient client.Client, namespace, dName string) wait.ConditionFunc {
// 	return func() (bool, error) {
// 		deployment := appsv1.Deployment{}
// 		err := ocClient.Get(context.Background(), client.ObjectKey{
// 			Namespace: namespace,
// 			Name:      dName,
// 		}, &deployment)
// 		if err != nil {
// 			return false, err
// 		}
// 		if deployment.Status.AvailableReplicas != deployment.Status.Replicas || deployment.Status.Replicas == 0 {
// 			for _, condition := range deployment.Status.Conditions {
// 				if len(condition.Message) > 0 {
// 					ginkgo.GinkgoWriter.Write([]byte(fmt.Sprintf("deployment not available with condition: %s\n", condition.Message)))
// 				}
// 			}
// 			return false, errors.New("deployment is not in a ready state")
// 		}
// 		return true, nil
// 	}
// }

// func AreApplicationPodsRunning(namespace string) wait.ConditionFunc {
// 	return func() (bool, error) {
// 		clientset, err := setUpClient()
// 		if err != nil {
// 			return false, err
// 		}
// 		// select Velero pod with this label
// 		veleroOptions := metav1.ListOptions{
// 			LabelSelector: "e2e-app=true",
// 		}
// 		// get pods in test namespace with labelSelector
// 		podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), veleroOptions)
// 		if err != nil {
// 			return false, nil
// 		}
// 		if len(podList.Items) == 0 {
// 			return false, nil
// 		}
// 		// get pod name and status with specified label selector
// 		for _, podInfo := range podList.Items {
// 			phase := podInfo.Status.Phase
// 			if phase != corev1.PodRunning && phase != corev1.PodSucceeded {
// 				ginkgo.GinkgoWriter.Write([]byte(fmt.Sprintf("Pod %v not yet succeeded", podInfo.Name)))
// 				ginkgo.GinkgoWriter.Write([]byte(fmt.Sprintf("status: %v", podInfo.Status)))
// 				return false, nil
// 			}
// 		}
// 		return true, err
// 	}
// }
