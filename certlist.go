package main

import (
	certutil "k8s.io/client-go/util/cert"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
)

// this functions copied from
// "k8s.io/kubernetes/cmd/kubeadm/app/phases/certs"
// for using in main without overrides.
type configMutatorsFunc func(*kubeadmapi.InitConfiguration, *certutil.Config) error

func makeAltNamesMutator(f func(*kubeadmapi.InitConfiguration) (*certutil.AltNames, error)) configMutatorsFunc {
	return func(mc *kubeadmapi.InitConfiguration, cc *certutil.Config) error {
		altNames, err := f(mc)
		if err != nil {
			return err
		}
		cc.AltNames = *altNames
		return nil
	}
}
