package v1alpha1

import (
	v1alpha1 "code.sapcloud.io/kubernetes/node-controller-manager/pkg/apis/node/v1alpha1"
	"code.sapcloud.io/kubernetes/node-controller-manager/pkg/client/clientset/scheme"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	rest "k8s.io/client-go/rest"
)

type NodeV1alpha1Interface interface {
	RESTClient() rest.Interface
	AWSInstanceClassesGetter
	InstancesGetter
	InstanceDeploymentsGetter
	InstanceSetsGetter
	InstanceTemplatesGetter
	ScalesGetter
}

// NodeV1alpha1Client is used to interact with features provided by the node.sapcloud.io group.
type NodeV1alpha1Client struct {
	restClient rest.Interface
}

func (c *NodeV1alpha1Client) AWSInstanceClasses() AWSInstanceClassInterface {
	return newAWSInstanceClasses(c)
}

func (c *NodeV1alpha1Client) Instances() InstanceInterface {
	return newInstances(c)
}

func (c *NodeV1alpha1Client) InstanceDeployments() InstanceDeploymentInterface {
	return newInstanceDeployments(c)
}

func (c *NodeV1alpha1Client) InstanceSets() InstanceSetInterface {
	return newInstanceSets(c)
}

func (c *NodeV1alpha1Client) InstanceTemplates(namespace string) InstanceTemplateInterface {
	return newInstanceTemplates(c, namespace)
}

func (c *NodeV1alpha1Client) Scales(namespace string) ScaleInterface {
	return newScales(c, namespace)
}

// NewForConfig creates a new NodeV1alpha1Client for the given config.
func NewForConfig(c *rest.Config) (*NodeV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &NodeV1alpha1Client{client}, nil
}

// NewForConfigOrDie creates a new NodeV1alpha1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *NodeV1alpha1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new NodeV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *NodeV1alpha1Client {
	return &NodeV1alpha1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1alpha1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *NodeV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}