package mattermost

import (
	"fmt"
	mattermostv1beta1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1"
	"strconv"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"

	mattermostv1alpha1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1alpha1"
	"github.com/mattermost/mattermost-operator/pkg/database"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
)

//const (
//	SetupJobName                = "mattermost-db-setup"
//	WaitForDBSetupContainerName = "init-wait-for-db-setup"
//)


type DatabaseConfig interface {
	EnvVars(mattermost *mattermostv1beta1.Mattermost) []corev1.EnvVar
	InitContainers(mattermost *mattermostv1beta1.Mattermost) []corev1.Container
}

type FileStoreConfig interface {
	InitContainers(mattermost *mattermostv1beta1.Mattermost) []corev1.Container
}

// GenerateService returns the service for the Mattermost app.
func GenerateServiceV1Beta(mattermost *mattermostv1beta1.Mattermost, serviceName, selectorName string) *corev1.Service {
	baseAnnotations := map[string]string{
		"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true",
	}

	if mattermost.Spec.UseServiceLoadBalancer {
		// Create a LoadBalancer service with additional annotations provided in
		// the Mattermost Spec. The LoadBalancer is directly accessible from
		// outside the cluster thus exposes ports 80 and 443.
		service := newServiceV1Beta(mattermost, serviceName, selectorName,
			mergeStringMaps(baseAnnotations, mattermost.Spec.ServiceAnnotations),
		)
		service.Spec.Ports = []corev1.ServicePort{
			{
				Name:       "http",
				Port:       80,
				TargetPort: intstr.FromString("app"),
			},
			{
				Name:       "https",
				Port:       443,
				TargetPort: intstr.FromString("app"),
			},
		}
		service.Spec.Type = corev1.ServiceTypeLoadBalancer

		return service
	}

	// Create a headless service which is not directly accessible from outside
	// the cluster and thus exposes a custom port.
	service := newServiceV1Beta(mattermost, serviceName, selectorName, baseAnnotations)
	service.Spec.Ports = []corev1.ServicePort{
		{
			Port:       8065,
			Name:       "app",
			TargetPort: intstr.FromString("app"),
		},
		{
			Port:       8067,
			Name:       "metrics",
			TargetPort: intstr.FromString("metrics"),
		},
	}
	service.Spec.ClusterIP = corev1.ClusterIPNone

	return service
}

