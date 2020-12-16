package mattermost

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	mattermostv1beta1 "github.com/mattermost/mattermost-operator/apis/mattermost/v1beta1"
	mattermostApp "github.com/mattermost/mattermost-operator/pkg/mattermost"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"
)

const updateJobName = "mattermost-update-check"

func (r *MattermostReconciler) checkMattermost(
	mattermost *mattermostv1beta1.Mattermost,
	dbInfo mattermostApp.DatabaseConfig,
	fileStoreInfo *mattermostApp.FileStoreInfo,
	reqLogger logr.Logger) error {
	reqLogger = reqLogger.WithValues("Reconcile", "mattermost")

	err := r.checkLicence(mattermost)
	if err != nil {
		return errors.Wrap(err, "failed to check mattermost license secret.")
	}

	err = r.checkMattermostService(mattermost, mattermost.Name, mattermost.GetProductionDeploymentName(), reqLogger)
	if err != nil {
		return err
	}

	err = r.checkMattermostRBAC(mattermost, mattermost.Name, mattermost.Name, reqLogger)
	if err != nil {
		return err
	}

	if !mattermost.Spec.UseServiceLoadBalancer {
		err = r.checkMattermostIngress(mattermost, mattermost.Spec.IngressName, reqLogger)
		if err != nil {
			return err
		}
	}

	err = r.checkMattermostDeployment(mattermost, dbInfo, fileStoreInfo, reqLogger)
	if err != nil {
		return err
	}

	return nil
}

func (r *MattermostReconciler) checkLicence(mattermost *mattermostv1beta1.Mattermost) error {
	if mattermost.Spec.LicenseSecret == "" {
		return nil
	}
	return r.assertSecretContains(mattermost.Spec.LicenseSecret, "license", mattermost.Namespace)
}

func (r *MattermostReconciler) checkMattermostService(mattermost *mattermostv1beta1.Mattermost, resourceName, selectorName string, reqLogger logr.Logger) error {
	desired := mattermostApp.GenerateServiceV1Beta(mattermost, resourceName, selectorName)

	err := r.ResCreator.CreateServiceIfNotExists(mattermost, desired, reqLogger)
	if err != nil {
		return err
	}

	current := &corev1.Service{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current)
	if err != nil {
		return err
	}

	return r.ResCreator.Update(current, desired, reqLogger)
}

func (r *MattermostReconciler) checkMattermostRBAC(mattermost *mattermostv1beta1.Mattermost, roleName, saName string, reqLogger logr.Logger) error {
	err := r.checkMattermostSA(mattermost, saName, reqLogger)
	if err != nil {
		return errors.Wrap(err, "failed to check mattermost ServiceAccount")
	}
	err = r.checkMattermostRole(mattermost, roleName, reqLogger)
	if err != nil {
		return errors.Wrap(err, "failed to check mattermost Role")
	}
	err = r.checkMattermostRoleBinding(mattermost, roleName, saName, reqLogger)
	if err != nil {
		return errors.Wrap(err, "failed to check mattermost RoleBinding")
	}

	return nil
}

func (r *MattermostReconciler) checkMattermostSA(mattermost *mattermostv1beta1.Mattermost, saName string, reqLogger logr.Logger) error {
	desired := mattermostApp.GenerateServiceAccountV1Beta(mattermost, saName)
	err := r.ResCreator.CreateServiceAccountIfNotExists(mattermost, desired, reqLogger)
	if err != nil {
		return err
	}

	current := &corev1.ServiceAccount{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current)
	if err != nil {
		return err
	}

	return r.ResCreator.Update(current, desired, reqLogger)
}

func (r *MattermostReconciler) checkMattermostRole(mattermost *mattermostv1beta1.Mattermost, roleName string, reqLogger logr.Logger) error {
	desired := mattermostApp.GenerateRoleV1Beta(mattermost, roleName)
	err := r.ResCreator.CreateRoleIfNotExists(mattermost, desired, reqLogger)
	if err != nil {
		return err
	}

	current := &rbacv1.Role{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current)
	if err != nil {
		return err
	}

	return r.ResCreator.Update(current, desired, reqLogger)
}

