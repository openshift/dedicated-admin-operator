#!/bin/bash -x

set -e

cd $(dirname $0)/..

# verify template is updated
# if running `make generate-syncset` changes anything, fail the build
# order is inconsistent across systems, sort the template file.. it's not perfect but it's better than nothing
cat build/templates/olm-artifacts-template.yaml.tmpl | sort > sorted-before.yaml.tmpl

IN_DOCKER_CONTAINER=true make generate-syncset

cat build/templates/olm-artifacts-template.yaml.tmpl | sort > sorted-after.yaml.tmpl

diff sorted-before.yaml.tmpl sorted-after.yaml.tmpl || (echo "Running 'make generate-syncset' caused changes.  Run 'make generate-syncset' and commit changes to the PR to try again." && rm -f sorted-before.yaml.tmpl sorted-after.yaml.tmpl && exit 1)

rm -f sorted-before.yaml.tmpl sorted-after.yaml.tmpl

# it's okay to omit the IMAGE_REPOSITORY since this is just a PR test
make docker-build && make clean
