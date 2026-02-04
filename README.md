# job-worker

Level1:


Quick Test commands for your CLI:
Start a long job: go run cmd/cli/main.go start sleep 10

Check Status: go run cmd/cli/main.go status <ID>

Get Output: go run cmd/cli/main.go logs <ID>

Kill Job: go run cmd/cli/main.go stop <ID>


Certificates
To run this, you'll need a self-signed certificate. You can generate one quickly with: 
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes

Level 2:

Step-by-Step Certificate Generation

You can't test this without valid certs. Run these commands:

Create CA: openssl req -x509 -newkey rsa:4096 -keyout ca.key -out ca.crt -days 365 -nodes -subj "/CN=MyJobWorkerCA"

Create Server Cert: 
openssl req -newkey rsa:4096 -keyout server.key -out server.csr -nodes -subj "/CN=localhost" 
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365

Create Client Cert: openssl req -newkey rsa:4096 -keyout client.key -out client.csr -nodes -subj "/CN=admin-client" openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365