func (r *MattermostReconciler) checkMattermostRoleBinding(mattermost *mattermostv1beta1.Mattermost, roleName, saName string, reqLogger logr.Logger) error {
	desired := mattermostApp.GenerateRoleBindingV1Beta(mattermost, roleName, saName)
	err := r.ResCreator.CreateRoleBindingIfNotExists(mattermost, desired, reqLogger)
	if err != nil {
		return err
	}

	current := &rbacv1.RoleBinding{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current)
	if err != nil {
		return err
	}

	return r.ResCreator.Update(current, desired, reqLogger)
}

func (r *MattermostReconciler) checkMattermostIngress(mattermost *mattermostv1beta1.Mattermost, ingressHost string, reqLogger logr.Logger) error {
	ingressAnnotations := map[string]string{
		"kubernetes.io/ingress.class":                 "nginx",
		"nginx.ingress.kubernetes.io/proxy-body-size": "1000M",
	}
	for k, v := range mattermost.Spec.IngressAnnotations {
		ingressAnnotations[k] = v
	}

	desired := mattermostApp.GenerateIngressV1Beta(mattermost, mattermost.Name, ingressHost, ingressAnnotations)

	err := r.ResCreator.CreateIngressIfNotExists(mattermost, desired, reqLogger)
	if err != nil {
		return err
	}

	current := &v1beta1.Ingress{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current)
	if err != nil {
		return err
	}

	return r.ResCreator.Update(current, desired, reqLogger)
}

func (r *MattermostReconciler) checkMattermostDeployment(
	mattermost *mattermostv1beta1.Mattermost,
	dbInfo mattermostApp.DatabaseConfig,
	fileStoreInfo *mattermostApp.FileStoreInfo, reqLogger logr.Logger) error {

	desired := mattermostApp.GenerateDeploymentV1Beta(
		mattermost,
		dbInfo,
		fileStoreInfo,
		mattermost.Name,
		mattermost.Spec.IngressName,
		mattermost.Name,
		mattermost.GetImageName(),
	)

	// TODO: DB setup job is temporarily disabled as `mattermost version` command
	// does not account for the custom configuration
	//err = r.checkMattermostDBSetupJob(mattermost, desired, reqLogger)
	//if err != nil {
	//	return errors.Wrap(err, "failed to check mattermost DB setup job")
	//}

	err := r.ResCreator.CreateDeploymentIfNotExists(mattermost, desired, reqLogger)
	if err != nil {
		return errors.Wrap(err, "failed to create mattermost deployment")
	}

	current := &appsv1.Deployment{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current)
	if err != nil {
		return errors.Wrap(err, "failed to get mattermost deployment")
	}

	err = r.updateMattermostDeployment(mattermost, current, desired, reqLogger)
	if err != nil {
		return errors.Wrap(err, "failed to update mattermost deployment")
	}

	return nil
}

func (r *MattermostReconciler) checkMattermostDBSetupJob(mattermost *mattermostv1beta1.Mattermost, deployment *appsv1.Deployment, reqLogger logr.Logger) error {
	desiredJob := prepareJobTemplate(mattermostApp.SetupJobName, mattermost.Namespace, deployment)
	desiredJob.OwnerReferences = mattermostApp.MattermostOwnerReference(mattermost)

	currentJob := &batchv1.Job{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: desiredJob.Name, Namespace: desiredJob.Namespace}, currentJob)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			reqLogger.Info("Creating DB setup job", "name", desiredJob.Name)
			return r.ResCreator.Create(mattermost, desiredJob, reqLogger)
		}
		return errors.Wrap(err, "failed to get current db setup job")
	}
	// For now, there is no need to perform job update, so just return.
	return nil
}

