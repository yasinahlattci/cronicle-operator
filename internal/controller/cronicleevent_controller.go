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
	"fmt"
	"github.com/yasinahlattci/cronicle-operator/pkg/cronicle_client"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"reflect"
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

	if !controllerutil.ContainsFinalizer(cronicleEvent, "cronicle.net/eventfinalizer") {
		controllerutil.AddFinalizer(cronicleEvent, "cronicle.net/eventfinalizer")
		err = r.Update(ctx, cronicleEvent)
		if err != nil {
			l.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	labelSelector := cronicleEvent.Spec.InstanceSelector

	service, err := r.getFirstMatchingService(ctx, req.Namespace, labelSelector)
	if err != nil {
		l.Error(err, "No instance found for the event")
		return ctrl.Result{}, err
	}
	serviceUrl := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", service.Name, service.Namespace, service.Spec.Ports[0].Port)

	cronicleClient := cronicle_client.NewClient(cronicle_client.Config{
		BaseUrl:       serviceUrl,
		APIKey:        os.Getenv("CRONICLE_API_KEY"),
		Timeout:       10 * time.Second,
		RetryAttempts: 2,
	})

	// Check if the event is being deleted
	if cronicleEvent.GetDeletionTimestamp() != nil {
		if cronicleEvent.Status.EventStatus == "markedForDeletion" {
			resp, err := cronicleClient.CheckRunningJobs(cronicleEvent.Status.EventId)
			if err != nil {
				l.Error(err, "Failed to check running jobs")
				return ctrl.Result{}, err
			}
			if resp {
				l.Info("Event has running jobs, queueing for deletion", "eventId", cronicleEvent.Status.EventId)
				return ctrl.Result{RequeueAfter: 60 * time.Second}, nil
			}

			err = cronicleClient.DeleteEvent(cronicleEvent.Status.EventId)

			if err != nil {
				l.Info("Failed to delete event", "eventId", cronicleEvent.Status.EventId)
				l.Info("Error", "err", err)
			}
			l.Info("Event deleted", "eventId", cronicleEvent.Status.EventId)
			cronicleEvent.Status.EventStatus = "readyForDeletion"
			controllerutil.RemoveFinalizer(cronicleEvent, "cronicle.net/eventfinalizer")
			err = r.Update(ctx, cronicleEvent)
			if err != nil {
				l.Error(err, "Failed to remove finalizer")
				return ctrl.Result{}, err
			}

		}
		if cronicleEvent.Status.EventStatus == "created" {
			err := cronicleClient.DisableEvent(cronicleEvent.Status.EventId)
			if err != nil {
				l.Error(err, "Failed to disable event")
				return ctrl.Result{}, err
			}
			l.Info("Event disabled", "resp", cronicleEvent.Status.EventId)
			cronicleEvent.Status.EventStatus = "markedForDeletion"
			err = r.Status().Update(ctx, cronicleEvent)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil
	}

	eventStatus := cronicleEvent.Status.EventStatus
	eventId := cronicleEvent.Status.EventId
	modifiedDate := time.Now().Unix()
	cronicleEvent.Status.Modified = modifiedDate

	if eventStatus == "" && eventId == "" {
		createEventData := cronicle_client.CreateEventRequest{
			CatchUp:       cronicleEvent.Spec.CatchUp,
			Category:      cronicleEvent.Spec.Category,
			CpuLimit:      cronicleEvent.Spec.CpuLimit,
			CpuSustain:    cronicleEvent.Spec.CpuSustain,
			Detached:      cronicleEvent.Spec.Detached,
			Enabled:       cronicleEvent.Spec.Enabled,
			LogMaxSize:    cronicleEvent.Spec.LogMaxSize,
			MaxChildren:   cronicleEvent.Spec.MaxChildren,
			MemoryLimit:   cronicleEvent.Spec.MemoryLimit,
			MemorySustain: cronicleEvent.Spec.MemorySustain,
			Multiplex:     cronicleEvent.Spec.Multiplex,
			Notes:         cronicleEvent.Spec.Notes,
			NotifyFail:    cronicleEvent.Spec.NotifyFail,
			NotifySuccess: cronicleEvent.Spec.NotifySuccess,
			Plugin:        cronicleEvent.Spec.Plugin,
			Retries:       cronicleEvent.Spec.Retries,
			RetryDelay:    cronicleEvent.Spec.RetryDelay,
			Target:        cronicleEvent.Spec.Target,
			Timeout:       cronicleEvent.Spec.Timeout,
			Timezone:      cronicleEvent.Spec.Timezone,
			Title:         cronicleEvent.Spec.Title,
			WebHook:       cronicleEvent.Spec.WebHook,
			Timing:        cronicleEvent.Spec.Timing,
			Params:        cronicleEvent.Spec.Params,
			Algorithm:     cronicleEvent.Spec.Algorithm,
		}
		eventID, err := cronicleClient.CreateEvent(createEventData)
		cronicleEvent.Status.EventId = eventID
		cronicleEvent.Status.EventStatus = "created"
		if err != nil {
			l.Error(err, "Failed to create event")
			return ctrl.Result{}, err
		}
		l.Info("Event created", "resp", eventID)
		cronicleEvent.Status.LastHandledSpec = cronicleEvent.Spec
		r.Status().Update(ctx, cronicleEvent)
		return ctrl.Result{}, nil
	}

	if !reflect.DeepEqual(cronicleEvent.Spec, cronicleEvent.Status.LastHandledSpec) {
		updateEventData := cronicle_client.UpdateEventRequest{
			Id:            cronicleEvent.Status.EventId,
			CatchUp:       cronicleEvent.Spec.CatchUp,
			Category:      cronicleEvent.Spec.Category,
			CpuLimit:      cronicleEvent.Spec.CpuLimit,
			CpuSustain:    cronicleEvent.Spec.CpuSustain,
			Detached:      cronicleEvent.Spec.Detached,
			Enabled:       cronicleEvent.Spec.Enabled,
			LogMaxSize:    cronicleEvent.Spec.LogMaxSize,
			MaxChildren:   cronicleEvent.Spec.MaxChildren,
			MemoryLimit:   cronicleEvent.Spec.MemoryLimit,
			MemorySustain: cronicleEvent.Spec.MemorySustain,
			Multiplex:     cronicleEvent.Spec.Multiplex,
			Notes:         cronicleEvent.Spec.Notes,
			NotifyFail:    cronicleEvent.Spec.NotifyFail,
			NotifySuccess: cronicleEvent.Spec.NotifySuccess,
			Plugin:        cronicleEvent.Spec.Plugin,
			Retries:       cronicleEvent.Spec.Retries,
			RetryDelay:    cronicleEvent.Spec.RetryDelay,
			Target:        cronicleEvent.Spec.Target,
			Timeout:       cronicleEvent.Spec.Timeout,
			Timezone:      cronicleEvent.Spec.Timezone,
			Title:         cronicleEvent.Spec.Title,
			WebHook:       cronicleEvent.Spec.WebHook,
			Timing:        cronicleEvent.Spec.Timing,
			Params:        cronicleEvent.Spec.Params,
			Algorithm:     cronicleEvent.Spec.Algorithm,
		}
		// It means event is already created, only update can be done, since delete is handled above
		err := cronicleClient.UpdateEvent(updateEventData)
		if err != nil {
			l.Error(err, "Failed to update event")
			return ctrl.Result{}, err
		}
		l.Info("Event updated", "resp", cronicleEvent.Status.EventId)
		cronicleEvent.Status.LastHandledSpec = cronicleEvent.Spec
		r.Status().Update(ctx, cronicleEvent)
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *CronicleEventReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&croniclenetv1.CronicleEvent{}).
		Complete(r)
}
