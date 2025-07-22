# Kubernetes 部署指南

本目录包含了将 MongoDB 分片集群和 etcd 集群部署到 Kubernetes 的配置文件。

## MongoDB 分片集群

MongoDB 部署包含：
- 3 个配置服务器（Config Server）副本集
- 2 个分片（Shard），每个分片有 2 个副本
- 2 个 Mongos 路由器
- 使用 cert-manager 自动管理 TLS 证书
- 所有连接和内部通信使用 TLS 加密

详细部署指南请参考：[mongo/README.md](mongo/README.md)

### 快速开始

```bash
cd mongo/

# 1. 安装 cert-manager（如果尚未安装）
./install-cert-manager.sh

# 2. 部署 MongoDB 集群
./deploy.sh

# 3. 提取客户端证书
./extract-client-certs.sh ./mongodb-certs

# 4. 连接测试
kubectl port-forward svc/mongos 27017:27017 -n mongodb
# 在另一个终端中测试
mongosh "mongodb://root:rootpassword@localhost:27017/?authSource=admin" \
  --tls \
  --tlsCertificateKeyFile ./mongodb-certs/client.pem \
  --tlsCAFile ./mongodb-certs/ca.crt
```

### 存储要求

- Config Server：每个节点 10Gi
- Shard：每个节点 20Gi

## etcd 集群

etcd 部署包含：
- 3 节点的 etcd 集群
- 使用 cert-manager 自动管理 TLS 证书
- 强制 TLS 认证（客户端和对等节点）
- 安全的 cipher suites 配置

详细部署指南请参考：[etcd/README.md](etcd/README.md)

### 快速开始

```bash
cd etcd/

# 1. 安装 cert-manager（如果尚未安装）
./install-cert-manager.sh

# 2. 部署 etcd 集群
./deploy.sh

# 3. 提取客户端证书
./extract-client-certs.sh ./etcd-certs

# 4. 连接测试
kubectl port-forward svc/etcd-client 2379:2379 -n etcd
# 在另一个终端中测试
export ETCDCTL_API=3
export ETCDCTL_ENDPOINTS=https://localhost:2379
export ETCDCTL_CACERT=./etcd-certs/ca.crt
export ETCDCTL_CERT=./etcd-certs/client.crt
export ETCDCTL_KEY=./etcd-certs/client.key
etcdctl --user=root:rootpassword member list
```

### 存储要求

- 每个 etcd 节点：10Gi

## 清理

删除 MongoDB 集群：
```bash
kubectl delete namespace mongodb
```

删除 etcd 集群：
```bash
kubectl delete namespace etcd
```

## 注意事项

1. **生产环境**：
   - 请使用强密码替换默认密码
   - 考虑使用 StorageClass 来动态分配持久卷
   - 根据实际需求调整资源请求和限制
   - 为 Pod 设置合适的反亲和性规则

2. **网络策略**：
   - 考虑添加 NetworkPolicy 来限制集群间的通信
   - 使用 LoadBalancer 或 Ingress 替代 NodePort

3. **监控和备份**：
   - 部署监控解决方案（如 Prometheus）
   - 定期备份数据
   - 设置日志收集和分析

4. **TLS 证书**：
   - etcd 使用 cert-manager 自动管理 TLS 证书
   - 证书会在过期前自动续期
   - 支持自定义 CA 或自签名证书 