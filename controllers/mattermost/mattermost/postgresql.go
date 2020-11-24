package mattermost

import (
	"errors"
	mattermostv1beta1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1"

	"github.com/go-logr/logr"
)

// TODO: implement postgres
func (r *MattermostReconciler) checkPostgres(mattermost *mattermostv1beta1.Mattermost, reqLogger logr.Logger) error {
	// reqLogger := reqLogger.WithValues("Reconcile", "postgres")

	return errors.New("database type 'postgres' not yet implemented")
}
