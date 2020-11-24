package mattermost

import (
	"context"
	mattermostv1beta1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1"

	"github.com/go-logr/logr"
	mysqlOperator "github.com/presslabs/mysql-operator/pkg/apis/mysql/v1alpha1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	mattermostmysql "github.com/mattermost/mattermost-operator/pkg/components/mysql"
)

func (r *MattermostReconciler) checkMySQLCluster(mattermost *mattermostv1beta1.Mattermost, reqLogger logr.Logger) error {
	reqLogger = reqLogger.WithValues("Reconcile", "mysql")
	desired := mattermostmysql.ClusterV1Beta(mattermost)

	err := r.createMySQLClusterIfNotExists(mattermost, desired, reqLogger)
	if err != nil {
		return err
	}

	current := &mysqlOperator.MysqlCluster{}
	if err := r.Client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current); err != nil {
		return err
	}

	return r.update(current, desired, reqLogger)
}

func (r *MattermostReconciler) createMySQLClusterIfNotExists(mattermost *mattermostv1beta1.Mattermost, cluster *mysqlOperator.MysqlCluster, reqLogger logr.Logger) error {
	foundCluster := &mysqlOperator.MysqlCluster{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}, foundCluster)
	if err != nil && k8sErrors.IsNotFound(err) {
		reqLogger.Info("Creating mysql cluster")
		return r.create(mattermost, cluster, reqLogger)
	} else if err != nil {
		reqLogger.Error(err, "Failed to check if mysql cluster exists")
		return err
	}

	return nil
}
