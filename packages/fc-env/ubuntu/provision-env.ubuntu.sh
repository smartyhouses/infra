#! /bin/bash

# This script is supposed to be executed in a running *Ubuntu* container.
# The container is then extracted to a rootfs image for the Firecracker VM.

set -euo pipefail

# Set up autologin.
mkdir /etc/systemd/system/serial-getty@ttyS0.service.d
cat <<EOF > /etc/systemd/system/serial-getty@ttyS0.service.d/autologin.conf
[Service]
ExecStart=
ExecStart=-/sbin/agetty --noissue --autologin root %I 115200,38400,9600 vt102
EOF

# --- Enable systemd services --- #
# Because this script runs in a container we can't use `systemctl`.
# Containers don't run init daemons. We have to enable the runner service manually.
mkdir -p /etc/systemd/system/multi-user.target.wants
ln -s /etc/systemd/system/devbookd.service /etc/systemd/system/multi-user.target.wants/devbookd.service
# ------------------------------- #

echo "export SHELL='/bin/bash'" > /etc/profile.d/shell.sh
echo "export PS1='\w \$ '" > /etc/profile.d/prompt.sh
echo "export PS1='\w \$ '" >> "/etc/profile"
echo "export PS1='\w \$ '" >> "/root/.bashrc"

mkdir -p /etc/ssh
touch /etc/ssh/sshd_config
echo "PermitRootLogin yes" >> /etc/ssh/sshd_config
echo "PermitEmptyPasswords yes" >> /etc/ssh/sshd_config
echo "PasswordAuthentication yes" >> /etc/ssh/sshd_config

# Remove password for root.
passwd -d root

# Add DNS.
echo "nameserver 8.8.8.8" > /etc/resolv.conf

# Start systemd services
systemctl enable devbookd
systemctl enable chrony

# Delete itself once done.
rm -- "$0"
