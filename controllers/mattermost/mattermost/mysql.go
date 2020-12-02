package mattermost

import (
	"context"
	mattermostv1beta1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1"
	"github.com/mattermost/mattermost-operator/pkg/components/utils"
	mattermostApp "github.com/mattermost/mattermost-operator/pkg/mattermost"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	"github.com/go-logr/logr"
	mysqlOperator "github.com/presslabs/mysql-operator/pkg/apis/mysql/v1alpha1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	mattermostmysql "github.com/mattermost/mattermost-operator/pkg/components/mysql"
)

func (r *MattermostReconciler) checkOperatorManagedMySQL(mattermost *mattermostv1beta1.Mattermost, reqLogger logr.Logger) (mattermostApp.DatabaseConfig, error) {
	reqLogger = reqLogger.WithValues("Reconcile", "mysql")

	err := r.checkMySQLCluster(mattermost, reqLogger)
	if err != nil {
		return nil, errors.Wrap(err, "error while checking MySQL cluster")
	}

	return r.getOrCreateMySQLSecrets(mattermost, reqLogger)
}

func (r *MattermostReconciler) checkMySQLCluster(mattermost *mattermostv1beta1.Mattermost, reqLogger logr.Logger) error {
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

func (r *MattermostReconciler) getOrCreateMySQLSecrets(mattermost *mattermostv1beta1.Mattermost, reqLogger logr.Logger) (mattermostApp.DatabaseConfig, error) {
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

func (r *MattermostReconciler) createMySQLSecret(mattermost *mattermostv1beta1.Mattermost, secretName string, reqLogger logr.Logger) (mattermostApp.DatabaseConfig, error) {
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
