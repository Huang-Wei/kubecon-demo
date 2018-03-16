/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	clientset "github.com/Huang-Wei/kubecon-demo/pkg/client/clientset/versioned"
	informers "github.com/Huang-Wei/kubecon-demo/pkg/client/informers/externalversions"
)

var (
	masterURL  string
	kubeconfig string
	hostname   string
)

func main() {
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	// stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	// kubeClient, err := kubernetes.NewForConfig(cfg)
	kubeClient, err := kubernetes.NewForConfig(rest.AddUserAgent(cfg, "leader-election"))
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	exampleClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building example clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*0)
	exampleInformerFactory := informers.NewSharedInformerFactory(exampleClient, time.Second*0)

	controller := NewController(kubeClient, exampleClient, kubeInformerFactory, exampleInformerFactory)

	// wrap the controller starting logic in a block to pass into leaderelector helper
	run := func(stopCh <-chan struct{}) {
		go kubeInformerFactory.Start(stopCh)
		go exampleInformerFactory.Start(stopCh)

		if err = controller.Run(2, stopCh); err != nil {
			glog.Fatalf("Error running controller: %s", err.Error())
		}
	}

	// rl, err := resourcelock.New(resourcelock.ConfigMapsResourceLock,
	rl, err := resourcelock.New(resourcelock.EndpointsResourceLock,
		getNamespace(),
		"kubecon-demo-lock",
		kubeClient.CoreV1(),
		resourcelock.ResourceLockConfig{
			Identity:      getHostname(),
			EventRecorder: createRecorder(kubeClient, "kubecon-demo"),
		})
	if err != nil {
		glog.Fatalf("Error creating lock: %v", err)
	}

	leaderelection.RunOrDie(leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: 60 * time.Second,
		RenewDeadline: 30 * time.Second,
		RetryPeriod:   20 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: run,
			OnStoppedLeading: func() {
				utilruntime.HandleError(fmt.Errorf("Lost master"))
			},
		},
	})

	glog.Fatalln("Lost lease")
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&hostname, "hostname", "", "Hostname to distinguish different replicas.")
}

func createRecorder(kubeClient *kubernetes.Clientset, comp string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: v1core.New(kubeClient.CoreV1().RESTClient()).Events(getNamespace())})
	// https://github.com/kubernetes/client-go/issues/255#issuecomment-318214361
	return eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: comp})
}

func getHostname() string {
	if hostname != "" {
		return hostname
	}
	hostname, err := os.Hostname()
	if err != nil {
		glog.Fatalf("Unable to get hostname: %v", err)
	}
	return hostname
}

func getNamespace() string {
	if ns := os.Getenv("NAMESPACE"); ns != "" {
		return ns
	}
	return "default" // or kube-system
}
