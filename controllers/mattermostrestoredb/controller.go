package mattermostrestoredb

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mattermostcomv1alpha1 "github.com/mattermost/mattermost-operator/api/v1alpha1"
)

// MattermostRestoreDBReconciler reconciles a MattermostRestoreDB object
type MattermostRestoreDBReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mattermost.com.mattermost.com,resources=mattermostrestoredbs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mattermost.com.mattermost.com,resources=mattermostrestoredbs/status,verbs=get;update;patch

func (r *MattermostRestoreDBReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("mattermostrestoredb", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *MattermostRestoreDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mattermostcomv1alpha1.MattermostRestoreDB{}).
		Complete(r)
}
