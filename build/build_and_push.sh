#!/bin/bash
image=quay.io/rogbas/dedicated-admin-operator

operator-sdk build $image
docker push $image