// GenerateIngress returns the ingress for the Mattermost app.
func GenerateIngressV1Beta(mattermost *mattermostv1beta1.Mattermost, name, ingressName string, ingressAnnotations map[string]string) *v1beta1.Ingress {
	ingress := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: mattermost.Namespace,
			Labels:    mattermost.MattermostLabels(name),
			OwnerReferences: MattermostOwnerReference(mattermost),
			Annotations: ingressAnnotations,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: ingressName,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: v1beta1.IngressBackend{
										ServiceName: name,
										ServicePort: intstr.FromInt(8065),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if mattermost.Spec.UseIngressTLS {
		ingress.Spec.TLS = []v1beta1.IngressTLS{
			{
				Hosts:      []string{ingressName},
				SecretName: strings.ReplaceAll(ingressName, ".", "-") + "-tls-cert",
			},
		}
	}

	return ingress
}

// GenerateDeployment returns the deployment for Mattermost app.
func GenerateDeploymentV1Beta(mattermost *mattermostv1beta1.Mattermost, db DatabaseConfig, fileStore *FileStoreInfo, deploymentName, ingressName, serviceAccountName, containerImage string) *appsv1.Deployment {
	envVarDB := db.EnvVars(mattermost)
	initContainers := db.InitContainers(mattermost)

	initContainers = append(initContainers, fileStore.config.InitContainers(mattermost)...)

	minioAccessEnv := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: fileStore.secretName,
			},
			Key: "accesskey",
		},
	}

	minioSecretEnv := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: fileStore.secretName,
			},
			Key: "secretkey",
		},
	}

	// Generate FileStore config
	envVarFileStore := []corev1.EnvVar{
		{
			Name:  "MM_FILESETTINGS_DRIVERNAME",
			Value: "amazons3",
		},
		{
			Name:      "MM_FILESETTINGS_AMAZONS3ACCESSKEYID",
			ValueFrom: minioAccessEnv,
		},
		{
			Name:      "MM_FILESETTINGS_AMAZONS3SECRETACCESSKEY",
			ValueFrom: minioSecretEnv,
		},
		{
			Name:  "MM_FILESETTINGS_AMAZONS3BUCKET",
			Value: fileStore.bucketName,
		},
		{
			Name:  "MM_FILESETTINGS_AMAZONS3ENDPOINT",
			Value: fileStore.url,
		},
		{
			Name:  "MM_FILESETTINGS_AMAZONS3SSL",
			Value: "false",
		},
	}

	// Add init container to wait for DB setup job to complete
	initContainers = append(initContainers, corev1.Container{
		Name:            WaitForDBSetupContainerName,
		Image:           "bitnami/kubectl:1.17",
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command: []string{
			"sh", "-c",
			fmt.Sprintf("kubectl wait --for=condition=complete --timeout 5m job/%s", SetupJobName),
		},
	})

	// ES section vars
	envVarES := []corev1.EnvVar{}
	if mattermost.Spec.ElasticSearch.Host != "" {
		envVarES = []corev1.EnvVar{
			{
				Name:  "MM_ELASTICSEARCHSETTINGS_ENABLEINDEXING",
				Value: "true",
			},
			{
				Name:  "MM_ELASTICSEARCHSETTINGS_ENABLESEARCHING",
				Value: "true",
			},
			{
				Name:  "MM_ELASTICSEARCHSETTINGS_CONNECTIONURL",
				Value: mattermost.Spec.ElasticSearch.Host,
			},
			{
				Name:  "MM_ELASTICSEARCHSETTINGS_USERNAME",
				Value: mattermost.Spec.ElasticSearch.UserName,
			},
			{
				Name:  "MM_ELASTICSEARCHSETTINGS_PASSWORD",
				Value: mattermost.Spec.ElasticSearch.Password,
			},
		}
	}

	siteURL := fmt.Sprintf("https://%s", ingressName)
	envVarGeneral := []corev1.EnvVar{
		{
			Name:  "MM_SERVICESETTINGS_SITEURL",
			Value: siteURL,
		},
		{
			Name:  "MM_PLUGINSETTINGS_ENABLEUPLOADS",
			Value: "true",
		},
		{
			Name:  "MM_METRICSSETTINGS_ENABLE",
			Value: "true",
		},
		{
			Name:  "MM_METRICSSETTINGS_LISTENADDRESS",
			Value: ":8067",
		},
		{
			Name:  "MM_CLUSTERSETTINGS_ENABLE",
			Value: "true",
		},
		{
			Name:  "MM_CLUSTERSETTINGS_CLUSTERNAME",
			Value: "production",
		},
		{
			Name:  "MM_INSTALL_TYPE",
			Value: "kubernetes-operator",
		},
	}

	valueSize := strconv.Itoa(defaultMaxFileSize * sizeMB)
	if !mattermost.Spec.UseServiceLoadBalancer {
		if _, ok := mattermost.Spec.IngressAnnotations["nginx.ingress.kubernetes.io/proxy-body-size"]; ok {
			size := mattermost.Spec.IngressAnnotations["nginx.ingress.kubernetes.io/proxy-body-size"]
			if strings.HasSuffix(size, "M") {
				maxFileSize, _ := strconv.Atoi(strings.TrimSuffix(size, "M"))
				valueSize = strconv.Itoa(maxFileSize * sizeMB)
			} else if strings.HasSuffix(size, "m") {
				maxFileSize, _ := strconv.Atoi(strings.TrimSuffix(size, "m"))
				valueSize = strconv.Itoa(maxFileSize * sizeMB)
			} else if strings.HasSuffix(size, "G") {
				maxFileSize, _ := strconv.Atoi(strings.TrimSuffix(size, "G"))
				valueSize = strconv.Itoa(maxFileSize * sizeGB)
			} else if strings.HasSuffix(size, "g") {
				maxFileSize, _ := strconv.Atoi(strings.TrimSuffix(size, "g"))
				valueSize = strconv.Itoa(maxFileSize * sizeGB)
			}
		}
	}
	envVarGeneral = append(envVarGeneral, corev1.EnvVar{
		Name:  "MM_FILESETTINGS_MAXFILESIZE",
		Value: valueSize,
	})

	volumes := mattermost.Spec.Volumes
	volumeMounts := mattermost.Spec.VolumeMounts
	podAnnotations := map[string]string{}

	// Mattermost License
	if len(mattermost.Spec.LicenseSecret) != 0 {
		envVarGeneral = append(envVarGeneral, corev1.EnvVar{
			Name:  "MM_SERVICESETTINGS_LICENSEFILELOCATION",
			Value: "/mattermost-license/license",
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			MountPath: "/mattermost-license",
			Name:      "mattermost-license",
			ReadOnly:  true,
		})

		volumes = append(volumes, corev1.Volume{
			Name: "mattermost-license",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: mattermost.Spec.LicenseSecret,
				},
			},
		})

		podAnnotations = map[string]string{
			"prometheus.io/scrape": "true",
			"prometheus.io/path":   "/metrics",
			"prometheus.io/port":   "8067",
		}
	}

	// EnvVars Section
	envVars := []corev1.EnvVar{}
	envVars = append(envVars, envVarDB...)
	envVars = append(envVars, envVarFileStore...)
	envVars = append(envVars, envVarES...)
	envVars = append(envVars, envVarGeneral...)

	// Merge our custom env vars in.
	envVars = mergeEnvVars(envVars, mattermost.Spec.MattermostEnv)

	revHistoryLimit := int32(defaultRevHistoryLimit)
	maxUnavailable := intstr.FromInt(defaultMaxUnavailable)
	maxSurge := intstr.FromInt(defaultMaxSurge)

	liveness, readiness := setProbes(mattermost.Spec.Advanced.LivenessProbe, mattermost.Spec.Advanced.ReadinessProbe)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: mattermost.Namespace,
			Labels:    mattermost.MattermostLabels(deploymentName),
			OwnerReferences: MattermostOwnerReference(mattermost),
		},
		Spec: appsv1.DeploymentSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &maxUnavailable,
					MaxSurge:       &maxSurge,
				},
			},
			RevisionHistoryLimit: &revHistoryLimit,
			Replicas:             mattermost.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: mattermostv1beta1.MattermostSelectorLabels(deploymentName),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      mattermost.MattermostLabels(deploymentName),
					Annotations: podAnnotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccountName,
					InitContainers:     initContainers,
					Containers: []corev1.Container{
						{
							Name:                     mattermostv1alpha1.MattermostAppContainerName,
							Image:                    containerImage,
							ImagePullPolicy:          corev1.PullIfNotPresent,
							TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
							Command:                  []string{"mattermost"},
							Env:                      envVars,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8065,
									Name:          "app",
								},
								{
									ContainerPort: 8067,
									Name:          "metrics",
								},
							},
							ReadinessProbe: readiness,
							LivenessProbe:  liveness,
							VolumeMounts:   volumeMounts,
							Resources:      mattermost.Spec.Advanced.Resources,
						},
					},
					Volumes:      volumes,
					Affinity:     mattermost.Spec.Advanced.Affinity,
					NodeSelector: mattermost.Spec.Advanced.NodeSelector,
				},
			},
		},
	}
}

