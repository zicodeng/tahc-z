#!/usr/bin/env bash

# Create droplets via doctl.

set -e

function createDroplet() {
    echo "Creating your new droplet..."
    doctl compute droplet create $1 \
    --image $2 \
    --size $3 \
    --region $4 \
    --ssh-keys a7:29:e6:56:6b:78:88:e8:43:14:1c:9c:1f:46:87:79 \
    --wait
}

function associateDomain() {
    echo "Please choose one of the droplet IP addresses below to associate a (sub) domain name:"
    doctl compute droplet list

    read -p "Droplet IP address: " addr
    read -p "Domain name: " dn

    doctl compute domain create $dn --ip-address $addr
}

function provisionDroplet() {
    echo "Please choose one of the droplet IP addresses below to provision:"
    doctl compute droplet list

    read -p "Droplet IP address: " addr

    ssh -oStrictHostKeyChecking=no root@$addr 'bash -s' < provision.sh
}

read -p "Droplet name: " name

echo "Please choose one of the available images below for your new droplet:"
doctl compute image list --public
read -p "Droplet image (use the string in the \"Slug\" column): " image

echo "Please choose one of the available sizes below for your new droplet:"
doctl compute size list
read -p "Droplet size (use the string in the \"Slug\" column): " size

echo "Please choose one of the available data centers below for your new droplet:"
doctl compute region list
read -p "Droplet region (use the string in the \"slug\" column): " region

createDroplet $name $image $size $region

continue=true
while $continue
do
    read -p "Would like to create another droplet with the same configuration (y/n)? " reply
    if [[ "$reply" =~ ^[Yy]$ ]]
    then
        continue=true

        read -p "Droplet name: " name
        createDroplet $name $image $size $region

    elif [[ "$reply" =~ ^[Nn]$ ]]
    then
        continue=false

    else
        continue=true
    fi
done

associateDomain

continue=true
while $continue
do
    read -p "Would like to associate another (sub) domain name (y/n)? " reply
    if [[ "$reply" =~ ^[Yy]$ ]]
    then
        continue=true
        associateDomain

    elif [[ "$reply" =~ ^[Nn]$ ]]
    then
        continue=false

    else
        continue=true
    fi
done

provisionDroplet

continue=true
while $continue
do
    read -p "Would like to provision another droplet? (y/n) " reply
    if [[ "$reply" =~ ^[Yy]$ ]]
    then
        continue=true
        provisionDroplet

    elif [[ "$reply" =~ ^[Nn]$ ]]
    then
        exit 1

    else
        continue=true
    fi
done