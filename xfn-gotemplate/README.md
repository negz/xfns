# Crossplane Composition Function for [Go Templates](https://pkg.go.dev/text/template)

Experimental composition function that renders given templates using the
observed state in the `FunctionIO`.

## Quick Start

1. Install Crossplane with the composition functions enabled:

```console
helm repo add crossplane-master https://charts.crossplane.io/master --force-update
helm repo update

kubectl create namespace crossplane-system

helm install crossplane --namespace crossplane-system crossplane-master/crossplane --devel \
  --set xfn.enabled=true \
  --set "args={--debug,--enable-composition-functions}" \
  --set "xfn.args={--debug}"
```

2. Apply example Composition and CompositeResourceDefinition:

```console
kubectl apply -f example/definition.yaml
kubectl apply -f example/composition.yaml
```

3. Deploy Provider GCP and Provider Kubernetes:

```console
cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-gcp
spec:
  package: xpkg.upbound.io/upbound/provider-gcp:v0.25.0
EOF

cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-kubernetes
spec:
  package: xpkg.upbound.io/crossplane-contrib/provider-kubernetes:v0.6.0
EOF
```

4. Create the example Composite Resource:

```console
kubectl apply -f example/composite.yaml
```

5. Check composed resources:

```console 
watch kubectl get managed
```

6. Change the values of `spec.parameters.createSecret` and `spec.parameters.configMapCount`
parameters to see the changes in the composed resources.

```console
kubectl edit -f example/composite.yaml
```

## Changing Templates

### Option A: Rebuild the Function Image with New Templates

1. Change the templates in `examples/templates` directory.
2. Build docker image with the new templates:

```console
docker build . -t <your-dockerhub-account>/xfn-gotemplate:v0.0.0
docker push  <your-dockerhub-account>/xfn-gotemplate:v0.0.0
```

3. Update the image for function in `example/composition.yaml` and reapply.

### Option B: Build your own Image overwriting the Templates from the base Image

1. Define a new Dockerfile as follows:

```Dockerfile
FROM turkenh/xfn-gotemplate:v0.1.0

COPY templates /templates
```

2. Create the following directory structure with the Dockerfile above and your templates:

```console
.
├── Dockerfile
└── templates
    ├── resource1.yaml
    ├── resource2.yaml
    └── resource3.yaml
```

3. Build your image:

```console
docker build . -t <your-dockerhub-account>/my-templates:v0.0.0
docker push  <your-dockerhub-account>/my-templates:v0.0.0
```