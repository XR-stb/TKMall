#!/bin/bash

ETCD_VER=v3.5.17

# choose either URL
GOOGLE_URL=https://storage.googleapis.com/etcd
GITHUB_URL=https://github.com/etcd-io/etcd/releases/download
DOWNLOAD_URL=${GOOGLE_URL}

# Define the installation directory
INSTALL_DIR=$(dirname "$0")/bin

# Create the installation directory if it doesn't exist
mkdir -p ${INSTALL_DIR}

# Remove any previous downloads
rm -f /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
rm -rf /tmp/etcd-download-test && mkdir -p /tmp/etcd-download-test

# Download etcd
curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz

# Extract etcd
tar xzvf /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz -C /tmp/etcd-download-test --strip-components=1

# Remove the downloaded tar file
rm -f /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz

# Move etcd binaries to the installation directory
mv /tmp/etcd-download-test/etcd ${INSTALL_DIR}/
mv /tmp/etcd-download-test/etcdctl ${INSTALL_DIR}/
mv /tmp/etcd-download-test/etcdutl ${INSTALL_DIR}/

# Clean up
rm -rf /tmp/etcd-download-test

# Verify the installation
${INSTALL_DIR}/etcd --version
${INSTALL_DIR}/etcdctl version
${INSTALL_DIR}/etcdutl version

# Start etcd
${INSTALL_DIR}/etcd --data-dir tools/etcd/bin
