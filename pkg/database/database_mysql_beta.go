package database
//
//import (
//	"fmt"
//	"github.com/go-logr/logr"
//	mattermostv1beta1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1"
//	corev1 "k8s.io/api/core/v1"
//)
//
//type MySQLDB struct {
//	secretName string
//	rootPassword string
//	userName     string
//	userPassword string
//	databaseName string
//}
//
//func NewMySQLDB(secret corev1.Secret) (*MySQLDB, error) {
//	rootPassword := string(secret.Data["ROOT_PASSWORD"])
//	if rootPassword == "" {
//		return nil, fmt.Errorf("database root password shouldn't be empty")
//	}
//	userName := string(secret.Data["USER"])
//	if userName == "" {
//		return nil, fmt.Errorf("database username shouldn't be empty")
//	}
//	userPassword := string(secret.Data["PASSWORD"])
//	if userPassword == "" {
//		return nil, fmt.Errorf("database password shouldn't be empty")
//	}
//	databaseName := string(secret.Data["DATABASE"])
//	if databaseName == "" {
//		return nil, fmt.Errorf("database name shouldn't be empty")
//	}
//
//	return &MySQLDB{
//		rootPassword:     rootPassword,
//		userName: userName,
//		userPassword:     userPassword,
//		databaseName:     databaseName,
//	}, nil
//
//}
//
//func (m *MySQLDB) EnvVars(mattermost *mattermostv1beta1.Mattermost) []corev1.EnvVar {
//	panic("implement me")
//}
//
//func (m *MySQLDB) InitContainers() []corev1.Container {
//	panic("implement me")
//}
//
//func (m *MySQLDB) Check(mattermost *mattermostv1beta1.Mattermost, reqLogger logr.Logger) error {
//	panic("implement me")
//}
//
