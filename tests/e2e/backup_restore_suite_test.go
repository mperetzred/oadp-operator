package e2e_test

import (
	"log"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/openshift/oadp-operator/tests/e2e/lib"
	utils "github.com/openshift/oadp-operator/tests/e2e/utils"

	//. "github.com/openshift/oadp-operator/tests/v1/apps"
	velero "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type VerificationFunction func(client.Client, string) error

var _ = Describe("AWS backup restore tests", func() {
	var currentBackup BackupInterface
	var currentApp App

	var _ = BeforeEach(func() {
		testSuiteInstanceName := "ts-" + instanceName
		dpaCR.Name = testSuiteInstanceName

		credData, err := utils.ReadFile(cloud)
		Expect(err).NotTo(HaveOccurred())
		err = CreateCredentialsSecret(credData, namespace, GetSecretRef(credSecretRef))
		Expect(err).NotTo(HaveOccurred())
	})

	var _ = AfterEach(func() {
		err := dpaCR.Delete()
		Expect(err).ToNot(HaveOccurred())
		log.Printf("Cleaning resources")
		if currentBackup != nil {
			err = currentBackup.CleanBackup()
			Expect(err).ToNot(HaveOccurred())
		}
		log.Printf("Cleaning app")
		if currentApp != nil {
			err = currentApp.Cleanup()
			Expect(err).ToNot(HaveOccurred())
		}

	})

	type BackupRestoreCase struct {
		Name          string
		BackupSpec    velero.BackupSpec
		MaxK8SVersion *K8sVersion
		MinK8SVersion *K8sVersion
	}

	DescribeTable("backup and restore applications",
		func(brCase BackupRestoreCase, backup BackupInterface, app App, expectedErr error) {

			err := dpaCR.Build(backup.GetType())
			Expect(err).NotTo(HaveOccurred())

			err = dpaCR.CreateOrUpdate(&dpaCR.CustomResource.Spec)
			Expect(err).NotTo(HaveOccurred())

			log.Printf("Waiting for velero pod to be running")
			Eventually(AreVeleroPodsRunning(namespace), timeoutMultiplier*time.Minute*3, time.Second*5).Should(BeTrue())

			if dpaCR.CustomResource.Spec.BackupImages == nil || *dpaCR.CustomResource.Spec.BackupImages {
				log.Printf("Waiting for registry pods to be running")
				Eventually(AreRegistryDeploymentsAvailable(namespace), timeoutMultiplier*time.Minute*3, time.Second*5).Should(BeTrue())
			}
			if notVersionTarget, reason := NotServerVersionTarget(brCase.MinK8SVersion, brCase.MaxK8SVersion); notVersionTarget {
				Skip(reason)
			}

			brCaseName := brCase.Name
			backup.NewBackup(dpaCR.Client, brCaseName, &brCase.BackupSpec)
			backupRestoreName := backup.GetBackupSpec().Name
			err = backup.PrepareBackup()
			Expect(err).ToNot(HaveOccurred())
			currentBackup = backup

			// install app
			log.Printf("Installing application for case %s", brCaseName)
			Expect(app.Deploy()).ToNot(HaveOccurred())
			currentApp = app
			err = app.Validate()
			Expect(err).ToNot(HaveOccurred())

			// create backup
			log.Printf("Creating backup %s for case %s", backupRestoreName, brCaseName)
			err = backup.CreateBackup()
			Expect(err).ToNot(HaveOccurred())

			// wait for backup to not be running
			Eventually(backup.IsBackupDone(), timeoutMultiplier*time.Minute*4, time.Second*10).Should(BeTrue())
			Expect(GetVeleroContainerFailureLogs(dpaCR.Namespace)).To(Equal([]string{}))

			// check if backup succeeded
			succeeded, err := backup.IsBackupCompletedSuccessfully()
			Expect(err).ToNot(HaveOccurred())
			Expect(succeeded).To(Equal(true))
			log.Printf("Backup for case %s succeeded", brCaseName)

			// uninstall app
			log.Printf("Uninstalling app for case %s", brCaseName)
			Expect(app.Cleanup()).ToNot(HaveOccurred())

			// run restore
			log.Printf("Creating restore %s for case %s", backupRestoreName, brCaseName)
			err = CreateRestoreFromBackup(dpaCR.Client, namespace, backupRestoreName, backupRestoreName)
			Expect(err).ToNot(HaveOccurred())
			Eventually(IsRestoreDone(dpaCR.Client, namespace, backupRestoreName), timeoutMultiplier*time.Minute*4, time.Second*10).Should(BeTrue())
			Expect(GetVeleroContainerFailureLogs(dpaCR.Namespace)).To(Equal([]string{}))

			// Check if restore succeeded
			succeeded, err = IsRestoreCompletedSuccessfully(dpaCR.Client, namespace, backupRestoreName)
			Expect(err).ToNot(HaveOccurred())
			Expect(succeeded).To(Equal(true))
			err = app.Validate()
			Expect(err).ToNot(HaveOccurred())

		},
		Entry("MySQL application CSI", Label("aws"),
			BackupRestoreCase{
				Name: "mysql-persistent",
				BackupSpec: velero.BackupSpec{
					IncludedNamespaces: []string{"mysql-persistent"},
				},
			},
			&BackupCsi{},
			&GenericApp{
				Name:      "ocp-mysql",
				Namespace: "mysql-persistent",
			}, nil),
		Entry("MSSQL application",
			BackupRestoreCase{
				Name: "mssql-persistent",
				BackupSpec: velero.BackupSpec{
					IncludedNamespaces: []string{"mssql-persistent"},
				},
			},
			&BackupVsl{CreateFromDpa: false},
			&GenericApp{
				Name:      "ocp-mssql",
				Namespace: "mssql-persistent",
			}, nil),
		FEntry("Django application",
			BackupRestoreCase{
				Name: "django-persistent",
				BackupSpec: velero.BackupSpec{
					IncludedNamespaces: []string{"django-persistent"},
				},
			},
			&BackupCsi{},
			&AccessUrlApp{
				GenericApp: GenericApp{
					Name:      "ocp-django",
					Namespace: "django-persistent",
				},
			}, nil),

		// 	Entry("MySQL application CSI", Label("aws"), BackupRestoreCase{

		// 		ApplicationTemplate: "./sample-applications/mysql-persistent/mysql-persistent-csi-template.yaml",
		// 		Name:                "mysql-persistent",
		// 		BackupSpec: velero.BackupSpec{
		// 			IncludedNamespaces: []string{"mysql-persistent"},
		// 		},
		// 		PreBackupVerify:   mysqlReady,
		// 		PostRestoreVerify: mysqlReady,
		// 	}, &BackupCsi{}, nil),
		// 	Entry("Parks application <4.8.0", BackupRestoreCase{
		// 		ApplicationTemplate: "./sample-applications/parks-app/manifest.yaml",
		// 		Name:                "parks-app",
		// 		BackupSpec: velero.BackupSpec{
		// 			IncludedNamespaces: []string{"parks-app"},
		// 		},
		// 		PreBackupVerify:   parksAppReady,
		// 		PostRestoreVerify: parksAppReady,
		// 		MaxK8SVersion:     &K8sVersionOcp47,
		// 	}, &BackupVsl{CreateFromDpa: false},
		// 		nil),
		// 	Entry("MySQL application", BackupRestoreCase{
		// 		ApplicationTemplate: "./sample-applications/mysql-persistent/mysql-persistent-template.yaml",
		// 		Name:                "mysql-persistent",
		// 		PreBackupVerify:     mysqlReady,
		// 		PostRestoreVerify:   mysqlReady,
		// 		BackupSpec: velero.BackupSpec{
		// 			IncludedNamespaces: []string{"mysql-persistent"},
		// 		},
		// 	}, &BackupRestic{}, nil),
		// 	Entry("Parks application >=4.8.0", BackupRestoreCase{
		// 		ApplicationTemplate: "./sample-applications/parks-app/manifest4.8.yaml",
		// 		Name:                "parks-app",
		// 		BackupSpec: velero.BackupSpec{
		// 			IncludedNamespaces: []string{"parks-app"},
		// 		},
		// 		PreBackupVerify:   parksAppReady,
		// 		PostRestoreVerify: parksAppReady,
		// 		MinK8SVersion:     &K8sVersionOcp48,
		// 	}, &BackupVsl{
		// 		CreateFromDpa: true,
		// 	}, nil),
	)
})