func (r *MattermostReconciler) launchUpdateJob(
	jobNamespace string,
	deployment *appsv1.Deployment,
) error {
	job := prepareJobTemplate(updateJobName, jobNamespace, deployment)

	err := r.Client.Create(context.TODO(), job)
	if err != nil && !k8sErrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

// restartUpdateJob removes existing update job if it exists and creates new one.
func (r *MattermostReconciler) restartUpdateJob(
	jobNamespace string,
	currentJob *batchv1.Job,
	deployment *appsv1.Deployment,
) error {
	err := r.Client.Delete(context.TODO(), currentJob, k8sClient.PropagationPolicy(metav1.DeletePropagationBackground))
	if err != nil && !k8sErrors.IsNotFound(err) {
		return errors.Wrapf(err, "failed to delete outdated update job")
	}

	job := prepareJobTemplate(updateJobName, jobNamespace, deployment)

	err = r.Client.Create(context.TODO(), job)
	if err != nil {
		return err
	}

	return nil
}

func prepareJobTemplate(name, namespace string, baseDeployment *appsv1.Deployment) *batchv1.Job {
	backoffLimit := int32(10)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": updateJobName},
				},
				Spec: *baseDeployment.Spec.Template.Spec.DeepCopy(),
			},
			BackoffLimit: &backoffLimit,
		},
	}

	// Remove init container that waits for db setup job.
	containerToRemove, found := findContainer(mattermostApp.WaitForDBSetupContainerName, job.Spec.Template.Spec.InitContainers)
	if found {
		job.Spec.Template.Spec.InitContainers = append(
			job.Spec.Template.Spec.InitContainers[:containerToRemove],
			job.Spec.Template.Spec.InitContainers[containerToRemove+1:]...)
	}

	// We dont need to validate the readiness/liveness for this short lived job.
	for i := range job.Spec.Template.Spec.Containers {
		job.Spec.Template.Spec.Containers[i].LivenessProbe = nil
		job.Spec.Template.Spec.Containers[i].ReadinessProbe = nil
	}

	// Override values for job-specific behavior.
	job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyNever
	for i := range job.Spec.Template.Spec.Containers {
		job.Spec.Template.Spec.Containers[i].Command = []string{"mattermost", "version"}
	}

	return job
}

// isMainDeploymentContainerImageSame checks whether main containers of specified deployments are the same or not.
func (r *MattermostReconciler) isMainDeploymentContainerImageSame(
	a *appsv1.Deployment,
	b *appsv1.Deployment,
) (bool, error) {
	// Sanity check
	if (a == nil) || (b == nil) {
		return false, errors.New("failed to find main container, no deployment provided")
	}

	isSameImage, err := r.isMainContainerImageSame(
		a.Spec.Template.Spec.Containers,
		b.Spec.Template.Spec.Containers,
	)
	if err != nil {
		return false, errors.Wrapf(err, "failed to compare deployment images, deployments: %s/%s, %s/%s", a.Namespace, a.Name, b.Namespace, b.Name)
	}

	return isSameImage, nil
}

// isMainContainerImageSame checks whether main containers of specified slices are the same or not.
func (r *MattermostReconciler) isMainContainerImageSame(
	a []corev1.Container,
	b []corev1.Container,
) (bool, error) {
	// Fetch containers to compare
	containerA := mattermostv1beta1.GetMattermostAppContainer(a)
	if containerA == nil {
		return false, errors.Errorf("failed to find main container in a list while comparing images")
	}
	containerB := mattermostv1beta1.GetMattermostAppContainer(b)
	if containerB == nil {
		return false, errors.Errorf("failed to find main container in a list while comparing images")
	}

	// Both containers fetched, can compare images
	return containerA.Image == containerB.Image, nil
}

