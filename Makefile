BINFILE=build/_output/bin/dedicated-admin-operator
MAINPACKAGE=./cmd/manager
GOENV=GOOS=linux GOARCH=amd64 CGO_ENABLED=0
GOFLAGS=-gcflags="all=-trimpath=${GOPATH}" -asmflags="all=-trimpath=${GOPATH}"

OPERATORNAME := dedicated-admin-operator

# Where do CSVs end up? Note, the CSV generator creates a subdirectory,
# called $(FINALCSVDIR)
CSVBASEDIR := csvs
# FINALCSVDIR is created by the Python script
FINALCSVDIR := $(CSVBASEDIR)/$(OPERATORNAME)

CATALOG_SOURCEIMG := quay.io/redhat/$(OPERATORNAME)-cs


.PHONY: check
check: ## Lint code
	gofmt -s -l $(shell go list -f '{{ .Dir }}' ./... ) | grep ".*\.go"; if [ "$$?" = "0" ]; then gofmt -s -d $(shell go list -f '{{ .Dir }}' ./... ); exit 1; fi
	go vet ./cmd/... ./pkg/...

.PHONY: build
build: ## Build binary
	${GOENV} go build ${GOFLAGS} -o ${BINFILE} ${MAINPACKAGE}

.PHONY: clean
clean:
	rm -rf $(FINALCSVDIR)

render-csv:
	mkdir -p $(FINALCSVDIR) ; \
	prevver=$$(find $(FINALCSVDIR) -type d -mindepth 1 | sort -V | tail -n1 | cut -d '/' -f 3) ; \
	[[ -z $$prevver ]] && prevver="__undefined__" ; \
	echo "Previous Version: $$prevver" ; \
	./scripts/gen_operator_csv.py $(CSVBASEDIR) $$prevver $(CATALOG_SOURCEIMG):latest ; \
	curver=$$(find $(FINALCSVDIR) -type d -mindepth 1 | sort -V | tail -n1 | cut -d '/' -f 3) ; \
	sed -e "s!\$$CURRENTCSV!$(OPERATORNAME).v$$curver!g" scripts/templates/package-template.yaml > $(FINALCSVDIR)/$(OPERATORNAME).package.yaml

.PHONY: build-catalogsource-image
build-catalogsource-image: render-csv
	curver=$$(find $(FINALCSVDIR) -type d -mindepth 1 | sort -V | tail -n1 | cut -d '/' -f 3) ; \
	docker build --build-arg csvsrc=$(FINALCSVDIR) --no-cache -f build/Dockerfile.catalogsource -t $(CATALOG_SOURCEIMG):latest . ; \
	docker tag $(CATALOG_SOURCEIMG):latest $(CATALOG_SOURCEIMG):v$$curver