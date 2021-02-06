#!/bin/bash

# Script to install minicni on a Kubernetes host.
# - Expects the host CNI binary path to be mounted at /host/opt/cni/bin.
# - Expects the host CNI network config path to be mounted at /host/etc/cni/net.d.
# - Expects the desired CNI config in the CNI_NETWORK_CONFIG env variable.
# - Expects the desired node name in the NODE_NAME env variable.

# Ensure all variables are defined, and that the script fails when an error is hit.
set -eu

CNI_NET_DIR=${CNI_NET_DIR:-/host/etc/cni/net.d}
CNI_CONF_NAME=${CNI_CONF_NAME:-10-minicni.conf}
CNI_PLUGINS_DIR=${CNI_PLUGINS_DIR:-/host/opt/cni/bin}
CURRENT_CNI_PLUGINS_DIR=${CURRENT_CNI_PLUGINS_DIR:-/cni-plugins}

function exit_with_message() {
    echo "$1"
    exit 1
}

# Clean up any existing binaries/config/assets.
function cleanup() {
    echo "Cleaning up and exiting."
    if [ -e "${CNI_NET_DIR}/${CNI_CONF_NAME}" ]; then
        echo "Removing minicni config ${CNI_NET_DIR}/${CNI_CONF_NAME} from CNI configuration directory."
        rm "${CNI_NET_DIR}/${CNI_CONF_NAME}"
    fi
    if [ -e "${CNI_PLUGINS_DIR}/minicni" ]; then
        echo "Removing minicni binary ${CNI_PLUGINS_DIR}/minicni from CNI plugins directory."
        rm "${CNI_PLUGINS_DIR}/minicni"
    fi
    echo "Exiting."
}

# error handle
trap cleanup EXIT

##########################################################################################
# Copy the CNI plugins to the plugin directory
##########################################################################################
if [ ! -w "${CNI_PLUGINS_DIR}" ];
then
    exit_with_message "$CNI_PLUGINS_DIR is not writeable"
fi

for plugin in "${CURRENT_CNI_PLUGINS_DIR}"/*;
do
    cp "${plugin}" "${CNI_PLUGINS_DIR}/" || exit_with_message "Failed to copy ${plugin} to ${CNI_PLUGINS_DIR}."
done

echo "COPY all CNI plugins from ${CURRENT_CNI_PLUGINS_DIR} to ${CNI_PLUGINS_DIR}."

##########################################################################################
# Generate the CNI configuration and move to CNI configuration directory
##########################################################################################

# The environment variables used to connect to the kube-apiserver
SERVICE_ACCOUNT_PATH=/var/run/secrets/kubernetes.io/serviceaccount
SERVICEACCOUNT_TOKEN=$(cat $SERVICE_ACCOUNT_PATH/token)
KUBE_CACERT=${KUBE_CACERT:-$SERVICE_ACCOUNT_PATH/ca.crt}
KUBERNETES_SERVICE_PROTOCOL=${KUBERNETES_SERVICE_PROTOCOL-https}

# Check if we're running as a k8s pod.
if [ -f "$SERVICE_ACCOUNT_PATH/token" ];
then
    # some variables should be automatically set inside a pod
    if [ -z "${KUBERNETES_SERVICE_HOST}" ]; then
        exit_with_message "KUBERNETES_SERVICE_HOST not set"
    fi
    if [ -z "${KUBERNETES_SERVICE_PORT}" ]; then
        exit_with_message "KUBERNETES_SERVICE_PORT not set"
    fi
fi

# exit if the NODE_NAME environment variable is not set.
if [[ -z "${NODE_NAME}" ]];
then
    exit_with_message "NODE_NAME not set."
fi


NODE_RESOURCE_PATH="${KUBERNETES_SERVICE_PROTOCOL}://${KUBERNETES_SERVICE_HOST}:${KUBERNETES_SERVICE_PORT}/api/v1/nodes/${NODE_NAME}"
NODE_SUBNET=$(curl --cacert "${KUBE_CACERT}" --header "Authorization: Bearer ${SERVICEACCOUNT_TOKEN}" -X GET "${NODE_RESOURCE_PATH}" | jq ".spec.podCIDR")

# Check if the node subnet is valid IPv4 CIDR address
IPV4_CIDR_REGEX="(((25[0-5]|2[0-4][0-9]|1?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|1?[0-9][0-9]?))(\/([8-9]|[1-2][0-9]|3[0-2]))([^0-9.]|$)"
if [[ ${NODE_SUBNET} =~ ${IPV4_CIDR_REGEX} ]]
then
    echo "${NODE_SUBNET} is a valid IPv4 CIDR address."
else
    exit_with_message "${NODE_SUBNET} is not a valid IPv4 CIDR address!"
fi

# exit if the NODE_NAME environment variable is not set.
if [[ -z "${CNI_NETWORK_CONFIG}" ]];
then
    exit_with_message "CNI_NETWORK_CONFIG not set."
fi

TMP_CONF='/minicni.conf.tmp'
cat >"${TMP_CONF}" <<EOF
${CNI_NETWORK_CONFIG}
EOF

# Replace the __NODE_SUBNET__
grep "__NODE_SUBNET__" "${TMP_CONF}" && sed -i s~__NODE_SUBNET__~"${NODE_SUBNET}"~g "${TMP_CONF}"

# Log the config file
echo "CNI config: $(cat "${TMP_CONF}")"

# Move the temporary CNI config into the CNI configuration directory.
mv "${TMP_CONF}" "${CNI_NET_DIR}/${CNI_CONF_NAME}" || \
  exit_with_error "Failed to move ${TMP_CONF} to ${CNI_CONF_NAME}."

echo "Created CNI config ${CNI_CONF_NAME}"

# Unless told otherwise, sleep forever.
# This prevents Kubernetes from restarting the pod repeatedly.
should_sleep=${SLEEP:-"true"}
echo "Done configuring CNI.  Sleep=$should_sleep"
while [ "$should_sleep" == "true"  ];
do
	sleep 10
done

