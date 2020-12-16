package mattermost

import (
	"context"

	"github.com/go-logr/logr"
	mattermostv1beta1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1"
	mattermostApp "github.com/mattermost/mattermost-operator/pkg/mattermost"
	"github.com/pkg/errors"
	mysqlOperator "github.com/presslabs/mysql-operator/pkg/apis/mysql/v1alpha1"
	"k8s.io/apimachinery/pkg/types"

	mattermostmysql "github.com/mattermost/mattermost-operator/pkg/components/mysql"
)

func (r *MattermostReconciler) checkOperatorManagedMySQL(mattermost *mattermostv1beta1.Mattermost, reqLogger logr.Logger) (mattermostApp.DatabaseConfig, error) {
	reqLogger = reqLogger.WithValues("Reconcile", "mysql")

	err := r.checkMySQLCluster(mattermost, reqLogger)
	if err != nil {
		return nil, errors.Wrap(err, "error while checking MySQL cluster")
	}

	dbSecretName := mattermostmysql.DefaultDatabaseSecretName(mattermost.Name)

	dbSecret, err := r.ResCreator.GetOrCreateMySQLSecrets(mattermost, dbSecretName, reqLogger)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get or create MySQL database secret")
	}

	return mattermostApp.NewMySQLDBConfig(*dbSecret)
}

func (r *MattermostReconciler) checkMySQLCluster(mattermost *mattermostv1beta1.Mattermost, reqLogger logr.Logger) error {
	desired := mattermostmysql.ClusterV1Beta(mattermost)

	err := r.ResCreator.CreateMySQLClusterIfNotExists(mattermost, desired, reqLogger)
	if err != nil {
		return err
	}

	current := &mysqlOperator.MysqlCluster{}
	if err := r.Client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current); err != nil {
		return err
	}

	return r.ResCreator.Update(current, desired, reqLogger)
}
