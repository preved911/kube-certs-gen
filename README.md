YAML configuration file example:
```yaml
etcd:
  servers:
    - name: first-etcd
      certSANs: 
        - etcd-server-dns-name
      certIPs:
        - etcd-server-ip-address
  peers:
    - name: first-etcd
      certSANs:
        - etcd-peer-dns-name
      certIPs:
        - etcd-peer-ip-address
api:
  servers:
    - name: first-api
      certSANs:
        - kube-apiserver-dns-name
      certIPs:
        - kube-apiserver-ip-address
clusterConfiguration:
  certificatesDir: /home/q/Downloads/exercism/certs
```
