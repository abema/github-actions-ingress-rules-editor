#!/bin/sh
set -e

BRANCH_NAME=$(echo $GITHUB_REF | sed -e "s/^refs\/heads\///g")
echo $BRANCH_NAME

FEATURE_NAME=$(echo $BRANCH_NAME | sed -e "s/^${BRANCH_PREFIX}//g")
echo $FEATURE_NAME

INGRESS_HOST="$FEATURE_NAME"
if [ "$HOST_PREFIX" != "" ] && [ "$HOST_SUFFIX" != "" ]; then
  INGRESS_HOST="${HOST_PREFIX}${FEATURE_NAME}${HOST_SUFFIX}"
elif [ "$HOST_SUFFIX" != "" ]; then
  INGRESS_HOST="${FEATURE_NAME}${HOST_SUFFIX}"
elif [ "$HOST_PREFIX" != "" ]; then
  INGRESS_HOST="${HOST_PREFIX}${FEATURE_NAME}"
fi
echo $INGRESS_HOST

SERVICE_NAME="$FEATURE_NAME"
if [ "$SERVICE_PREFIX" != "" ] && [ "$SERVICE_SUFFIX" != "" ]; then
  SERVICE_NAME="${SERVICE_PREFIX}${FEATURE_NAME}${SERVICE_SUFFIX}"
elif [ "$SERVICE_SUFFIX" != "" ]; then
  SERVICE_NAME="${FEATURE_NAME}${SERVICE_SUFFIX}"
elif [ "$SERVICE_PREFIX" != "" ]; then
  SERVICE_NAME="${SERVICE_PREFIX}${FEATURE_NAME}"
fi
echo $SERVICE_NAME

/builder/ingress_rules_editor ${OPERATION} -ingress="$INGRESS_NAME" -host="$INGRESS_HOST" -service=$"SERVICE_NAME" -port=$"SERVICE_PORT" -path="$INGRESS_PATH" -namespace="$NAMESPACE"
