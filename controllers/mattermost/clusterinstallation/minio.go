package clusterinstallation

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	mattermostv1alpha1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1alpha1"
	mattermostMinio "github.com/mattermost/mattermost-operator/pkg/components/minio"

	minioOperator "github.com/minio/minio-operator/pkg/apis/miniocontroller/v1beta1"
)

func (r *ClusterInstallationReconciler) checkMinio(mattermost *mattermostv1alpha1.ClusterInstallation, reqLogger logr.Logger) error {
	reqLogger = reqLogger.WithValues("Reconcile", "minio")

	err := r.checkMinioSecret(mattermost, reqLogger)
	if err != nil {
		return err
	}

	if mattermost.Spec.Minio.IsExternal() {
		return nil
	}

	return r.checkMinioInstance(mattermost, reqLogger)
}

func (r *ClusterInstallationReconciler) checkCustomMinioSecret(mattermost *mattermostv1alpha1.ClusterInstallation, reqLogger logr.Logger) error {
	secret := &corev1.Secret{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: mattermost.Spec.Minio.Secret, Namespace: mattermost.Namespace}, secret)
	if err != nil {
		reqLogger.Error(err, "failed to check if custom minio secret exists")
		return err
	}
	// Validate custom secret required fields
	if _, ok := secret.Data["accesskey"]; !ok {
		return fmt.Errorf("custom Minio Secret %s does not have an 'accesskey' value", mattermost.Spec.Minio.Secret)
	}
	if _, ok := secret.Data["secretkey"]; !ok {
		return fmt.Errorf("custom Minio Secret %s does not have an 'secretkey' value", mattermost.Spec.Minio.Secret)
	}
	return nil
}

func (r *ClusterInstallationReconciler) checkMattermostMinioSecret(mattermost *mattermostv1alpha1.ClusterInstallation, reqLogger logr.Logger) error {
	desiredSecret := mattermostMinio.Secret(mattermost)
	return r.ResCreator.CreateOrUpdateMinioSecret(mattermost, desiredSecret, reqLogger)
}

func (r *ClusterInstallationReconciler) checkMinioSecret(mattermost *mattermostv1alpha1.ClusterInstallation, reqLogger logr.Logger) error {
	if mattermost.Spec.Minio.Secret != "" {
		return r.checkCustomMinioSecret(mattermost, reqLogger)
	}
	return r.checkMattermostMinioSecret(mattermost, reqLogger)
}

func (r *ClusterInstallationReconciler) checkMinioInstance(mattermost *mattermostv1alpha1.ClusterInstallation, reqLogger logr.Logger) error {
	desired := mattermostMinio.Instance(mattermost)

	err := r.ResCreator.CreateMinioInstanceIfNotExists(mattermost, desired, reqLogger)
	if err != nil {
		return err
	}

	current := &minioOperator.MinIOInstance{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current)
	if err != nil {
		return err
	}

	// Note:
	// For some reason, our current minio operator seems to remove labels on
	// the instance resource when we add them. For that reason, trying to
	// ensure the labels are correct doesn't work.
	return r.ResCreator.Update(current, desired, reqLogger)
}

func (r *ClusterInstallationReconciler) getMinioService(mattermost *mattermostv1alpha1.ClusterInstallation, reqLogger logr.Logger) (string, error) {
	minioServiceName := fmt.Sprintf("%s-minio-hl-svc", mattermost.Name)
	minioService := &corev1.Service{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: minioServiceName, Namespace: mattermost.Namespace}, minioService)
	if err != nil {
		return "", err
	}

	connectionString := fmt.Sprintf("%s.%s:%d", minioService.Name, mattermost.Namespace, minioService.Spec.Ports[0].Port)
	return connectionString, nil
}
