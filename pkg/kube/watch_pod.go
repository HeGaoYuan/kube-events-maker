package kube

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

type WatchPod struct {
	informer        cache.SharedInformer
	stopper         chan struct{}
	eventRecorder   record.EventRecorder
}

func NewPodWatcher(clientSet kubeclientset.Interface, namespace string) *WatchPod {
	factory := informers.NewSharedInformerFactoryWithOptions(clientSet, 0, informers.WithNamespace(namespace))
	informer := factory.Core().V1().Pods().Informer()

	watcher := &WatchPod{
		informer:        informer,
		stopper:         make(chan struct{}),
	}

	informer.AddEventHandler(watcher)
	return watcher
}

func (w *WatchPod) Start() {
	go w.informer.Run(w.stopper)
}

func (w *WatchPod) Stop() {
	w.stopper <- struct{}{}
	close(w.stopper)
}

func (w *WatchPod) OnAdd(obj interface{}) {
	pod := obj.(*corev1.Pod)
	processPod(pod)
}

func (w *WatchPod) OnUpdate(oldObj, newObj interface{}) {
	pod := newObj.(*corev1.Pod)
	processPod(pod)
}

func (w *WatchPod) OnDelete(obj interface{}) {
	// do nothing
}

func processPod(pod *corev1.Pod) {
	// 对pod中的每个containerStatus进行处理
	deleteGraceSeconds := pod.DeletionGracePeriodSeconds
	for _, containerStatus := range pod.Status.ContainerStatuses {
		// 如果当前的containerStatus为Terminated, 且deleteGraceSeconds为nil，那么就是普通的pod由于正确或者错误退出
		if containerStatus.State.Terminated != nil && deleteGraceSeconds == nil {
			constructEvent(pod, containerStatus.Name, containerStatus.State.Terminated)
		}
		// 如果当前的containerStatus为Terminated, 且deleteGraceSeconds不为nil，且值为0（因为值还可能为30），那么就是delete的terminated
		if containerStatus.State.Terminated != nil && deleteGraceSeconds != nil && *deleteGraceSeconds == 0 {
			constructEvent(pod, containerStatus.Name, containerStatus.State.Terminated)
		}
	}
}

func constructEvent(pod *corev1.Pod, containerName string, containerStateTerminated *corev1.ContainerStateTerminated) {
	eventType := "Warning"
	if containerStateTerminated.Reason == "Completed" && containerStateTerminated.ExitCode == 0 {
		eventType = "Normal"
	}

	terminateReason := "Empty"
	if containerStateTerminated.Reason != "" {
		terminateReason = containerStateTerminated.Reason
	}
	eventReason := fmt.Sprintf("ContainerTerminated%s", terminateReason)

	eventMessage := ""
	if containerStateTerminated.Message == "" {
		eventMessage = fmt.Sprintf("%s container exits(%d) for %s reason",
			containerName, containerStateTerminated.ExitCode, terminateReason)
	} else {
		containerStateTerminatedMessage := ""
		if len(containerStateTerminated.Message) > 1000 {
			containerStateTerminatedMessage = containerStateTerminated.Message[:1000]
		} else {
			containerStateTerminatedMessage = containerStateTerminated.Message
		}
		eventMessage = fmt.Sprintf("%s container exits(%d) for %s reason with the %s",
			containerName, containerStateTerminated.ExitCode, terminateReason, containerStateTerminatedMessage)
	}

	eventRecorder.Event(pod, eventType, eventReason, eventMessage)
}
