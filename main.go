package main

import (
	"github.com/HeGaoYuan/kube-events-maker/cmd"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	cmd.Run()
}
