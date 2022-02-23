# kube-events-maker
**kube-events-maker** makes the custom Event through wathing the Pod's state transition. It helps you to know the reason why your Pod terminated.

# Motivation
When a Pod failed, we always use the command `kubectl describe ${PodName}`, it will show the Events of the pod, which helps you to know the Pod's history state.
But in practice, we always see the below Event. It tells us nothing except the Pod failed and was restaring. We don't know why the Pod is failed.
```
Warning  BackOff  3m6s (x3808 over 17h)  kubelet   Back-off restarting failed container
```

Actually, the Pod's status field has some information about the Pod, but it is transient. So my motivation is wathing the Pod's status filed, and make the custom Event to tell users why the Pod terminated. 

# Effect
The custom Event likes below, it tell us the Pod terminated because of OOMKilled. **The code is simple, you are free to see how it achieves： https://github.com/HeGaoYuan/kube-events-maker/blob/master/pkg/kube/watch_pod.go.**

```
Warning  ContainerTerminatedOOMKilled  51s   kube-events-maker   demo container exits(137) for OOMKilled reason
```
# Usage
Deploy the yaml files under deploy folder.

# Build
`docker build -t ${Your_Image_name} -f Dockerfile `

