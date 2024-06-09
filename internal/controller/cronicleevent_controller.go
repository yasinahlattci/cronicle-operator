/*
Copyright 2024 Yasin AHLATCI.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	cronicleApiClient "github.com/yasinahlattci/cronicle-go-client/api"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	croniclenetv1 "github.com/yasinahlattci/cronicle-operator/api/v1"
)

// CronicleEventReconciler reconciles a CronicleEvent object
type CronicleEventReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cronicle.net,resources=cronicleevents,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cronicle.net,resources=cronicleevents/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cronicle.net,resources=cronicleevents/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CronicleEvent object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
func (r *CronicleEventReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Reconciling CronicleEvent", "req", req)

	cronicleEvent := &croniclenetv1.CronicleEvent{}
	r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, cronicleEvent)

	eventStatus := cronicleEvent.Status.EventStatus
	eventId := cronicleEvent.Status.EventId
	if eventStatus != "" && eventId != "" {
		l.Info("Event already created", "eventStatus", eventStatus)
		return ctrl.Result{}, nil
	} else {
		cronicleClient := cronicleApiClient.NewClient(cronicleApiClient.Config{
			BaseUrl:       "http://localhost:3012",
			APIKey:        "b488c195302bae22908c1b89e94b9c14",
			Timeout:       10 * time.Second,
			RetryAttempts: 2,
		})
		createEventData := cronicleApiClient.CreateEventRequest{
			CatchUp:       1,
			Category:      "general",
			CpuLimit:      100,
			CpuSustain:    0,
			Detached:      0,
			Enabled:       cronicleEvent.Spec.Enabled,
			LogMaxSize:    0,
			MaxChildren:   1,
			MemoryLimit:   0,
			MemorySustain: 0,
			Multiplex:     0,
			Notes:         "Hello from operator",
			NotifyFail:    "",
			NotifySuccess: "",
			Params: map[string]interface{}{
				"db_host":          "idb01.mycompany.com",
				"verbose":          1,
				"cust":             "Sales",
				"additional_param": "value",
			},
			Plugin:     "test",
			Retries:    0,
			RetryDelay: 30,
			Target:     "db1.int.myserver.com",
			Timeout:    3600,
			Timezone:   "America/New_York",
			Timing: map[string]interface{}{
				"days":    []int{1, 2, 3, 4, 5}, // Monday to Friday
				"hours":   []int{21},
				"minutes": []int{20, 40},
			},
			Title:   "Hello from operator Test",
			WebHook: "http://myserver.com/notify-chronos.php",
		}
		resp, err := cronicleClient.CreateEvent(createEventData)
		if err != nil {
			l.Error(err, "Failed to create event")
			return ctrl.Result{}, err
		}
		l.Info("Event created", "resp", resp)
		cronicleEvent.Status.EventId = resp.ID
		cronicleEvent.Status.EventStatus = "ready"
		r.Status().Update(ctx, cronicleEvent)
		return ctrl.Result{}, nil

	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *CronicleEventReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&croniclenetv1.CronicleEvent{}).
		Complete(r)
}
