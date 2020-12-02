package mattermost

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	mattermostv1beta1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pkg/errors"
)

// handleCheckMattermostHealth checks the health and correctness of the k8s
// objects that make up a Mattermost installation.
//
// NOTE: this is a vital health check. Every reconciliation loop should run this
// check at the very end to ensure that everything in the installation is as it
// should be. Over time, more types of checks should be added here as needed.
func (r *MattermostReconciler) handleCheckMattermostHealth(mattermost *mattermostv1beta1.Mattermost) (mattermostv1beta1.MattermostStatus, error) {
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

	err := r.checkRolloutStarted(mattermost.Name, mattermost.Namespace, listOptions)
	if err != nil {
		return status, errors.Wrap(err, "rollout not yet started")
	}

	pods := &corev1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
	}

	// We use non-cached client to make sure we get the real state of pods instead of
	// potentially outdated cached data.
	err = r.NonCachedAPIReader.List(context.TODO(), pods, listOptions...)
	if err != nil {
		return status, errors.Wrap(err, "unable to get pod list")
	}

	status.Replicas = int32(len(pods.Items))

	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodRunning || pod.DeletionTimestamp != nil {
			return status, fmt.Errorf("mattermost pod %s is in state '%s'", pod.Name, pod.Status.Phase)
		}
		if len(pod.Spec.Containers) == 0 {
			return status, fmt.Errorf("mattermost pod %s has no containers", pod.Name)
		}
		if pod.Spec.Containers[0].Image != mattermost.GetImageName() {
			return status, fmt.Errorf("mattermost pod %s is running incorrect image", pod.Name)
		}

		podIsReady := false
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady {
				if condition.Status == corev1.ConditionTrue {
					podIsReady = true
					break
				} else {
					return status, fmt.Errorf("mattermost pod %s is not ready", pod.Name)
				}
			}
		}
		if !podIsReady {
			return status, fmt.Errorf("mattermost pod %s is not ready", pod.Name)
		}

		status.UpdatedReplicas++
	}

	var replicas int32 = 1
	if mattermost.Spec.Replicas != nil {
		replicas = *mattermost.Spec.Replicas
	}

	if int32(len(pods.Items)) != replicas {
		return status, fmt.Errorf("found %d pods, but wanted %d", len(pods.Items), replicas)
	}

	status.Image = mattermost.Spec.Image
	status.Version = mattermost.Spec.Version

	status.Endpoint = "not available"
	if mattermost.Spec.UseServiceLoadBalancer {
		svc := &corev1.ServiceList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
		}
		err := r.Client.List(context.TODO(), svc, listOptions...)
		if err != nil {
			return status, errors.Wrap(err, "unable to get service list")
		}
		if len(svc.Items) != 1 {
			return status, fmt.Errorf("should return just one service, but returned %d", len(svc.Items))
		}
		if svc.Items[0].Status.LoadBalancer.Ingress == nil {
			return status, errors.New("waiting for the Load Balancer to be active")
		}
		lbIngress := svc.Items[0].Status.LoadBalancer.Ingress[0]
		if lbIngress.Hostname != "" {
			status.Endpoint = lbIngress.Hostname
		} else if lbIngress.IP != "" {
			status.Endpoint = lbIngress.IP
		}
	} else {
		ingress := &v1beta1.IngressList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Ingress",
				APIVersion: "v1",
			},
		}
		err := r.Client.List(context.TODO(), ingress, listOptions...)
		if err != nil {
			return status, errors.Wrap(err, "unable to get ingress list")
		}
		if len(ingress.Items) != 1 {
			return status, fmt.Errorf("should return just one ingress, but returned %d", len(ingress.Items))
		}
		status.Endpoint = ingress.Items[0].Spec.Rules[0].Host
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

func (r *MattermostReconciler) checkSecret(secretName, keyName, namespace string) error {
	foundSecret := &corev1.Secret{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: namespace}, foundSecret)
	if err != nil {
		return errors.Wrap(err, "error getting secret")
	}

	for key := range foundSecret.Data {
		if keyName == key {
			return nil
		}
	}

	return fmt.Errorf("secret %s is missing data key: %s", secretName, keyName)
}

// setStateReconciling sets the Mattermost state to reconciling.
func (r *MattermostReconciler) setStateReconciling(mattermost *mattermostv1beta1.Mattermost, reqLogger logr.Logger) error {
	return r.setState(mattermost, mattermostv1beta1.Reconciling, reqLogger)
}

// setStateReconcilingAndLogError attempts to set the Mattermost state
// to reconciling. Any errors attempting this are logged, but not returned. This
// should only be used when the outcome of setting the state can be ignored.
func (r *MattermostReconciler) setStateReconcilingAndLogError(mattermost *mattermostv1beta1.Mattermost, reqLogger logr.Logger) {
	err := r.setStateReconciling(mattermost, reqLogger)
	if err != nil {
		reqLogger.Error(err, "Failed to set state to reconciling")
	}
}

// setState sets the provided Mattermost to the provided state if that
// is different from the current state.
func (r *MattermostReconciler) setState(mattermost *mattermostv1beta1.Mattermost, desired mattermostv1beta1.RunningState, reqLogger logr.Logger) error {
	if mattermost.Status.State == desired {
		return nil
	}

	status := mattermost.Status
	status.State = desired
	err := r.updateStatus(mattermost, status, reqLogger)
	if err != nil {
		return errors.Wrapf(err, "failed to set state to %s", desired)
	}

	return nil
}

func (r *MattermostReconciler) updateStatus(mattermost *mattermostv1beta1.Mattermost, status mattermostv1beta1.MattermostStatus, reqLogger logr.Logger) error {
	if reflect.DeepEqual(mattermost.Status, status) {
		return nil
	}

	if mattermost.Status.State != status.State {
		reqLogger.Info(fmt.Sprintf("Updating Mattermost state from '%s' to '%s'", mattermost.Status.State, status.State))
	}

	mattermost.Status = status
	err := r.Client.Status().Update(context.TODO(), mattermost)
	if err != nil {
		return errors.Wrap(err, "failed to update the Mattermost status")
	}

	return nil
}

func getRevision(annotations map[string]string) string {
	if annotations == nil {
		return ""
	}
	return annotations["deployment.kubernetes.io/revision"]
}
