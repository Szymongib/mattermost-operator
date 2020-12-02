package mattermost

import (
	"fmt"
	mattermostv1beta1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1"
	"github.com/mattermost/mattermost-operator/pkg/database"
	corev1 "k8s.io/api/core/v1"
)

type ExternalDBConfig struct {
	secretName string
	dbType             string
	hasReaderEndpoints bool
	hasDBCheckURL      bool
}

func NewExternalDBInfo(mattermost *mattermostv1beta1.Mattermost, secret corev1.Secret) (*ExternalDBConfig, error) {
	if mattermost.Spec.Database.External == nil {
		return nil, fmt.Errorf("external database config not provided")
	}
	if mattermost.Spec.Database.External.Secret == "" {
		return nil, fmt.Errorf("external database Secret not provided")
	}

	connectionStr, ok := secret.Data["DB_CONNECTION_STRING"]
	if !ok {
		return nil, fmt.Errorf("external database Secret does not containt DB_CONNECTION_STRING key")
	}
	if len(connectionStr) == 0 {
		return nil, fmt.Errorf("external database connection string is empty")
	}

	externalDB := &ExternalDBConfig{
		secretName: mattermost.Spec.Database.External.Secret,
		dbType:    database.GetTypeFromConnectionString(string(connectionStr)),
	}

	if _, ok := secret.Data["MM_SQLSETTINGS_DATASOURCEREPLICAS"]; ok {
		externalDB.hasReaderEndpoints = true
	}
	if _, ok := secret.Data["DB_CONNECTION_CHECK_URL"]; ok {
		externalDB.hasDBCheckURL = true
	}

	return externalDB, nil
}

func (e *ExternalDBConfig) EnvVars(_ *mattermostv1beta1.Mattermost) []corev1.EnvVar {
	var dbEnvVars []corev1.EnvVar = []corev1.EnvVar{
		{
			Name: "MM_CONFIG",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: e.secretName,
					},
					Key: "DB_CONNECTION_STRING",
				},
			},
		},
	}

	if e.hasReaderEndpoints {
		dbEnvVars = append(dbEnvVars, corev1.EnvVar{
			Name: "MM_SQLSETTINGS_DATASOURCEREPLICAS",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: e.secretName,
					},
					Key: "MM_SQLSETTINGS_DATASOURCEREPLICAS",
				},
			},
		})
	}

	return dbEnvVars
}

func (e *ExternalDBConfig) InitContainers(_ *mattermostv1beta1.Mattermost) []corev1.Container {
	var initContainers []corev1.Container
	// TODO: move this func here
	if e.hasDBCheckURL {
		container := GetDBCheckInitContainerV1Beta(e.secretName, e.dbType)
		if container != nil {
			initContainers = append(initContainers, *container)
		}
	}

	return initContainers
}
