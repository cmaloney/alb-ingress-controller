package main

import (
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/golang/glog"

	"github.com/coreos/alb-ingress-controller/controller"
	"github.com/coreos/alb-ingress-controller/controller/config"
	"github.com/coreos/alb-ingress-controller/log"
	ingresscontroller "k8s.io/ingress/core/pkg/ingress/controller"
)

func main() {
	clusterName := os.Getenv("CLUSTER_NAME")
	if clusterName == "" {
		glog.Exit("A CLUSTER_NAME environment variable must be defined")
	}

	logLevel := os.Getenv("LOG_LEVEL")
	log.SetLogLevel(logLevel)

	awsDebug, _ := strconv.ParseBool(os.Getenv("AWS_DEBUG"))

	conf := &config.Config{
		ClusterName: clusterName,
		AWSDebug:    awsDebug,
	}

	if len(clusterName) > 11 {
		glog.Exit("CLUSTER_NAME must be 11 characters or less")
	}

	ac := controller.NewALBController(&aws.Config{MaxRetries: aws.Int(5)}, conf)
	ic := ingresscontroller.NewIngressController(ac)

	ac.IngressClass = ic.IngressClass()
	if ac.IngressClass != "" {
		log.Infof("Ingress class set to %s", "controller", ac.IngressClass)
	}
	// All flags are parsed / config is setup. Use that config to run a first
	// alb controller sync

	//

	defer func() {
		glog.Infof("Shutting down ingress controller...")
		ic.Stop()
	}()
	ic.Start()
}
