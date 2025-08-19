#!/bin/sh
cat <<EOF > database.properties
database.hostname = ${DB_HOST}
database.port = ${DB_PORT}
database.username = ${DB_USER}
database.password = ${DB_PASS}
EOF

# Wait for MySQL to be ready
echo "Waiting for MySQL at ${DB_HOST}:${DB_PORT}..."
until nc -z ${DB_HOST} ${DB_PORT}; do
  sleep 2
done
echo "MySQL is up!"

exec "$@"