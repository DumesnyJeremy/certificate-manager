#!/bin/bash

BINARY_FILE="certificate-manager"
SERVICE_PATH="/etc/systemd/system"
CONF_DIR_NAME="certificate-manager"
BINARY_PATH="/usr/local/bin"
LOG_PATH="/var/log"

function usage() {
  echo "Usage: install: $0 -i"
  echo "Usage: daemon: $0 -d"
  echo "Usage: timer: $0 -t"
  echo "Usage: uninstall: $0 -u"
  echo "Usage: purge: $0 -p"
}

function install() {

  # Copy the binary
  if [ -f "$BINARY_FILE" ]; then
    cp ${BINARY_FILE} ${BINARY_PATH}
    echo "* ${BINARY_FILE} copied to /usr/local/bin ."
  else
    echo "${BINARY_FILE} Not found."
    exit
  fi

  # Create default configuration directory
  mkdir -p /etc/${CONF_DIR_NAME}/letsencrypt/{certificates,account}
  chmod -R 640 /etc/${CONF_DIR_NAME}/letsencrypt/{certificates,account}
  echo "* Certificates and Account directory created in /etc/${CONF_DIR_NAME} ."

  # Copy the 3 config..sample to /etc/${CONF_DIR_NAME}
  array=(.json .toml .yaml)
  for i in "${array[@]}"; do
    cp configuration-files/config"$i".sample /etc/${CONF_DIR_NAME}/
    echo "* config$i.sample copied."
  done
  echo "== Installations Done =="
}

function daemon() {
  cat <<EOF >${SERVICE_PATH}/${BINARY_FILE}.service
[Unit]
Description=Verify certificate expiration and renew them
Wants=network.target
After=syslog.target network-online.target
[Service]
Environment="HOME=/root"
Type=simple
ExecStart=/usr/local/bin/${BINARY_FILE} -d
#User=deepak
#Group=admin
Restart=on-failure
RestartSec=10
KillMode=process
[Install]
WantedBy=multi-user.target
EOF
  chmod 640 ${SERVICE_PATH}/${BINARY_FILE}.service
  systemctl daemon-reload
  echo "* ${BINARY_FILE}.service copied in ${SERVICE_PATH} ."
  systemctl enable ${BINARY_FILE}.service
  echo "* Systemctl enable."
  mkdir "${LOG_PATH}/${BINARY_FILE}"
  touch "${LOG_PATH}/${BINARY_FILE}/${BINARY_FILE}.log"
  chmod 666 "${LOG_PATH}/${BINARY_FILE}/${BINARY_FILE}.log"
  echo "== Daemon created and launched =="
}

function timer() {
  cat <<EOF >${SERVICE_PATH}/${BINARY_FILE}.timer
[Unit]
Description=Run certificate-manager daily
Requires=${BINARY_FILE}.service

[Timer]
OnCalendar=*-*-* 1:0:0
AccuracySec=1s

[Install]
WantedBy=timer.target
EOF

  cat <<EOF >${SERVICE_PATH}/${BINARY_FILE}.service
[Unit]
Description=Verify certificate expiration and renew them
Wants=network.target
After=syslog.target network-online.target
[Service]
Environment="HOME=/root"
Type=standalone
ExecStart=/usr/local/bin/${BINARY_FILE}
StandardOutput=${LOG_PATH}/${BINARY_FILE}/${BINARY_FILE}.log
#User=deepak
#Group=admin
Restart=on-failure
RestartSec=10
KillMode=process
[Install]
WantedBy=multi-user.target
EOF
  mkdir "${LOG_PATH}/${BINARY_FILE}"
  touch "${LOG_PATH}/${BINARY_FILE}/${BINARY_FILE}.log"
  chmod 666 "${LOG_PATH}/${BINARY_FILE}/${BINARY_FILE}.log"
}

function uninstall() {
  local BINARY_FILE_PATH=${BINARY_PATH}/${BINARY_FILE}

  # Delete the ${BINARY_FILE}
  if [ -f "${BINARY_FILE_PATH}" ]; then
    rm -f ${BINARY_FILE_PATH}
    echo "* ${BINARY_FILE} deleted from /usr/local/bin/ ."
  else
    echo "${BINARY_FILE_PATH} Not found."
  fi

  # Stop, Disable and Delete ${BINARY_FILE}.service
  if [ -f "${SERVICE_PATH}/${BINARY_FILE}.service" ]; then
    systemctl stop ${BINARY_FILE}.service
    systemctl stop ${BINARY_FILE}.timer
    systemctl disable ${BINARY_FILE}.service
    rm -f ${SERVICE_PATH}/${BINARY_FILE}.service
    rm -f ${SERVICE_PATH}/${BINARY_FILE}.timer
    rm -rf ${LOG_PATH}/${BINARY_FILE}/
    systemctl daemon-reload
    echo "* ${BINARY_FILE}.service deleted from ${SERVICE_PATH} ."
  else
    echo "${BINARY_FILE}.service Not found."
  fi
}

function purge() {
  if [ -d "/etc/${CONF_DIR_NAME}" ]; then
    rm -rf /etc/${CONF_DIR_NAME}
    echo "* ${CONF_DIR_NAME} deleted from /etc/ ."
  else
    echo "${CONF_DIR_NAME} Not found."
  fi

  echo "== Deletions Done =="
}

function is_user_root() { [ "$(id -u)" -eq 0 ]; }

if is_user_root; then
  while getopts "idtup" OPT; do
    case "${OPT}" in
    i)
      install
      ;;
    d)
      daemon
      ;;
    t)
      timer
      ;;
    u)
      uninstall
      ;;
    p)
      purge
      ;;
    \?)
      usage
      ;;
    esac
  done
else
  echo "Please run as root."
  exit 1
fi
