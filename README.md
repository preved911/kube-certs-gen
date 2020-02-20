[![Build Status](https://travis-ci.com/preved911/kube-certs-gen.svg?branch=master)](https://travis-ci.com/preved911/kube-certs-gen)

YAML configuration file example:
```yaml
etcd:
  servers:
    - name: first-etcd
      certs:
        SANs: 
          - etcd-server-dns-name
        IPs:
          - etcd-server-ip-address
  peers:
    - name: first-etcd
      certs:
        SANs:
          - etcd-peer-dns-name
        IPs:
          - etcd-peer-ip-address
api:
  servers:
    - name: first-api
      certs:
        SANs:
          - kube-apiserver-dns-name
        IPs:
          - kube-apiserver-ip-address
nodes:
  - "kube-node-0"
  - "kube-node-1"
clusterConfiguration:
  certificatesDir: /home/q/Downloads/exercism/certs
```
