#!/usr/bin/env bash

# This file will be run on droplets.

# Enable firewall if it is not active
# and open networking ports.
ufw status | grep -qw active
status=$?

# 1 indicates inactive firewall.
if [[ $status -gt 0 ]]
then
    ufw --force enable
fi

sudo ufw allow 80
sudo ufw allow 443

apt update && apt install -y letsencrypt

read -p "Domain name for letsencript to generate TLS certs and keys: " dn
sudo letsencrypt certonly --standalone -d $dn -n --agree-tos --email zicodeng@gmail.com

