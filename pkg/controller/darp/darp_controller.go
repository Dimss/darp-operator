package darp

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	oktov1alpha1 "github.com/darp-operator/pkg/apis/okto/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
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

	// Check if Root CA certs secret already exists, if not create a new one
	rootCaCertsSecret := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: darp.Spec.RootCaSecret, Namespace: darp.Namespace}, rootCaCertsSecret)
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
		reqLogger.Error(err, "Failed to get Root CA Secret.")
		return reconcile.Result{}, err
	}

	// Check if server certs secret already exists, if not create a new one
	serverCertsSecret := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: darp.Spec.ServerCertsSecret, Namespace: darp.Namespace}, serverCertsSecret)
	if err != nil && errors.IsNotFound(err) {
		proxyServerCerts, err := r.serverCertsForDarp(darp)
		if err != nil {
			reqLogger.Error(err, "error getting server certs secret")
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new Root CA Secret.", "Secret.Namespace", serverCertsSecret.Namespace, "Secret.Name", serverCertsSecret.Name)
		err = r.client.Create(context.TODO(), proxyServerCerts)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Server Certs secret.", "Secret.Namespace", serverCertsSecret.Namespace, "Secret.Name", serverCertsSecret.Name)
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Server Secret.")
		return reconcile.Result{}, err
	}

	//Check if server config map already exists, if not create a new one
	serverConfigMap := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: darp.Spec.ServerConfigMap, Namespace: darp.Namespace}, serverConfigMap)
	if err != nil && errors.IsNotFound(err) {
		serverConf, err := r.configMapForDarp(darp)
		if err != nil {
			reqLogger.Error(err, "error getting server configmap")
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new server config map.", "ConfigMap.Namespace", serverConf.Namespace, "ConfigMap.Name", serverConf.Name)
		err = r.client.Create(context.TODO(), serverConf)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Server ConfigMap.", "ConfigMap.Namespace", serverConf.Namespace, "ConfigMap.Name", serverConf.Name)
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get server configmap.")
		return reconcile.Result{}, err
	}

	//Check if deployment already exists, if not create a new one
	deployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: darp.Name, Namespace: darp.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		serverDeployment, err := r.deploymentForDarp(darp)
		if err != nil {
			reqLogger.Error(err, "error getting server deployment")
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new server deployment.", "Deployment.Namespace", serverDeployment.Namespace, "Deployment.Name", serverDeployment.Name)
		err = r.client.Create(context.TODO(), serverDeployment)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Server Deployment.", "Deployment.Namespace", serverDeployment.Namespace, "Deployment.Name", serverDeployment.Name)
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get server deployment.")
		return reconcile.Result{}, err
	}

	//Check if service already exists, if not create a new one
	service := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: darp.Name, Namespace: darp.Namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		serverService, err := r.serviceForDarp(darp)
		if err != nil {
			reqLogger.Error(err, "error getting server service")
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new service.", "Service.Namespace", serverService.Namespace, "Service.Name", serverService.Name)
		err = r.client.Create(context.TODO(), serverService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Server Service.", "Service.Namespace", serverService.Namespace, "Service.Name", serverService.Name)
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get server service.")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileDarp) rootCaSecretForDarp(darp *oktov1alpha1.Darp) (*corev1.Secret, error) {
	caCerts := CACerts{}
	if err := caCerts.generateRootCerts(); err != nil {
		log.Error(err, "Error during creating root certificates")
		return nil, err
	}
	labels := map[string]string{
		"app": darp.Name,
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      darp.Spec.RootCaSecret,
			Namespace: darp.Namespace,
			Labels:    labels,
		},
		StringData: map[string]string{
			"crt": string(caCerts.CAPem),
			"key": string(caCerts.CAPrivPem),
		},
	}
	if err := controllerutil.SetControllerReference(darp, secret, r.scheme); err != nil {
		log.Error(err, "Error set controller reference for root ca secret")
		return nil, err
	}
	return secret, nil
}

func (r *ReconcileDarp) serverCertsForDarp(darp *oktov1alpha1.Darp) (*corev1.Secret, error) {

	rootCaCertsSecret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: darp.Spec.RootCaSecret, Namespace: darp.Namespace}, rootCaCertsSecret)
	if err != nil && errors.IsNotFound(err) {
		log.Error(err, "Root CA Certificate Secret not found")
		return nil, err
	}
	caCerts := CACerts{
		CAPem:     rootCaCertsSecret.Data["crt"],
		CAPrivPem: rootCaCertsSecret.Data["key"],
	}
	if err := caCerts.loadRootCertificates(); err != nil {
		log.Error(err, "Failed to load root certificates ")
		return nil, err
	}
	crt, key, err := caCerts.generateCertificates(darp.Name)
	if err != nil {
		log.Error(err, "Failed to generate server certificates")
		return nil, err
	}
	labels := map[string]string{
		"app": darp.Name,
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      darp.Spec.ServerCertsSecret,
			Namespace: darp.Namespace,
			Labels:    labels,
		},
		StringData: map[string]string{
			"crt": string(crt),
			"key": string(key),
		},
	}
	if err := controllerutil.SetControllerReference(darp, secret, r.scheme); err != nil {
		log.Error(err, "Error set controller reference for server certs")
		return nil, err
	}
	return secret, nil
}

func (r *ReconcileDarp) configMapForDarp(darp *oktov1alpha1.Darp) (*corev1.ConfigMap, error) {
	labels := map[string]string{
		"app": darp.Name,
	}
	data := map[string]string{
		"config.json": fmt.Sprintf("{\"http\": {\"crt\":\"%v/crt\",\"key\":\"%v/key\"},\"upstreams\": []}",
			darp.Spec.CertsMountPath,
			darp.Spec.CertsMountPath),
	}
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      darp.Spec.ServerConfigMap,
			Namespace: darp.Namespace,
			Labels:    labels,
		},
		Data: data,
	}
	if err := controllerutil.SetControllerReference(darp, configMap, r.scheme); err != nil {
		log.Error(err, "Error set controller reference for config map")
		return nil, err
	}
	return configMap, nil
}

func (r *ReconcileDarp) deploymentForDarp(darp *oktov1alpha1.Darp) (*appsv1.Deployment, error) {
	var replicas int32
	replicas = 1
	labels := map[string]string{
		"app": darp.Name,
	}
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      darp.Name,
			Namespace: darp.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": darp.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   darp.Name,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            darp.Name,
							Image:           darp.Spec.Image,
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "certs",
									MountPath: darp.Spec.CertsMountPath,
								},
								{
									Name:      "config",
									MountPath: darp.Spec.ConfMountPath,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "certs",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: darp.Spec.ServerCertsSecret,
								},
							},
						},
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: darp.Spec.ServerConfigMap,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(darp, dep, r.scheme); err != nil {
		log.Error(err, "Error set controller reference for server deployment")
		return nil, err
	}
	return dep, nil
}

func (r *ReconcileDarp) serviceForDarp(darp *oktov1alpha1.Darp) (*corev1.Service, error) {
	labels := map[string]string{
		"app": darp.Name,
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      darp.Name,
			Namespace: darp.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": darp.Name},
			Ports: []corev1.ServicePort{
				{
					Name:       "https",
					Port:       443,
					TargetPort: intstr.IntOrString{IntVal: 8080},
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(darp, service, r.scheme); err != nil {
		log.Error(err, "Error set controller reference for server service")
		return nil, err
	}
	return service, nil
}
