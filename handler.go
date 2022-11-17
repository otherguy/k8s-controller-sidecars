package main

import (
	"strings"
	"time"

	"github.com/avast/retry-go"

	log "github.com/Sirupsen/logrus"
	core_v1 "k8s.io/api/core/v1"

	"k8s.io/client-go/tools/clientcmd"

	set "github.com/deckarep/golang-set"
)

// Handler interface contains the methods that are required
type Handler interface {
	Init() error
	ObjectCreated(obj interface{})
	ObjectDeleted(obj interface{})
	ObjectUpdated(objOld, objNew interface{})
}

// SidecarShutdownHandler is a sample implementation of Handler
type SidecarShutdownHandler struct{}

// Init handles any handler initialization
func (t *SidecarShutdownHandler) Init() error {
	log.Info("SidecarShutdownHandler.Init")
	return nil
}

// Send a shutdown signal to all containers in the Pod
func sendShutdownSignal(pod *core_v1.Pod, containers set.Set) {
	log.Infof("It's going down, I'm yelling TIMBERRRRR for pod: %s", pod.Name)

	// Multiple arguments must be provided as separate "command" parameters
	// The first one is added automatically.
	// Todo: Update requestFromConfig to handle this better
	command := "sh&command=-c&command=kill+-s+TERM+1" // "kill -s TERM 1"
	// creates the connection
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		log.Fatal(err)
	}

	// Create a round tripper with all necessary kubernetes security details
	wrappedRoundTripper, err := roundTripperFromConfig(config)
	if err != nil {
		log.Fatalln(err)
	}

	for _, c := range containers.ToSlice() {
		// Create a request out of config and the query parameters
		req, err := requestFromConfig(config, pod.Name, c.(string), pod.Namespace, command)
		if err != nil {
			log.Infoln(err)
		}

		err = retry.Do(
			func() error {
				// Send the request and let the callback do its work
				_, err = wrappedRoundTripper.RoundTrip(req)
				if err != nil {
					return err
				}
				return nil
			},
			retry.Delay(3*time.Second),
			retry.Attempts(5),
		)
		if err != nil {
			log.Errorln(err)
		}
	}
}

// ObjectCreated is called when an object is created
func (t *SidecarShutdownHandler) ObjectCreated(obj interface{}) {
	log.Debug("SidecarShutdownHandler.ObjectCreated")
	// assert the type to a Pod object to pull out relevant data
	pod := obj.(*core_v1.Pod)

	sidecarsString, exists := pod.Annotations["otherguy.io/sidecars"]

	if exists {
		log.Debugf("    ResourceTrackable: true")
		log.Infof("    Sidecars: %s, node: %s, phase: %s", sidecarsString, pod.Spec.NodeName, pod.Status.Phase)
	} else {
		return
	}

	sidecars := set.NewSet()

	for _, s := range strings.Split(sidecarsString, ",") {
		sidecars.Add(s)
	}

	allContainers := set.NewSet()
	runningContainers := set.NewSet()
	completedContainers := set.NewSet()

	for _, containerStatus := range pod.Status.ContainerStatuses {
		allContainers.Add(containerStatus.Name)

		if containerStatus.Ready {
			runningContainers.Add(containerStatus.Name)
		} else {
			terminated := containerStatus.State.Terminated
			if terminated != nil && (terminated.Reason == "Completed" || terminated.Reason == "Error") {
				completedContainers.Add(containerStatus.Name)
			}
		}
	}

	log.Debugf("    all       : %s", allContainers)
	log.Debugf("    running   : %s", runningContainers)
	log.Debugf("    completed : %s", completedContainers)
	log.Debugf("    sidecars  : %s", sidecars)

	// If we have accounted for all of the containers, and the sidecar containers are the only
	// ones still running, issue them each a shutdown command
	if runningContainers.Union(completedContainers).Equal(allContainers) {
		log.Debugf("  We have all the containers")
		if runningContainers.Equal(sidecars) {
			log.Infof("    Sending shutdown signal to containers: %s, %s", pod.Name, sidecars)
			sendShutdownSignal(pod, sidecars)
		}
	}
}

// ObjectDeleted is called when an object is deleted
func (t *SidecarShutdownHandler) ObjectDeleted(obj interface{}) {
	log.Debug("SidecarShutdownHandler.ObjectDeleted")
}

// ObjectUpdated is called when an object is updated.
// Note that the controller in this repo will never call this function properly.
// It uses only ObjectCreated
func (t *SidecarShutdownHandler) ObjectUpdated(objOld, objNew interface{}) {
	log.Debug("SidecarShutdownHandler.ObjectUpdated")
}
