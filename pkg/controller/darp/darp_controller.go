package darp

import (
	"context"
	"github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	oktov1alpha1 "github.com/darp-operator/pkg/apis/okto/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	//"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	routev1 "github.com/openshift/api/route/v1"
)

var log = logf.Log.WithName("controller_darp")

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDarp{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("darp-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	// Watch for changes to primary resource Darp
	err = c.Watch(&source.Kind{Type: &oktov1alpha1.Darp{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	// Watch for changes to Deployment
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &oktov1alpha1.Darp{},
	})
	if err != nil {
		return err
	}
	// Watch for changes to Service
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &oktov1alpha1.Darp{},
	})
	if err != nil {
		return err
	}
	// Watch for changes to Secret
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &oktov1alpha1.Darp{},
	})
	if err != nil {
		return err
	}
	// Watch for changes to ConfigMap
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &oktov1alpha1.Darp{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileDarp implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileDarp{}

type ReconcileDarp struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

func (r *ReconcileDarp) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Darp")

	// Fetch the Darp darp
	darp := &oktov1alpha1.Darp{}
	err := r.client.Get(context.TODO(), request.NamespacedName, darp)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	//Check if Root CA certs secret already exists, if not create a new one
	rootCaCertsSecret := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: darp.Spec.RootCaSecretName, Namespace: darp.Namespace}, rootCaCertsSecret)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Root CA Secret
		rootCaCert, err := r.rootCaSecretForDarp(darp)
		if err != nil {
			reqLogger.Error(err, "error getting root ca secret")
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new Root CA Secret.", "Secret.Namespace", rootCaCert.Namespace, "Secret.Name", rootCaCert.Name)
		err = r.client.Create(context.TODO(), rootCaCert)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Root CA Certs secrets.", "Secret.Namespace", rootCaCert.Namespace, "Secret.Name", rootCaCert.Name)
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment.")
		return reconcile.Result{}, err
	}
	// Define a new Pod object
	//route := newRouteForCR(darp)
	//
	//// Set Darp darp as the owner and controller
	//if err := controllerutil.SetControllerReference(darp, route, r.scheme); err != nil {
	//	return reconcile.Result{}, err
	//}
	//
	////Check if this Route already exists
	//found := &routev1.Route{}
	//err = r.client.Get(context.TODO(), types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, found)
	//if err != nil && errors.IsNotFound(err) {
	//	reqLogger.Info("Creating a new Route", "Route.Namespace", route.Namespace, "Route.Name", route.Name)
	//	err = r.client.Create(context.TODO(), route)
	//	if err != nil {
	//		return reconcile.Result{}, err
	//	}
	//	// Pod created successfully - don't requeue
	//	return reconcile.Result{}, nil
	//} else if err != nil {
	//	return reconcile.Result{}, err
	//}

	// Pod already exists - don't requeue
	//reqLogger.Info("Skip reconcile: Route already exists", "Route.Namespace", found.Namespace, "Route.Name", found.Name)
	return reconcile.Result{}, nil
}

func (r *ReconcileDarp) rootCaSecretForDarp(depr *oktov1alpha1.Darp) (*corev1.Secret, error) {
	caCerts := CACerts{}
	if err := caCerts.generateRootCerts(); err != nil {
		log.Error(err, "Error during creating root certificates")
		return nil, err
	}
	labels := map[string]string{
		"app": depr.Name,
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "darp-root-ca-certs",
			Namespace: depr.Namespace,
			Labels:    labels,
		},
		StringData: map[string]string{
			"crt": string(caCerts.CAPem),
			"key": string(caCerts.CAPrivPem),
		},
	}
	if err := controllerutil.SetControllerReference(depr, secret, r.scheme); err != nil {
		log.Error(err, "Error set controller reference for root ca secret")
		return nil, err
	}
	return secret, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newRouteForCR(cr *oktov1alpha1.Darp) *routev1.Route {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &v1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-route",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			TLS: &routev1.TLSConfig{
				Termination: routev1.TLSTerminationEdge,
			},
			To: routev1.RouteTargetReference{
				Kind: "Service", Name: cr.Name + "-route",
			},
		},
	}
	//return &corev1.Pod{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      cr.Name + "-pod",
	//		Namespace: cr.Namespace,
	//		Labels:    labels,
	//	},
	//	Spec: corev1.PodSpec{
	//		Containers: []corev1.Container{
	//			{
	//				Name:    "busybox",
	//				Image:   "busybox",
	//				Command: []string{"sleep", "3600"},
	//			},
	//		},
	//	},
	//}
}
