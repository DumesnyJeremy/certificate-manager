#!/bin/bash

BINARY_FILE="certificate-manager"
SERVICE_PATH="/etc/systemd/system"
CONF_DIR_NAME="certificate-manager"
BINARY_PATH="/usr/local/bin"
LOG_PATH="/var/log"

function usage() {
  echo "Flags:"
  echo "  Install:   $0 -i   Copy the binary in ${BINARY_PATH} and create the configuration file in /etc/${CONF_DIR_NAME}."
  echo "  Uninstall: $0 -u   Delete the copy of the binary, and if the Daemon or the Timer are setup, it will disable them."
  echo "  Purge:     $0 -p   Delete the configuration file."
  echo "  Daemon:    $0 -d   Create the .service file in ${SERVICE_PATH}/."
  echo "  Timer:     $0 -t   Create the .timer and the .service in ${SERVICE_PATH}/."
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
StandardOutput=${LOG_PATH}/${BINARY_FILE}/${BINARY_FILE}.log
#User=deepak
#Group=admin
Restart=on-failure
RestartSec=10
KillMode=process
[Install]
WantedBy=multi-user.target
EOF
  chmod 640 ${SERVICE_PATH}/${BINARY_FILE}.service
  echo "* ${BINARY_FILE}.service created in ${SERVICE_PATH} ."
  mkdir "${LOG_PATH}/${BINARY_FILE}"
  touch "${LOG_PATH}/${BINARY_FILE}/${BINARY_FILE}.log"
  chmod 666 "${LOG_PATH}/${BINARY_FILE}/${BINARY_FILE}.log"
  echo "* Log File created."
  echo "== Daemon created and launched =="
}

function timer() {
  cat <<EOF >${SERVICE_PATH}/${BINARY_FILE}.timer
[Unit]
Description=Run certificate-manager daily
Requires=${BINARY_FILE}.service

[Timer]
OnCalendar=daily
Persistant = vrai

[Install]
WantedBy=timer.target
EOF
echo "* ${BINARY_FILE}.timer created in ${SERVICE_PATH} ."
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
  echo "* ${BINARY_FILE}.timer created in ${SERVICE_PATH} ."
  mkdir "${LOG_PATH}/${BINARY_FILE}"
  touch "${LOG_PATH}/${BINARY_FILE}/${BINARY_FILE}.log"
  chmod 666 "${LOG_PATH}/${BINARY_FILE}/${BINARY_FILE}.log"
  echo "* Log File created."
  echo "== Timer created and launched =="
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


  while getopts "idtuph" OPT; do
    case "${OPT}" in
    i)
      if is_user_root; then
      install
      else
  echo "Please run as root."
  exit 1
  fi
  ;;
    d)
       if is_user_root; then
      daemon
      else
  echo "Please run as root."
  exit 1
  fi
      ;;
    t)
       if is_user_root; then
      timer
      else
  echo "Please run as root."
  exit 1
  fi
      ;;
    u)
       if is_user_root; then
      uninstall
      else
  echo "Please run as root."
  exit 1
  fi
      ;;
    p)
       if is_user_root; then
      purge
      else
  echo "Please run as root."
  exit 1
  fi
      ;;
    h)
      usage
      ;;
    esac
  done
