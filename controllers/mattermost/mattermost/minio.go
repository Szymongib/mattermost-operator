package mattermost

import (
	"context"
	"fmt"

	mattermostv1beta1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1"
	"github.com/pkg/errors"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	mattermostMinio "github.com/mattermost/mattermost-operator/pkg/components/minio"

	minioOperator "github.com/minio/minio-operator/pkg/apis/miniocontroller/v1beta1"
)

func (r *MattermostReconciler) checkMattermostMinioSecret(mattermost *mattermostv1beta1.Mattermost, reqLogger logr.Logger) (*corev1.Secret, error) {
	current := &corev1.Secret{}
	desired := mattermostMinio.SecretV1Beta(mattermost)
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return r.createMinioSecret(mattermost, desired, reqLogger)
		}
		reqLogger.Error(err, "failed to check if secret exists")
		return nil, err
	}

	// Validate secret required fields, if not exist recreate.
	if _, ok := current.Data["accesskey"]; !ok {
		reqLogger.Info("minio secret does not have an 'accesskey' value, overriding", "name", desired.Name)
		err = r.update(current, desired, reqLogger)
		if err != nil {
			return nil, errors.Wrap(err, "failed to update Minio secret")
		}
	}
	if _, ok := current.Data["secretkey"]; !ok {
		reqLogger.Info("minio secret does not have an 'secretkey' value, overriding", "name", desired.Name)
		err = r.update(current, desired, reqLogger)
		if err != nil {
			return nil, errors.Wrap(err, "failed to update Minio secret")
		}
	}

	// Preserve data fields
	desired.Data = current.Data
	err = r.update(current, desired, reqLogger)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update Minio secret")
	}
	return desired, nil
}

func (r *MattermostReconciler) createMinioSecret(mattermost *mattermostv1beta1.Mattermost, desired *corev1.Secret, reqLogger logr.Logger) (*corev1.Secret, error) {
	reqLogger.Info("creating minio secret", "name", desired.Name, "namespace", desired.Namespace)
	err := r.create(mattermost, desired, reqLogger)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Minio secret")
	}
	return desired, nil
}

func (r *MattermostReconciler) checkMinioInstance(mattermost *mattermostv1beta1.Mattermost, reqLogger logr.Logger) error {
	desired := mattermostMinio.InstanceV1Beta(mattermost)

	err := r.createMinioInstanceIfNotExists(mattermost, desired, reqLogger)
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
	return r.update(current, desired, reqLogger)
}

func (r *MattermostReconciler) createMinioInstanceIfNotExists(mattermost *mattermostv1beta1.Mattermost, instance *minioOperator.MinIOInstance, reqLogger logr.Logger) error {
	foundInstance := &minioOperator.MinIOInstance{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, foundInstance)
	if err != nil && kerrors.IsNotFound(err) {
		reqLogger.Info("Creating minio instance")
		return r.create(mattermost, instance, reqLogger)
	} else if err != nil {
		reqLogger.Error(err, "Unable to get minio instance")
		return err
	}

	return nil
}

func (r *MattermostReconciler) getMinioService(mattermost *mattermostv1beta1.Mattermost, reqLogger logr.Logger) (string, error) {
	minioServiceName := fmt.Sprintf("%s-minio-hl-svc", mattermost.Name)
	minioService := &corev1.Service{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: minioServiceName, Namespace: mattermost.Namespace}, minioService)
	if err != nil {
		return "", err
	}

	connectionString := fmt.Sprintf("%s.%s:%d", minioService.Name, mattermost.Namespace, minioService.Spec.Ports[0].Port)
	return connectionString, nil
}
