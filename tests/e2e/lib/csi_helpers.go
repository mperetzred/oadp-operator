package lib

import (
	"context"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned/typed/volumesnapshot/v1"
	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetCsiDriversList() (*v1.CSIDriverList, error) {
	clientset, err := setUpClient()
	if err != nil {
		return nil, err
	}

	clientcsi, err := clientset.StorageV1().CSIDrivers().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return clientcsi, nil
}

func SetUpSnapshotClient() (*snapshotv1.SnapshotV1Client, error) {
	kubeConf := getKubeConfig()

	client, err := snapshotv1.NewForConfig(kubeConf)
	if err != nil {
		return nil, err
	}
	return client, nil
}
