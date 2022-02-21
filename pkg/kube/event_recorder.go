package kube

import (
	corev1 "k8s.io/api/core/v1"
	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
)

var eventRecorder record.EventRecorder

const EventSourceName = "kube-events-maker"

func InitEventRecorder(clientSet kubeclientset.Interface) {
	eventBroadcaster := record.NewBroadcasterWithCorrelatorOptions(record.CorrelatorOptions{
		BurstSize: 25,
		QPS: 1. / 240.,
	})
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: clientSet.CoreV1().Events("")})
	eventRecorder = eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: EventSourceName})
}
