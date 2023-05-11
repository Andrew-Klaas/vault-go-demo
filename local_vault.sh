#!/bin/bash
export VAULT_ADDR=http://127.0.0.1:8200

vault login root

cat << EOF > transit-app-example.policy
path "*" {
    capabilities = ["read", "list", "create", "update", "delete"]
}
path "transit/*" {
    capabilities = ["read", "list", "create", "update", "delete"]
}
EOF
vault policy write transit-app-example transit-app-example.policy

vault secrets enable database

vault write database/config/my-postgresql-database \
    plugin_name=postgresql-database-plugin \
    allowed_roles="my-role, vault_go_demo" \
    connection_url="postgresql://{{username}}:{{password}}@127.0.0.1:5432/vault_go_demo?sslmode=disable" \
    username="vault" \
    password="MySecretPassW0rd"

vault write database/roles/vault_go_demo \
    db_name=my-postgresql-database \
    creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
    ALTER USER \"{{name}}\" WITH SUPERUSER;" \
    default_ttl="1h" \
    max_ttl="24h"

vault read database/creds/vault_go_demo


vault secrets enable transit
vault write -f transit/keys/my-key

#Set your google oauth2 app client_id and client_secret as env variables
vault kv put secret/oauth2/config \
    client_id=$CLIENT_ID \
    client_secret=$CLIENT_SECRET


exit 0


kubectl apply -f go_vault_demo/

#Local postgres setup
psql postgres
CREATE USER vault WITH PASSWORD 'MySecretPassW0rd';
ALTER USER vault WITH SUPERUSER;
CREATE DATABASE vault_go_demo;