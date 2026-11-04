STEAMPIPE_INSTALL_DIR ?= ~/.steampipe
BUILD_TAGS = netgo
install:
	go build -o $(STEAMPIPE_INSTALL_DIR)/plugins/hub.steampipe.io/plugins/hepapi/azuredevops-self-hosted@latest/steampipe-plugin-azuredevops-self-hosted.plugin -tags "${BUILD_TAGS}" *.go
