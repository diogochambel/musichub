#!/bin/sh
set -e

sed -i "s|http://localhost:8080|${API_URL}|g" proxy.conf.json

exec npm start -- --host 0.0.0.0
