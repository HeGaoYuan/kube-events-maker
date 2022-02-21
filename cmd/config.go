package cmd

import "github.com/HeGaoYuan/kube-events-maker/pkg/kube"

type Config struct {
	KubeConfig     string                    `yaml:"kubeconfig"`
	MasterURL      string                    `yaml:"masterURL"`
	Namespace      string                    `yaml:"namespace"`
	LeaderElection kube.LeaderElectionConfig `yaml:"leaderElection"`
}
