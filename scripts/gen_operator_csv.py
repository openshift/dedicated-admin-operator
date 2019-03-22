#!/usr/bin/env python
#
# Generate an operator bundle for publishing to OLM. Copies appropriate files
# into a directory, and composes the ClusterServiceVersion which needs bits and
# pieces of our rbac and deployment files.
#
# Usage ./scripts/gen_operator_csv.py OUTPUT_DIR PREVIOUS_VERSION IMAGE_NAME
# Example: scripts/gen_operator_csv.py tmp 0.1 quay.io/redhat/dedicated-admin-operator:latest

import datetime
import os
import sys
import yaml
import shutil
import subprocess


def get_git_sha():
    sha = subprocess.check_output('git rev-parse HEAD', shell=True)
    return str(sha)[0:7]

def get_num_commits():
    num = subprocess.check_output('git rev-list HEAD --count', shell=True)
    return num.rstrip()


if __name__ == '__main__':

    # This script will append the current number of commits given as an arg
    # (presumably since some past base tag), and the git hash arg for a final
    # version like: 0.1.189-3f73a592
    VERSION_BASE = "0.1"

    if len(sys.argv) != 4:
        print("USAGE: %s OUTPUT_DIR PREVIOUS_VERSION IMAGE_NAME" % sys.argv[0])
        sys.exit(1)

    outdir = sys.argv[1]
    prev_version = sys.argv[2]
    operator_image = sys.argv[3]
    git_num_commits = get_num_commits()
    git_sha = get_git_sha()

    full_version = "%s.%s-%s" % (VERSION_BASE, git_num_commits, git_sha)
    print("Generating CSV for version: %s" % full_version)

    if not os.path.exists(outdir):
        os.mkdir(outdir)

    version_dir = os.path.join(outdir, full_version)
    if not os.path.exists(version_dir):
        os.mkdir(version_dir)

    with open('scripts/templates/csv-template.yaml', 'r') as stream:
        csv = yaml.load(stream)

    csv['spec']['install']['spec']['clusterPermissions'] = []

    cluster_roles = [
        {'permissions': 'manifests/05-dedicated-admin-operator.ClusterRole.yaml', 'serviceAccountName': 'dedicated-admin-operator'}
    ]

    for role in cluster_roles:
    # Add our operator role to the CSV:
        with open(role['permissions'], 'r') as stream:
            operator_role = yaml.load(stream)
            csv['spec']['install']['spec']['clusterPermissions'].append(
                {
                    'rules': operator_role['rules'],
                    'serviceAccountName': role['serviceAccountName'],
                })


    # Add our deployment spec
    with open('manifests/51-dedicated-admin-operator.Deployment.yaml', 'r') as stream:
        operator_components = []
        operator = yaml.load_all(stream)
        for doc in operator:
            operator_components.append(doc)
        operator_deployment = operator_components[0]
        csv['spec']['install']['spec']['deployments'][0]['spec'] = operator_deployment['spec']

    # Update the deployment to use the defined image:
    csv['spec']['install']['spec']['deployments'][0]['spec']['template']['spec']['containers'][0]['image'] = operator_image

    # Update the versions to include git hash:
    csv['metadata']['name'] = "dedicated-admin-operator.v%s" % full_version
    csv['spec']['version'] = full_version
    csv['spec']['replaces'] = "dedicated-admin-operator.v%s" % prev_version

    # Set the CSV createdAt annotation:
    now = datetime.datetime.now()
    csv['metadata']['annotations']['createdAt'] = now.strftime("%Y-%m-%dT%H:%M:%SZ")

    # Write the CSV to disk:
    csv_filename = "dedicated-admin-operator.v%s.clusterserviceversion.yaml" % full_version
    csv_file = os.path.join(version_dir, csv_filename)
    with open(csv_file, 'w') as outfile:
        yaml.dump(csv, outfile, default_flow_style=False)
    print("Wrote ClusterServiceVersion: %s" % csv_file)

