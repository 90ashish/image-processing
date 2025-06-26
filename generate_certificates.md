## Run these commands in your project root (requires OpenSSL):

# 1. Create a CA key & cert
openssl genrsa -out ca.key 4096
openssl req -x509 -new -nodes -key ca.key \
  -subj "/CN=MyCA" -days 3650 -out ca.crt

# 2. Create server key & CSR, then have CA sign it
openssl genrsa -out server.key 4096
openssl req -new -key server.key \
  -subj "/CN=localhost" -out server.csr
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key \
  -CAcreateserial -out server.crt -days 3650

# 3. (Optional) Create client key & CSR, CA-sign it
openssl genrsa -out client.key 4096
openssl req -new -key client.key \
  -subj "/CN=client1" -out client.csr
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key \
  -CAcreateserial -out client.crt -days 3650
