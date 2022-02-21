package cmd

import (
	"context"
	"flag"
	"github.com/HeGaoYuan/kube-events-maker/pkg/kube"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	kubeclientset "k8s.io/client-go/kubernetes"
	restclientset "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
	"os/signal"
	"syscall"
)

var (
	conf = flag.String("conf", "config.yaml", "The config path file")
)

func Run() {
	flag.Parse()
	b, err := ioutil.ReadFile(*conf)
	if err != nil {
		klog.Fatalf("Error reading %s", *conf)
	}
	b = []byte(os.ExpandEnv(string(b)))

	var cfg Config
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		klog.Fatal("Error parsing kubeConfig to YAML")
	}

	namespace := cfg.Namespace
	if namespace == corev1.NamespaceAll {
		klog.Info("Using cluster scoped operator")
	} else {
		klog.Infof("Scoping operator to namespace %s", namespace)
	}

	kubeConfig, err := clientcmd.BuildConfigFromFlags(cfg.MasterURL, cfg.KubeConfig)
	if err != nil {
		klog.Fatalf("Error building kubeConfig: %s", err.Error())
	}

	clientSet, err := kubeclientset.NewForConfig(restclientset.AddUserAgent(kubeConfig, kube.EventSourceName))
	if err != nil {
		klog.Fatalf("Error creating kubeClientSet: %s", err.Error())
	}
	kube.InitEventRecorder(clientSet)
	podWatcher := kube.NewPodWatcher(clientSet, namespace)

	ctx, cancel := context.WithCancel(context.Background())
	leaderLost := make(chan bool)
	if cfg.LeaderElection.Enabled {
		l, err := kube.NewLeaderElector(cfg.LeaderElection.LeaderElectionID, kubeConfig,
			func(_ context.Context) {
				klog.Info("leader election got")
				podWatcher.Start()
			},
			func() {
				klog.Error("leader election lost")
				leaderLost <- true
			},
		)
		if err != nil {
			klog.Fatal("create leader elector failed")
		}
		go l.Run(ctx)
	} else {
		podWatcher.Start()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	gracefulExit := func() {
		defer close(c)
		defer close(leaderLost)
		cancel()
		podWatcher.Stop()
		klog.Info("Exiting")
	}

	select {
	case sig := <-c:
		klog.Infof("Received signal() to exit", sig.String())
		gracefulExit()
	case <-leaderLost:
		klog.Warning("Leader election lost")
		gracefulExit()
	}
}
