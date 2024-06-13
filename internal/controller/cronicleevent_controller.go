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

	"errors"
	//"fmt"
	cronicleApiClient "github.com/yasinahlattci/cronicle-go-client/api"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

func (r *CronicleEventReconciler) getFirstMatchingService(ctx context.Context, namespace string, instanceSelector *metav1.LabelSelector) (*corev1.Service, error) {
	// Convert to selector
	selector, err := metav1.LabelSelectorAsSelector(instanceSelector)
	if err != nil {
		return nil, err
	}

	serviceList := &corev1.ServiceList{}
	listOptions := &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: selector,
	}
	err = r.List(ctx, serviceList, listOptions)
	if err != nil {
		return nil, err
	}

	if len(serviceList.Items) == 0 {
		return nil, errors.New("no matching services found")
	}

	// Return the first matching service
	return &serviceList.Items[0], nil
}

func (r *CronicleEventReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	// Get the service URL

	cronicleEvent := &croniclenetv1.CronicleEvent{}

	err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, cronicleEvent)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	cronicleClient := cronicleApiClient.NewClient(cronicleApiClient.Config{
		BaseUrl:       "http://localhost:3012",
		APIKey:        "b488c195302bae22908c1b89e94b9c14",
		Timeout:       10 * time.Second,
		RetryAttempts: 2,
	})

	//labelSelector := cronicleEvent.Spec.InstanceSelector

	//service, err := r.getFirstMatchingService(ctx, req.Namespace, labelSelector)
	//if err != nil {
	//	l.Error(err, "No instance found for the event")
	//	return ctrl.Result{}, err
	//}
	//serviceUrl := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", service.Spec.ClusterIP, service.Namespace, service.Spec.Ports[0].Port)
	//l.Info("Service URL", "serviceUrl", serviceUrl)

	// Check if the event is being deleted
	if cronicleEvent.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(cronicleEvent, "cronicle.net/eventfinalizer") {
			if cronicleEvent.Status.EventStatus == "readyForDeletion" {
				resp, err := cronicleClient.CheckRunningJobs(cronicleEvent.Status.EventId)
				if err != nil {
					l.Error(err, "Failed to check running jobs")
					return ctrl.Result{}, err
				}
				if resp {
					l.Info("Event has running jobs, doing nothing")
					return ctrl.Result{RequeueAfter: 60 * time.Second}, nil
				}
				cronicleClient.DeleteEvent(cronicleEvent.Status.EventId)
				controllerutil.RemoveFinalizer(cronicleEvent, "cronicle.net/eventfinalizer")
				err = r.Update(ctx, cronicleEvent)
				if err != nil {
					l.Error(err, "Failed to remove finalizer")
					return ctrl.Result{}, err
				}

			}
			if cronicleEvent.Status.EventStatus != "markedForDeletion" {

				resp, err := cronicleClient.DisableEvent(cronicleEvent.Status.EventId)
				if err != nil {
					l.Error(err, "Failed to disable event")
					return ctrl.Result{}, err
				}
				l.Info("Event disabled", "resp", resp)
				cronicleEvent.Status.EventStatus = "markedForDeletion"
				err = r.Status().Update(ctx, cronicleEvent)
				return ctrl.Result{}, nil
			}
		}
	}

	if !controllerutil.ContainsFinalizer(cronicleEvent, "cronicle.net/eventfinalizer") {
		controllerutil.AddFinalizer(cronicleEvent, "cronicle.net/eventfinalizer")
		r.Update(ctx, cronicleEvent)
	}

	eventStatus := cronicleEvent.Status.EventStatus
	eventId := cronicleEvent.Status.EventId
	modified := time.Now().Unix()

	cronicleEvent.Status.EventStatus = "ready"
	cronicleEvent.Status.Modified = modified

	if eventStatus == "ready" && eventId != "" {
		l.Info("Event is ready", "eventId", eventId)

		updateEventData := cronicleApiClient.UpdateEventRequest{
			Id:            cronicleEvent.Status.EventId,
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
			Notes:         "Hello from go client",
			NotifyFail:    "",
			NotifySuccess: "",
			Params: map[string]interface{}{
				"script":   "#!/bin/sh\n\nsleep 100",
				"annotate": 1,
				"json":     0,
			},
			Plugin:     "shellplug",
			Retries:    0,
			RetryDelay: 30,
			Target:     "6712b650ec8b",
			Timeout:    3600,
			Timezone:   "Europe/Istanbul",
			Timing: map[string]interface{}{
				"days":    []int{1, 2, 3, 4, 5}, // Monday to Friday
				"hours":   []int{21},
				"minutes": []int{20, 40},
			},
			Title:   "Go Operator Test Cron",
			WebHook: "",
		}
		// It means event is already created, only update can be done, since delete is handled above
		resp, err := cronicleClient.UpdateEvent(updateEventData)
		if err != nil {
			l.Error(err, "Failed to update event")
			return ctrl.Result{}, err
		}
		l.Info("Event updated", "resp", resp)
		r.Status().Update(ctx, cronicleEvent)
		return ctrl.Result{}, nil
	}
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
		Notes:         "Hello from go client",
		NotifyFail:    "",
		NotifySuccess: "",
		Params: map[string]interface{}{
			"script":   "#!/bin/sh\n\nsleep 100",
			"annotate": 1,
			"json":     0,
		},
		Plugin:     "shellplug",
		Retries:    0,
		RetryDelay: 30,
		Target:     "6712b650ec8b",
		Timeout:    3600,
		Timezone:   "Europe/Istanbul",
		Timing: map[string]interface{}{
			"days":    []int{1, 2, 3, 4, 5}, // Monday to Friday
			"hours":   []int{21},
			"minutes": []int{20, 40},
		},
		Title:   "Go Operator Test Cron",
		WebHook: "",
	}
	resp, err := cronicleClient.CreateEvent(createEventData)
	cronicleEvent.Status.EventId = resp.ID
	if err != nil {
		l.Error(err, "Failed to create event")
		return ctrl.Result{}, err
	}
	l.Info("Event created", "resp", resp)
	r.Status().Update(ctx, cronicleEvent)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CronicleEventReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&croniclenetv1.CronicleEvent{}).
		Complete(r)
}
