type CommonConfiguration struct {
    // NodeSelector is a selector which must be true for the pod to fit on a node.
    // Selector which must match a node's labels for the pod to be scheduled on that node.
    // More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
    // +optional
    NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,7,rep,name=nodeSelector"`
    // Host networking requested for this pod. Use the host's network namespace.
    // If this option is set, the ports that will be used must be specified.
    // Default to false.
    // +k8s:conversion-gen=false
    // +optional
    HostNetwork bool `json:"hostNetwork,omitempty" protobuf:"varint,11,opt,name=hostNetwork"`
    // ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
    // If specified, these secrets will be passed to individual puller implementations for them to use. For example,
    // in the case of docker, only DockerConfig type secrets are honored.
    // More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
    // +optional
    // +patchMergeKey=name
    // +patchStrategy=merge
    ImagePullSecrets []LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"`
    // If specified, the pod's tolerations.
    // +optional
    Tolerations []Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`
    // Number of desired pods. This is a pointer to distinguish between explicit
    // zero and not specified. Defaults to 1.
    // +optional
    Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`
}

type VrouterSpec struct {
    CommonConfiguration ServiceCommon
    VrouterConfiguration VrouterConfiguration
    Profiles []Profiles
}

type VrouterConfiguration struct {

}