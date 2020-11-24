package mattermost

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	mattermostv1beta1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1"
	mattermostmysql "github.com/mattermost/mattermost-operator/pkg/components/mysql"
	"github.com/mattermost/mattermost-operator/pkg/components/utils"
	mattermostApp "github.com/mattermost/mattermost-operator/pkg/mattermost"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// TODO: MM resource builder?


func (r *MattermostReconciler) checkDatabaseSecret(mattermost *mattermostv1beta1.Mattermost, reqLogger logr.Logger) (mattermostApp.DatabaseInfo, error) {
	if mattermost.Spec.Database.IsExternal() {
		return r.readExternalDBSecret(mattermost)
	}

	if mattermost.Spec.Database.OperatorManaged == nil {
		return nil, fmt.Errorf("operator managed database config not present for non external db")
	}

	switch mattermost.Spec.Database.OperatorManaged.Type {
	case "mysql":
		return r.getOrCreateMySQLSecrets(mattermost, reqLogger)
	case "postgres":

	}

	return nil, fmt.Errorf("unsupported database")
}

func  (r *MattermostReconciler) readExternalDBSecret(mattermost *mattermostv1beta1.Mattermost) (mattermostApp.DatabaseInfo, error) {
	secretName := types.NamespacedName{Name: mattermost.Spec.Database.External.Secret, Namespace: mattermost.Namespace}

	var secret corev1.Secret
	err := r.Client.Get(context.TODO(), secretName, &secret)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get external db Secret")
	}

	return mattermostApp.NewExternalDB(mattermost, secret)
}


func (r *MattermostReconciler) getOrCreateMySQLSecrets(mattermost  *mattermostv1beta1.Mattermost, reqLogger logr.Logger) (mattermostApp.DatabaseInfo, error) {
	var err error
	dbSecret := &corev1.Secret{}
	dbSecretName := mattermostmysql.DefaultDatabaseSecretName(mattermost.Name)

	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: dbSecretName, Namespace: mattermost.Namespace}, dbSecret)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return r.createMySQLSecret(mattermost, dbSecretName, reqLogger)
		}

		reqLogger.Error(err, "failed to check if mysql secret exists")
		return nil, err
	}

	return mattermostApp.NewMySQLDB(*dbSecret)
}

func (r *MattermostReconciler) createMySQLSecret(mattermost *mattermostv1beta1.Mattermost, secretName string, reqLogger logr.Logger) (mattermostApp.DatabaseInfo, error) {
	reqLogger.Info("Creating new mysql secret")

	dbSecret := &corev1.Secret{}

	dbSecret.SetName(secretName)
	dbSecret.SetNamespace(mattermost.Namespace)
	userName := "mmuser"
	dbName := "mattermost"
	rootPassword := string(utils.New16ID())
	userPassword := string(utils.New16ID())

	dbSecret.Data = map[string][]byte{
		"ROOT_PASSWORD": []byte(rootPassword),
		"USER":          []byte(userName),
		"PASSWORD":      []byte(userPassword),
		"DATABASE":      []byte(dbName),
	}

	err := r.create(mattermost, dbSecret, reqLogger)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create mysql secret")
	}

	return mattermostApp.NewMySQLDB(*dbSecret)
}