// GenerateSecret returns the secret for Mattermost
func GenerateSecretV1Beta(mattermost *mattermostv1beta1.Mattermost, secretName string, labels map[string]string, values map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    labels,
			Name:      secretName,
			Namespace: mattermost.Namespace,
			OwnerReferences: MattermostOwnerReference(mattermost),
		},
		Data: values,
	}
}

// GenerateServiceAccount returns the Service Account for Mattermost
func GenerateServiceAccountV1Beta(mattermost *mattermostv1beta1.Mattermost, saName string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:            saName,
			Namespace:       mattermost.Namespace,
			OwnerReferences: MattermostOwnerReference(mattermost),
		},
	}
}

// GenerateRole returns the Role for Mattermost
func GenerateRoleV1Beta(mattermost *mattermostv1beta1.Mattermost, roleName string) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:            roleName,
			Namespace:       mattermost.Namespace,
			OwnerReferences: MattermostOwnerReference(mattermost),
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:         []string{"get", "list", "watch"},
				APIGroups:     []string{"batch"},
				Resources:     []string{"jobs"},
				ResourceNames: []string{SetupJobName},
			},
		},
	}
}

// GenerateRoleBinding returns the RoleBinding for Mattermost
func GenerateRoleBindingV1Beta(mattermost *mattermostv1beta1.Mattermost, roleName, saName string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            roleName,
			Namespace:       mattermost.Namespace,
			OwnerReferences: MattermostOwnerReference(mattermost),
		},
		Subjects: []rbacv1.Subject{
			{Kind: "ServiceAccount", Name: saName, Namespace: mattermost.Namespace},
		},
		RoleRef: rbacv1.RoleRef{Kind: "Role", Name: roleName},
	}
}

func MattermostOwnerReference(mattermost *mattermostv1beta1.Mattermost) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(mattermost, schema.GroupVersionKind{
			Group:   mattermostv1beta1.GroupVersion.Group,
			Version: mattermostv1beta1.GroupVersion.Version,
			Kind:    "Mattermost",
		}),
	}
}

// newService returns semi-finished service with common parts filled.
// Returned service is expected to be completed by the caller.
func newServiceV1Beta(mattermost *mattermostv1beta1.Mattermost, serviceName, selectorName string, annotations map[string]string) *corev1.Service {
	return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Labels:          mattermost.MattermostLabels(serviceName),
				Name:            serviceName,
				Namespace:       mattermost.Namespace,
				OwnerReferences: MattermostOwnerReference(mattermost),
				Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
		Selector: mattermostv1beta1.MattermostSelectorLabels(selectorName),
	},
	}
}

// GetDBCheckInitContainer tries to prepare init container that checks database readiness.
// Returns nil if database type is unknown.
func GetDBCheckInitContainerV1Beta(secretName, dbType string) *corev1.Container {
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
