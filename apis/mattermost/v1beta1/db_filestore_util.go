// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package v1beta1

import "github.com/mattermost/mattermost-operator/pkg/utils"

// FileStore utils

// SetDefaults sets the missing values in FileStore to the default ones.
func (fs *FileStore) SetDefaults() {
	if fs.IsExternal() {
		return
	}

	fs.ensureDefault()
	fs.OperatorManaged.SetDefaults()
}

// IsExternal returns true if the MinIO/S3 instance is external.
func (fs *FileStore) IsExternal() bool {
	return fs.External != nil && fs.External.URL != ""
}

func (fs *FileStore) ensureDefault() {
	if fs.OperatorManaged == nil {
		fs.OperatorManaged = &OperatorManagedMinio{}
	}
}

// SetDefaults sets the missing values in OperatorManagedMinio to the default ones.
func (omm *OperatorManagedMinio) SetDefaults() {
	if omm.StorageSize == "" {
		omm.StorageSize = DefaultFilestoreStorageSize
	}
}

func (fs *FileStore) SetDefaultReplicasAndResources() {
	if fs.IsExternal() {
		return
	}
	fs.ensureDefault()
	fs.OperatorManaged.SetDefaultReplicasAndResources()
}

func (omm *OperatorManagedMinio) SetDefaultReplicasAndResources() {
	if omm.Replicas == nil {
		omm.Replicas = &defaultSize.Minio.Replicas
	}
	if omm.Resources.Size() == 0 {
		omm.Resources = defaultSize.Minio.Resources
	}
}

func (fs *FileStore) OverrideReplicasAndResourcesFromSize(size MattermostSize) {
	if fs.IsExternal() {
		return
	}
	fs.ensureDefault()
	fs.OperatorManaged.OverrideReplicasAndResourcesFromSize(size)
}

func (omm *OperatorManagedMinio) OverrideReplicasAndResourcesFromSize(size MattermostSize) {
	omm.Replicas = utils.NewInt32(size.Minio.Replicas)
	omm.Resources = size.Minio.Resources
}

// Database utils

// SetDefaults sets the missing values in Database to the default ones.
func (db *Database) SetDefaults() {
	if db.IsExternal() {
		return
	}

	if db.OperatorManaged == nil {
		db.OperatorManaged = &OperatorManagedDatabase{}
	}

	db.OperatorManaged.SetDefaults()
}

// IsExternal returns true if the Database is set to external.
func (db *Database) IsExternal() bool {
	return db.External != nil && db.External.Secret != ""
}

func (db *Database) ensureDefault() {
	if db.OperatorManaged == nil {
		db.OperatorManaged = &OperatorManagedDatabase{}
	}
}

// SetDefaults sets the missing values in OperatorManagedDatabase to the default ones.
func (omd *OperatorManagedDatabase) SetDefaults() {
	if omd.Type == "" {
		omd.Type = DefaultMattermostDatabaseType
	}
	if omd.StorageSize == "" {
		omd.StorageSize = DefaultStorageSize
	}
}

func (db *Database) SetDefaultReplicasAndResources() {
	if db.IsExternal() {
		return
	}
	db.ensureDefault()
	db.OperatorManaged.SetDefaultReplicasAndResources()
}

func (omd *OperatorManagedDatabase) SetDefaultReplicasAndResources() {
	if omd.Replicas == nil {
		omd.Replicas = &defaultSize.Database.Replicas
	}
	if omd.Resources.Size() == 0 {
		omd.Resources = defaultSize.Database.Resources
	}
}

func (db *Database) OverrideReplicasAndResourcesFromSize(size MattermostSize) {
	if db.IsExternal() {
		return
	}
	db.ensureDefault()
	db.OperatorManaged.OverrideReplicasAndResourcesFromSize(size)
}

func (omd *OperatorManagedDatabase) OverrideReplicasAndResourcesFromSize(size MattermostSize) {
	omd.Replicas = utils.NewInt32(size.Database.Replicas)
	omd.Resources = size.Database.Resources
}

// MySQLLabels returns the labels for selecting the resources belonging to the
// given mysql cluster.
func MySQLLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/component":  "database",
		"app.kubernetes.io/instance":   "db",
		"app.kubernetes.io/managed-by": "mysql.presslabs.org",
		"app.kubernetes.io/name":       "mysql",
	}
}
