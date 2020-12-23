// +build !ignore_autogenerated

// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

// Code generated by openapi-gen. DO NOT EDIT.

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1beta1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1.Mattermost":     schema_mattermost_operator_apis_mattermost_v1beta1_Mattermost(ref),
		"github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1.MattermostSpec": schema_mattermost_operator_apis_mattermost_v1beta1_MattermostSpec(ref),
	}
}

func schema_mattermost_operator_apis_mattermost_v1beta1_Mattermost(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "Mattermost is the Schema for the mattermosts API",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Default: map[string]interface{}{},
							Ref:     ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Default: map[string]interface{}{},
							Ref:     ref("github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1.MattermostSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Default: map[string]interface{}{},
							Ref:     ref("github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1.MattermostStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1.MattermostSpec", "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1.MattermostStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_mattermost_operator_apis_mattermost_v1beta1_MattermostSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "MattermostSpec defines the desired state of Mattermost",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"size": {
						SchemaProps: spec.SchemaProps{
							Description: "Size defines the size of the Mattermost. This is typically specified in number of users. This will override replica and resource requests/limits appropriately for the provided number of users. This is a write-only field - its value is erased after setting appropriate values of resources. Accepted values are: 100users, 1000users, 5000users, 10000users, and 250000users. If replicas and resource requests/limits are not specified, and Size is not provided the configuration for 5000users will be applied. Setting 'Replicas', 'Scheduling.Resources', 'FileStore.Replicas', 'FileStore.Resource', 'Database.Replicas', or 'Database.Resources' will override the values set by Size. Setting new Size will override previous values regardless if set by Size or manually.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"image": {
						SchemaProps: spec.SchemaProps{
							Description: "Image defines the Mattermost Docker image.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"version": {
						SchemaProps: spec.SchemaProps{
							Description: "Version defines the Mattermost Docker image version.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"replicas": {
						SchemaProps: spec.SchemaProps{
							Description: "Replicas defines the number of replicas to use for the Mattermost app servers.",
							Type:        []string{"integer"},
							Format:      "int32",
						},
					},
					"mattermostEnv": {
						SchemaProps: spec.SchemaProps{
							Description: "Optional environment variables to set in the Mattermost application pods.",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Default: map[string]interface{}{},
										Ref:     ref("k8s.io/api/core/v1.EnvVar"),
									},
								},
							},
						},
					},
					"licenseSecret": {
						SchemaProps: spec.SchemaProps{
							Description: "LicenseSecret is the name of the secret containing a Mattermost license.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"ingressName": {
						SchemaProps: spec.SchemaProps{
							Description: "IngressName defines the name to be used when creating the ingress rules",
							Default:     "",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"ingressAnnotations": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							AdditionalProperties: &spec.SchemaOrBool{
								Allows: true,
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Default: "",
										Type:    []string{"string"},
										Format:  "",
									},
								},
							},
						},
					},
					"useIngressTLS": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"useServiceLoadBalancer": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"serviceAnnotations": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							AdditionalProperties: &spec.SchemaOrBool{
								Allows: true,
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Default: "",
										Type:    []string{"string"},
										Format:  "",
									},
								},
							},
						},
					},
					"resourceLabels": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							AdditionalProperties: &spec.SchemaOrBool{
								Allows: true,
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Default: "",
										Type:    []string{"string"},
										Format:  "",
									},
								},
							},
						},
					},
					"volumes": {
						SchemaProps: spec.SchemaProps{
							Description: "Volumes allows for mounting volumes from various sources into the Mattermost application pods.",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Default: map[string]interface{}{},
										Ref:     ref("k8s.io/api/core/v1.Volume"),
									},
								},
							},
						},
					},
					"volumeMounts": {
						SchemaProps: spec.SchemaProps{
							Description: "Defines additional volumeMounts to add to Mattermost application pods.",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Default: map[string]interface{}{},
										Ref:     ref("k8s.io/api/core/v1.VolumeMount"),
									},
								},
							},
						},
					},
					"imagePullPolicy": {
						SchemaProps: spec.SchemaProps{
							Description: "Specify Mattermost deployment pull policy.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"database": {
						SchemaProps: spec.SchemaProps{
							Description: "External Services",
							Default:     map[string]interface{}{},
							Ref:         ref("github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1.Database"),
						},
					},
					"fileStore": {
						SchemaProps: spec.SchemaProps{
							Default: map[string]interface{}{},
							Ref:     ref("github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1.FileStore"),
						},
					},
					"elasticSearch": {
						SchemaProps: spec.SchemaProps{
							Default: map[string]interface{}{},
							Ref:     ref("github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1.ElasticSearch"),
						},
					},
					"scheduling": {
						SchemaProps: spec.SchemaProps{
							Description: "Scheduling defines the configuration related to scheduling of the Mattermost pods as well as resource constraints. These settings generally don't need to be changed.",
							Default:     map[string]interface{}{},
							Ref:         ref("github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1.Scheduling"),
						},
					},
					"probes": {
						SchemaProps: spec.SchemaProps{
							Description: "Probes defines configuration of liveness and readiness probe for Mattermost pods. These settings generally don't need to be changed.",
							Default:     map[string]interface{}{},
							Ref:         ref("github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1.Probes"),
						},
					},
				},
				Required: []string{"ingressName"},
			},
		},
		Dependencies: []string{
			"github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1.Database", "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1.ElasticSearch", "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1.FileStore", "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1.Probes", "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1.Scheduling", "k8s.io/api/core/v1.EnvVar", "k8s.io/api/core/v1.Volume", "k8s.io/api/core/v1.VolumeMount"},
	}
}
