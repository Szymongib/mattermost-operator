package mattermost

import (
	"context"
	"fmt"

	mattermostv1beta1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1"
	"github.com/pkg/errors"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	mattermostMinio "github.com/mattermost/mattermost-operator/pkg/components/minio"

	minioOperator "github.com/minio/minio-operator/pkg/apis/miniocontroller/v1beta1"
)

func (r *MattermostReconciler) checkMattermostMinioSecret(mattermost *mattermostv1beta1.Mattermost, logger logr.Logger) (*corev1.Secret, error) {
	desired := mattermostMinio.SecretV1Beta(mattermost)
	err := r.ResCreator.CreateOrUpdateMinioSecret(mattermost, desired, logger)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create or update Minio Secret")
	}
	return desired, nil
}

func (r *MattermostReconciler) checkMinioInstance(mattermost *mattermostv1beta1.Mattermost, reqLogger logr.Logger) error {
	desired := mattermostMinio.InstanceV1Beta(mattermost)

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

func (r *MattermostReconciler) getMinioService(mattermost *mattermostv1beta1.Mattermost) (string, error) {
	minioServiceName := fmt.Sprintf("%s-minio-hl-svc", mattermost.Name)
	minioService := &corev1.Service{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: minioServiceName, Namespace: mattermost.Namespace}, minioService)
	if err != nil {
		return "", err
	}

	connectionString := fmt.Sprintf("%s.%s:%d", minioService.Name, mattermost.Namespace, minioService.Spec.Ports[0].Port)
	return connectionString, nil
}
