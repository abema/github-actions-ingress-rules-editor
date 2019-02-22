#!/bin/sh
set -e

BRANCH_NAME="$GITHUB_REF"
echo $BRANCH_NAME

FEATURE_NAME=$(subst ${BRANCH_PREFIX},,$(BRANCH_NAME))
echo $FEATURE_NAME

INGRESS_HOST="${FEATURE_NAME}-${HOST_SUFFIX}"
echo $INGRESS_HOST

echo $GITHUB_EVENT_NAME

if [ "$GITHUB_EVENT_NAME" = "push" ]
then
    /builder/ingress_rules_editor add -ingress=${INGRESS_NAME} -host=${INGRESS_HOST} -service=${SERVICE_NAME} -port=${SERVICE_PORT} -path=${INGRESS_PATH} -namespace=${NAMESPACE}
elif [ "$GITHUB_EVENT_NAME" = "remove" ]
then
    /builder/ingress_rules_editor remove -ingress=${INGRESS_NAME} -host=${INGRESS_HOST} -path=${INGRESS_PATH} -namespace=${NAMESPACE}
else
    exit 1
fi

exit 0
