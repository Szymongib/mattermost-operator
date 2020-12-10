package mattermost

import (
	"fmt"

	mattermostv1beta1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1"
	"github.com/mattermost/mattermost-operator/pkg/database"
	corev1 "k8s.io/api/core/v1"
)

type ExternalDBConfig struct {
	secretName         string
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
		dbType:     database.GetTypeFromConnectionString(string(connectionStr)),
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
	dbEnvVars := []corev1.EnvVar{
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
	if e.hasDBCheckURL {
		container := getDBCheckInitContainerV1Beta(e.secretName, e.dbType)
		if container != nil {
			initContainers = append(initContainers, *container)
		}
	}

	return initContainers
}

// getDBCheckInitContainer tries to prepare init container that checks database readiness.
// Returns nil if database type is unknown.
func getDBCheckInitContainerV1Beta(secretName, dbType string) *corev1.Container {
	envVars := []corev1.EnvVar{
		{
			Name: "DB_CONNECTION_CHECK_URL",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: "DB_CONNECTION_CHECK_URL",
				},
			},
		},
	}

	switch dbType {
	case database.MySQLDatabase:
		return &corev1.Container{
			Name:            "init-check-database",
			Image:           "appropriate/curl:latest",
			ImagePullPolicy: corev1.PullIfNotPresent,
			Env:             envVars,
			Command: []string{
				"sh", "-c",
				"until curl --max-time 5 $DB_CONNECTION_CHECK_URL; do echo waiting for database; sleep 5; done;",
			},
		}
	case database.PostgreSQLDatabase:
		return &corev1.Container{
			Name:            "init-check-database",
			Image:           "postgres:13",
			ImagePullPolicy: corev1.PullIfNotPresent,
			Env:             envVars,
			Command: []string{
				"sh", "-c",
				"until pg_isready --dbname=\"$DB_CONNECTION_CHECK_URL\"; do echo waiting for database; sleep 5; done;",
			},
		}
	default:
		return nil
	}
}