// updateMattermostDeployment performs deployment update if necessary.
// If a deployment update is necessary, an update job is launched to check new image.
func (r *MattermostReconciler) updateMattermostDeployment(
	mattermost *mattermostv1beta1.Mattermost,
	current *appsv1.Deployment,
	desired *appsv1.Deployment,
	reqLogger logr.Logger,
) error {
	sameImage, err := r.isMainDeploymentContainerImageSame(current, desired)
	if err != nil {
		return err
	}

	if sameImage {
		// Need to update other fields only, update job is not required
		return r.ResCreator.Update(current, desired, reqLogger)
	}

	// Image is not the same
	// Run a single-pod job with the new mattermost image
	// It will check whether new image is operational
	// and may perform any database migrations before altering the deployment.
	// If this fails, we will return and not upgrade the deployment.

	reqLogger.Info("Current image is not the same as the requested, will upgrade the Mattermost installation")

	job, err := r.checkUpdateJob(mattermost.Namespace, desired, reqLogger)
	if job != nil {
		// Job is done, need to cleanup
		defer r.cleanupUpdateJob(job, reqLogger)
	}
	if err != nil {
		return err
	}

	// Job completed successfully

	return r.ResCreator.Update(current, desired, reqLogger)
}

// checkUpdateJob checks whether update job status. In case job is not running it is launched
func (r *MattermostReconciler) checkUpdateJob(
	jobNamespace string,
	baseDeployment *appsv1.Deployment,
	reqLogger logr.Logger,
) (*batchv1.Job, error) {
	reqLogger.Info(fmt.Sprintf("Running Mattermost update image job check for image %s", mattermostv1beta1.GetMattermostAppContainerFromDeployment(baseDeployment).Image))
	job, err := r.fetchRunningUpdateJob(jobNamespace)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			reqLogger.Info("Launching update image job")
			if err = r.launchUpdateJob(jobNamespace, baseDeployment); err != nil {
				return nil, errors.Wrap(err, "Launching update image job failed")
			}
			return nil, errors.New("Began update image job")
		}

		return nil, errors.Wrap(err, "failed to determine if an update image job is already running")
	}

	// Job is either running or completed

	// If desired deployment image does not match the one used by update job, restart it.
	isSameImage, err := r.isMainContainerImageSame(
		baseDeployment.Spec.Template.Spec.Containers,
		job.Spec.Template.Spec.Containers,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to compare image of update job and desired deployment")
	}
	if !isSameImage {
		reqLogger.Info("Mattermost image changed, restarting update job")
		err := r.restartUpdateJob(jobNamespace, job, baseDeployment)
		if err != nil {
			return nil, errors.Wrap(err, "failed to restart update job")
		}

		return nil, errors.New("Restarted update image job")
	}

	if job.Status.CompletionTime == nil {
		return nil, errors.New("update image job still running")
	}

	// Job is completed, can check completion status

	if job.Status.Failed > 0 {
		return job, errors.New("update image job failed")
	}

	reqLogger.Info("Update image job ran successfully")

	return job, nil
}

// cleanupUpdateJob deletes update job and all pods of the job
func (r *MattermostReconciler) cleanupUpdateJob(job *batchv1.Job, reqLogger logr.Logger) {
	reqLogger.Info(fmt.Sprintf("Deleting update image job %s/%s", job.GetNamespace(), job.GetName()))

	err := r.Client.Delete(context.TODO(), job, k8sClient.PropagationPolicy(metav1.DeletePropagationBackground))
	if err != nil {
		// Do not return error on fail as it is not critical
		reqLogger.Error(err, "Unable to cleanup update image job")
	}
}

// fetchRunningUpdateJob gets update job
func (r *MattermostReconciler) fetchRunningUpdateJob(namespace string) (*batchv1.Job, error) {
	job := &batchv1.Job{}
	err := r.Client.Get(
		context.TODO(),
		types.NamespacedName{
			Name:      updateJobName,
			Namespace: namespace,
		},
		job,
	)
	return job, err
}

func findContainer(name string, containers []corev1.Container) (int, bool) {
	for i, cont := range containers {
		if cont.Name == name {
			return i, true
		}
	}
	return -1, false
}
