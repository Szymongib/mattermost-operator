package mattermost

import (
	"errors"

	"github.com/go-logr/logr"

	mattermostv1alpha1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1alpha1"
)

// TODO: implement postgres
func (r *MattermostReconciler) checkPostgres(mattermost *mattermostv1alpha1.ClusterInstallation, reqLogger logr.Logger) error {
	// reqLogger := reqLogger.WithValues("Reconcile", "postgres")

	return errors.New("database type 'postgres' not yet implemented")
}
