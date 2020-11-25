package mattermost

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	mattermostv1beta1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1"
	mattermostApp "github.com/mattermost/mattermost-operator/pkg/mattermost"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *MattermostReconciler) checkDatabase(mattermost *mattermostv1beta1.Mattermost, reqLogger logr.Logger) (mattermostApp.DatabaseConfig, error) {
	reqLogger = reqLogger.WithValues("Reconcile", "database")

	if mattermost.Spec.Database.IsExternal() {
		return r.readExternalDBSecret(mattermost)
	}

	return r.checkOperatorManagedDB(mattermost, reqLogger)
}

func  (r *MattermostReconciler) readExternalDBSecret(mattermost *mattermostv1beta1.Mattermost) (mattermostApp.DatabaseConfig, error) {
	secretName := types.NamespacedName{Name: mattermost.Spec.Database.External.Secret, Namespace: mattermost.Namespace}

	var secret corev1.Secret
	err := r.Client.Get(context.TODO(), secretName, &secret)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get external db Secret")
	}

	return mattermostApp.NewExternalDBInfo(mattermost, secret)
}

func (r *MattermostReconciler) checkOperatorManagedDB(mattermost  *mattermostv1beta1.Mattermost, reqLogger logr.Logger) (mattermostApp.DatabaseConfig, error) {
	if mattermost.Spec.Database.OperatorManaged == nil {
		return nil, fmt.Errorf("configuration for Operator managed database not provided")
	}

	switch mattermost.Spec.Database.OperatorManaged.Type {
	case "mysql":
		return r.checkOperatorManagedMySQL(mattermost, reqLogger)
	case "postgres":
		return nil, errors.New("database type 'postgres' not yet implemented")
	}

	return nil, fmt.Errorf("database of type '%s' is not supported", mattermost.Spec.Database.OperatorManaged.Type)
}
