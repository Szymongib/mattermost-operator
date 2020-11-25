package mysql

import (
	mattermostv1alpha1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1alpha1"
	mattermostv1beta1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1"
	mattermostApp "github.com/mattermost/mattermost-operator/pkg/mattermost"
	mysqlOperator "github.com/presslabs/mysql-operator/pkg/apis/mysql/v1alpha1"

	componentUtils "github.com/mattermost/mattermost-operator/pkg/components/utils"

	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Cluster returns the MySQL cluster to deploy
func ClusterV1Beta(mattermost *mattermostv1beta1.Mattermost) *mysqlOperator.MysqlCluster {
	mysql := &mysqlOperator.MysqlCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      componentUtils.HashWithPrefix("db", mattermost.Name),
			Namespace: mattermost.Namespace,
			Labels:    mattermostv1alpha1.ClusterInstallationResourceLabels(mattermost.Name),
			OwnerReferences: mattermostApp.MattermostOwnerReference(mattermost),
		},
		Spec: mysqlOperator.MysqlClusterSpec{
			MysqlVersion: "5.7",
			Replicas:     mattermost.Spec.Database.OperatorManaged.Replicas,
			SecretName:   DefaultDatabaseSecretName(mattermost.Name),
			VolumeSpec: mysqlOperator.VolumeSpec{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						"ReadWriteOnce",
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse(mattermost.Spec.Database.OperatorManaged.StorageSize),
						},
					},
				},
			},
			BackupSchedule:           mattermost.Spec.Database.OperatorManaged.BackupSchedule,
			BackupURL:                mattermost.Spec.Database.OperatorManaged.BackupURL,
			BackupSecretName:         mattermost.Spec.Database.OperatorManaged.BackupRestoreSecretName,
			BackupRemoteDeletePolicy: mysqlOperator.DeletePolicy(mattermost.Spec.Database.OperatorManaged.BackupRemoteDeletePolicy),
		},
	}

	if mattermost.Spec.Database.OperatorManaged.InitBucketURL != "" && mattermost.Spec.Database.OperatorManaged.BackupRestoreSecretName != "" {
		mysql.Spec.InitBucketURL = mattermost.Spec.Database.OperatorManaged.InitBucketURL
		mysql.Spec.InitBucketSecretName = mattermost.Spec.Database.OperatorManaged.BackupRestoreSecretName
	}

	return mysql
}

//// DefaultDatabaseSecretName returns the default database secret name based on
//// the provided installation name.
//func DefaultDatabaseSecretName(installationName string) string {
//	return fmt.Sprintf("%s-mysql-root-password", installationName)
//}
