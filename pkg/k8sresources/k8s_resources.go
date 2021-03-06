package k8sresources

import (
	"sort"
	"time"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kubectlfzf/pkg/util"
)

// K8sResource is the generic information of a k8s entity
type K8sResource interface {
	HasChanged(k K8sResource) bool
	ToString() string
	FromRuntime(obj interface{}, config CtorConfig)
}

// ResourceMeta is the generic information of a k8s entity
type ResourceMeta struct {
	name         string
	namespace    string
	cluster      string
	labels       map[string]string
	creationTime time.Time
}

// FromObjectMeta copies meta information to the object
func (r *ResourceMeta) FromObjectMeta(meta metav1.ObjectMeta, config CtorConfig) {
	r.name = meta.Name
	r.namespace = meta.Namespace
	r.cluster = config.Cluster
	r.labels = meta.Labels
	r.creationTime = meta.CreationTimestamp.Time
}

// FromDynamicMeta copies meta information to the object
func (r *ResourceMeta) FromDynamicMeta(u *unstructured.Unstructured, config CtorConfig) {
	metadata := u.Object["metadata"].(map[string]interface{})
	r.name = metadata["name"].(string)
	r.namespace = metadata["namespace"].(string)
	r.cluster = config.Cluster
	var err error
	var found bool
	r.labels, found, err = unstructured.NestedStringMap(u.Object, "metadata", "labels")
	util.FatalIf(err)
	if !found {
		glog.V(3).Infof("metadata.labels was not found in %#v", u.Object)
	}
	r.creationTime, err = time.Parse(time.RFC3339, metadata["creationTimestamp"].(string))
	util.FatalIf(err)
}

func (r *ResourceMeta) resourceAge() string {
	return util.TimeToAge(r.creationTime)
}

// ExcludedLabels is a list of excluded label/selector from the dump
var ExcludedLabels = map[string]string{"pod-template-generation": "",
	"app.kubernetes.io/name": "", "controller-revision-hash": "",
	"app.kubernetes.io/managed-by": "", "pod-template-hash": "",
	"statefulset.kubernetes.io/pod-name": "",
	"controler-uid":                      ""}

func (r *ResourceMeta) labelsString() string {
	if len(r.labels) == 0 {
		return "None"
	}
	els := util.JoinStringMap(r.labels, ExcludedLabels, "=")
	sort.Strings(els)
	return util.JoinSlicesOrNone(els, ",")
}
