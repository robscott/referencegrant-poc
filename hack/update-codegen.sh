echo "Generating v1alpha1 CRDs and deepcopy"
go run sigs.k8s.io/controller-tools/cmd/controller-gen \
        object:headerFile=./hack/boilerplate/boilerplate.generatego.txt \
        crd:crdVersions=v1 \
        output:crd:artifacts:config=config/crd \
        paths=./apis/v1alpha1