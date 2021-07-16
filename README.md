# mastro-operator

A walk through the creation an example operator using the operator SDK.

## Install the operator-sdk bin
* Download a suitable binary from [here](https://github.com/operator-framework/operator-sdk/releases)
* `sudo chmod +x /usr/local/bin/operator-sdk`

## Init a new operator project
```
export GO111MODULE=on
operator-sdk init \
--domain=data-mill.cloud \
--repo=github.com/pilillo/mastro-operator \
--license apache2 \
--owner "pilillo"
```

## Create a new resource API
```
operator-sdk create api \
    --version=v1alpha1 \
    --kind=Catalogue
```

A skeleton for the CR definition is now available at the `api/v1alpha1` subfolder.  
A skeleton for the controller is now available at the `controllers/catalogue_controller.go` subfolder.

## Generate CRD code at any modification of the struct type
```
make generate
```

## Generate RBAC manifests after setting the `+kubebuilder:rbac` markers
```
make manifests
```

## Controller

The `catalogue_controller.go` defines the controller logic:

The function `SetupWithManager` defines a new controller and specifies to watch **For** resources of kind `datamillcloudv1alpha1.Catalogue`.  
Thus, upon any add/update/delete event the reconcile loop will be sent a reconcile *Request* for the object.
Additional resources can be watched (using *Owns*) as well as additional options can be set using *WithOptions* (see [here](https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/#implement-the-controller))

```go
// SetupWithManager sets up the controller with the Manager.
func (r *CatalogueReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datamillcloudv1alpha1.Catalogue{}).
		Complete(r)
}
```

The *Reconcile* function implements the reconcile loop, which is passed the *Request* argument (a namespace/name key).

```go
func (r *CatalogueReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// your logic here

	return ctrl.Result{}, nil
}
```

So we can directly retrieve the watched resource and act accordingly:
```go
func (r *CatalogueReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// your logic here
	catalogue := &datamillcloudv1alpha1.Catalogue{}
	err := r.Get(ctx, req.NamespacedName, catalogue)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
```

Based on the returned error, the Request may be requeued and the reconcile triggered again:
```go
// Reconcile successful - don't requeue
return ctrl.Result{}, nil
// Reconcile failed due to error - requeue
return ctrl.Result{}, err
// Requeue for any reason other than an error
return ctrl.Result{Requeue: true}, nil
```

## Testing the operator
The operator can be run either outside or on the cluster.  

To install the CRD:
```bash
make install
```

To run the controller on the host (outside the cluster):
```bash
make run .
```

To deploy the controller as deployment on the cluster:
```bash
make deploy
```

Alternatively, the [operator lifecycle manager (OLM)](https://sdk.operatorframework.io/docs/olm-integration/quickstart-bundle/#enabling-olm) can be used similarly to,  
run locally
```bash
operator-sdk run --local
```

## Building the operator
To build the operator image and push it to the desired registry the Makefile can be used.  
Specifically:
```bash
export IMG=datamillcloud/mastro-operator:v0.0.1
make docker-build
make docker-push
```

which will build and push the image to Dockerhub.  
The Makefile is like follows:  
```bash
docker-build: test ## Build docker image with the manager.
	docker build -t ${IMG} .

docker-push: ## Push docker image with the manager.
	docker push ${IMG}
```