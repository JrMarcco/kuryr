#!/bin/bash
# 在 scripts/ 目录下创建 gen_etcd_ed25519_certs.sh
set -e

echo "=== 生成 etcd Ed25519 TLS 证书 ==="

# 创建证书目录
CERT_DIR="./kuryr-certs"
mkdir -p $CERT_DIR
cd $CERT_DIR

# 清理旧证书
rm -f *.pem *.csr *.srl

echo "1. 生成 CA 私钥 (Ed25519)"
# 生成 Ed25519 CA 私钥
openssl genpkey -algorithm Ed25519 -out ca-key.pem

echo "2. 生成 CA 证书"
# 生成 CA 证书
openssl req -new -x509 -key ca-key.pem -days 365 -out ca.pem \
    -subj "/C=CN/ST=Beijing/L=Beijing/O=Kuryr/OU=CA/CN=etcd-ca"

echo "3. 生成 etcd 服务器私钥 (Ed25519)"
# 生成 etcd 服务器 Ed25519 私钥
openssl genpkey -algorithm Ed25519 -out server-key.pem

echo "4. 创建服务器证书配置文件"
# 创建服务器证书配置文件，支持多个 SAN
cat > server.conf <<EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = CN
ST = Beijing
L = Beijing
O = Kuryr
OU = Server
CN = etcd-server

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = etcd-kuryr
DNS.3 = etcd-server
IP.1 = 127.0.0.1
IP.2 = 192.168.3.3
IP.3 = ::1
EOF

echo "5. 生成 etcd 服务器证书请求"
# 生成 etcd 服务器证书请求
openssl req -new -key server-key.pem -out server.csr -config server.conf

echo "6. 签发 etcd 服务器证书"
# 签发 etcd 服务器证书
openssl x509 -req -in server.csr -CA ca.pem -CAkey ca-key.pem \
    -CAcreateserial -out server.pem -days 365 \
    -extensions v3_req -extfile server.conf

echo "7. 生成 etcd 客户端私钥 (Ed25519)"
# 生成 etcd 客户端 Ed25519 私钥
openssl genpkey -algorithm Ed25519 -out client-key.pem

echo "8. 创建客户端证书配置文件"
# 创建客户端证书配置文件
cat > client.conf <<EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = CN
ST = Beijing
L = Beijing
O = Kuryr
OU = Client
CN = etcd-client

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth
EOF

echo "9. 生成 etcd 客户端证书请求"
# 生成 etcd 客户端证书请求
openssl req -new -key client-key.pem -out client.csr -config client.conf

echo "10. 签发 etcd 客户端证书"
# 签发 etcd 客户端证书
openssl x509 -req -in client.csr -CA ca.pem -CAkey ca-key.pem \
    -CAcreateserial -out client.pem -days 365 \
    -extensions v3_req -extfile client.conf

echo "11. 设置证书文件权限"
# 设置适当的文件权限
chmod 600 *-key.pem
chmod 644 *.pem
chmod 644 *.conf

echo "12. 清理临时文件"
# 清理临时文件
rm -f *.csr *.srl

echo "=== 证书验证 ==="
echo "CA 证书信息："
openssl x509 -in ca.pem -text -noout | grep -E "(Subject:|Issuer:|Public Key Algorithm|Signature Algorithm)"

echo -e "\n服务器证书信息："
openssl x509 -in server.pem -text -noout | grep -E "(Subject:|Issuer:|Public Key Algorithm|Signature Algorithm|DNS:|IP Address:)"

echo -e "\n客户端证书信息："
openssl x509 -in client.pem -text -noout | grep -E "(Subject:|Issuer:|Public Key Algorithm|Signature Algorithm)"

echo -e "\n=== 证书文件列表 ==="
ls -la *.pem *.conf

echo -e "\n✅ Ed25519 TLS 证书生成成功！"
echo "证书位置: $(pwd)"
echo -e "\n证书文件说明："
echo "  ca.pem         - CA 根证书"
echo "  ca-key.pem     - CA 私钥"
echo "  server.pem     - etcd 服务器证书"
echo "  server-key.pem - etcd 服务器私钥"
echo "  client.pem     - etcd 客户端证书"
echo "  client-key.pem - etcd 客户端私钥"
echo "  server.conf    - 服务器证书配置文件"
echo "  client.conf    - 客户端证书配置文件"
