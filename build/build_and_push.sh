#!/bin/bash
image=quay.io/redhat/dedicated-admin-operator

operator-sdk build $image
docker push $image
