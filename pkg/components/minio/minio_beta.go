package minio

import (
	"fmt"
	mattermostv1beta1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1"

	mattermostv1alpha1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1alpha1"
	"github.com/mattermost/mattermost-operator/pkg/components/utils"
	mattermostApp "github.com/mattermost/mattermost-operator/pkg/mattermost"

	minioOperator "github.com/minio/minio-operator/pkg/apis/miniocontroller/v1beta1"

	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Instance returns the Minio component to deploy
func InstanceV1Beta(mattermost *mattermostv1beta1.Mattermost) *minioOperator.MinIOInstance {
	minioName := fmt.Sprintf("%s-minio", mattermost.Name)

	return &minioOperator.MinIOInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      minioName,
			Namespace: mattermost.Namespace,
			Labels:    mattermostv1alpha1.ClusterInstallationResourceLabels(mattermost.Name),
			OwnerReferences: mattermostApp.MattermostOwnerReference(mattermost),
		},
		Spec: minioOperator.MinIOInstanceSpec{
			Replicas:    *mattermost.Spec.Filestore.OperatorManaged.Replicas,
			Mountpath:   "/export",
			CredsSecret: &corev1.LocalObjectReference{Name: minioName},
			VolumeClaimTemplate: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: minioName,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						"ReadWriteOnce",
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse(mattermost.Spec.Filestore.OperatorManaged.StorageSize),
						},
					},
				},
			},
		},
	}
}

// Secret returns the secret name created to use togehter with Minio deployment
func SecretV1Beta(mattermost *mattermostv1beta1.Mattermost) *corev1.Secret {
	secretName := DefaultMinioSecretName(mattermost.Name)
	data := make(map[string][]byte)
	data["accesskey"] = utils.New16ID()
	data["secretkey"] = utils.New28ID()

	return mattermostApp.GenerateSecretV1Beta(
		mattermost,
		secretName,
		mattermostv1alpha1.ClusterInstallationResourceLabels(mattermost.Name),
		data,
	)
}

//// DefaultMinioSecretName returns the default minio secret name based on
//// the provided installation name.
//func DefaultMinioSecretName(installationName string) string {
//	return fmt.Sprintf("%s-minio", installationName)
//}
