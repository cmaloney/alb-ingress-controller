package controller

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/coreos/alb-ingress-controller/awsutil"
	"github.com/coreos/alb-ingress-controller/controller/config"
	"github.com/coreos/alb-ingress-controller/log"
	"github.com/golang/glog"
	"github.com/spf13/pflag"

	api "k8s.io/client-go/pkg/api/v1"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/ingress/core/pkg/ingress"
	"k8s.io/ingress/core/pkg/ingress/defaults"
)

// ALBController is our main controller
type ALBController struct {
	storeLister  ingress.StoreLister
	ALBIngresses ALBIngressesT
	clusterName  *string
	IngressClass string
}

// NewALBController returns an ALBController
func NewALBController(awsconfig *aws.Config, conf *config.Config) *ALBController {
	ac := &ALBController{
		clusterName: aws.String(conf.ClusterName),
	}

	// TODO(cmaloney): Move this all to OverrideFlags
	awsutil.AWSDebug = conf.AWSDebug
	awsutil.Session = awsutil.NewSession(awsconfig)
	awsutil.ALBsvc = awsutil.NewELBV2(awsutil.Session)
	awsutil.Ec2svc = awsutil.NewEC2(awsutil.Session)
	awsutil.ACMsvc = awsutil.NewACM(awsutil.Session)

	return ac
}

// OnUpdate is a callback invoked from the sync queue when ingress resources, or resources ingress
// resources touch, change. On each new event a new list of ALBIngresses are created and evaluated
// against the existing ALBIngress list known to the ALBController. Eventually the state of this
// list is synced resulting in new ingresses causing resource creation, modified ingresses having
// resources modified (when appropriate) and ingresses missing from the new list deleted from AWS.
func (ac *ALBController) OnUpdate(ic ingress.Configuration) ([]byte, error) {

	log.Debugf("OnUpdate event seen by ALB ingress controller.", "controller")

	// ingressConfiguration.Backends has all the backends which we need to convert to TargetGroups
	// but unfortunately AWS ALBs require that a target is only used by a single ALB, so we need
	// to make TargetGroups per ALB created.

	// For now ignoring TCPEndpoints and UDPEndpoints
	var lbs LBController
	lbs.UnmarkIngresses()

	// Convert servers into individual ALBs
	for server = range ic.Servers {
		newLb := lbs.FromIngress(server)
	}

	panic("TODO")

	// Create new ALBIngress list for this invocation.
	var ALBIngresses ALBIngressesT
	// Find every ingress currently in Kubernetes.
	for _, ingress := range ac.storeLister.Ingress.List() {
		ingResource := ingress.(*extensions.Ingress)
		// Ensure the ingress resource found contains an appropriate ingress class.
		if !ac.validIngress(ingResource) {
			continue
		}
	}

	return []byte(""), nil
}

// validIngress checks whether the ingress controller has an IngressClass set. If it does, it will
// only return true if the ingress resource passed in has the same class specified via the
// kubernetes.io/ingress.class annotation.
func (ac ALBController) validIngress(i *extensions.Ingress) bool {
	if ac.IngressClass == "" {
		return true
	}
	if i.Annotations["kubernetes.io/ingress.class"] == ac.IngressClass {
		return true
	}
	return false
}

// Reload executes the state synchronization for our ingresses
func (ac *ALBController) Reload(data []byte) ([]byte, bool, error) {
	awsutil.ReloadCount.Add(float64(1))

	panic("TODO")

	return []byte(""), true, nil
}

// OverrideFlags configures optional override flags for the ingress controller
func (ac *ALBController) OverrideFlags(flags *pflag.FlagSet) {
}

// SetConfig configures a configmap for the ingress controller
func (ac *ALBController) SetConfig(cfgMap *api.ConfigMap) {
	glog.Infof("Config map %+v", cfgMap)
}

// SetListers sets the configured store listers in the generic ingress controller
func (ac *ALBController) SetListers(lister ingress.StoreLister) {
	ac.storeLister = lister
}

// BackendDefaults returns default configurations for the backend
func (ac *ALBController) BackendDefaults() defaults.Backend {
	var backendDefaults defaults.Backend
	return backendDefaults
}

// Name returns the ingress controller name
func (ac *ALBController) Name() string {
	return "AWS Application Load Balancer Controller"
}

// Check tests the ingress controller configuration
func (ac *ALBController) Check(_ *http.Request) error {
	// TODO(cmaloney): Validate that the current ingresses match the expected ingresses.
	return nil
}

// DefaultIngressClass returns thed default ingress class
func (ac *ALBController) DefaultIngressClass() string {
	return "alb"
}

// Info returns information on the ingress contoller
func (ac *ALBController) Info() *ingress.BackendInfo {
	return &ingress.BackendInfo{
		Name:       "ALB Ingress Controller",
		Release:    "0.0.1",
		Build:      "git-00000000",
		Repository: "git://github.com/coreos/alb-ingress-controller",
	}
}
