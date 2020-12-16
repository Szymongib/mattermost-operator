package mattermost

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	mattermostv1beta1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1"
	"github.com/mattermost/mattermost-operator/pkg/mattermost/healthcheck"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// checkMattermostHealth checks the health and correctness of the k8s
// objects that make up a Mattermost installation.
//
// NOTE: this is a vital health check. Every reconciliation loop should run this
// check at the very end to ensure that everything in the installation is as it
// should be. Over time, more types of checks should be added here as needed.
func (r *MattermostReconciler) checkMattermostHealth(mattermost *mattermostv1beta1.Mattermost, logger logr.Logger) (mattermostv1beta1.MattermostStatus, error) {
	status := mattermostv1beta1.MattermostStatus{
		State:           mattermostv1beta1.Reconciling,
		Replicas:        0,
		UpdatedReplicas: 0,
	}

	labels := mattermost.MattermostLabels(mattermost.Name)
	listOptions := []client.ListOption{
		client.InNamespace(mattermost.Namespace),
		client.MatchingLabels(labels),
	}

	healthChecker := healthcheck.NewHealthChecker(r.NonCachedAPIReader, listOptions, logger)

	err := healthChecker.AssertDeploymentRolloutStarted(mattermost.Name, mattermost.Namespace)
	if err != nil {
		return status, errors.Wrap(err, "rollout not yet started")
	}

	podsStatus, err := healthChecker.CheckPodsRolledOut(mattermost.GetImageName())
	if err != nil {
		return status, errors.Wrap(err, "failed to check pods status")
	}

	status.UpdatedReplicas = podsStatus.UpdatedReplicas
	status.Replicas = podsStatus.Replicas

	var replicas int32 = 0
	if mattermost.Spec.Replicas != nil {
		replicas = *mattermost.Spec.Replicas
	}

	if podsStatus.Replicas != replicas {
		return status, fmt.Errorf("found %d pods, but wanted %d", podsStatus.Replicas, replicas)
	}
	if podsStatus.UpdatedReplicas != replicas {
		return status, fmt.Errorf("found %d updated replicas, but wanted %d", podsStatus.UpdatedReplicas, replicas)
	}

	status.Image = mattermost.Spec.Image
	status.Version = mattermost.Spec.Version

	status.Endpoint = "not available"
	var endpoint string

	if mattermost.Spec.UseServiceLoadBalancer {
		endpoint, err = healthChecker.CheckServiceLoadBalancer()
		if err != nil {
			return status, errors.Wrap(err, "failed to check service load balancer")
		}
	} else {
		endpoint, err = healthChecker.CheckIngressLoadBalancer()
		if err != nil {
			return status, errors.Wrap(err, "failed to check ingress load balancer")
		}
	}

	if endpoint != "" {
		status.Endpoint = endpoint
	}

	// Everything checks out. The installation is stable.
	status.State = mattermostv1beta1.Stable

	return status, nil
}

func (r *MattermostReconciler) checkRolloutStarted(name, namespace string, listOpts []client.ListOption) error {
	// To prevent race condition that new pods did not start rolling out and
	// old ones are still ready, we need to check if Deployment was picked up by controller.
	// We use non-cached client to make sure it won't return old Deployment where
	// the generation and observedGeneration still match.
	deployment := &appsv1.Deployment{}
	deploymentKey := types.NamespacedName{Name: name, Namespace: namespace}
	err := r.NonCachedAPIReader.Get(context.TODO(), deploymentKey, deployment)
	if err != nil {
		return errors.Wrap(err, "failed to get deployment")
	}
	if deployment.Generation != deployment.Status.ObservedGeneration {
		return fmt.Errorf("mattermost deployment not yet picked up by the Deployment controller")
	}

	// We check if new ReplicaSet was created and it was observed by the controller
	// to guarantee that new pods are created.
	replicaSets := &appsv1.ReplicaSetList{}
	err = r.Client.List(context.TODO(), replicaSets, listOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to list replicaSets")
	}

	replicaSetReady := false
	for _, rep := range replicaSets.Items {
		if getRevision(rep.Annotations) == getRevision(deployment.Annotations) {
			if rep.Status.ObservedGeneration > 0 {
				replicaSetReady = true
				break
			}
		}
	}
	if !replicaSetReady {
		return fmt.Errorf("replicaSet did not start rolling pods")
	}

	return nil
}

func getRevision(annotations map[string]string) string {
	if annotations == nil {
		return ""
	}
	return annotations["deployment.kubernetes.io/revision"]
}
