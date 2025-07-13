
BIN_DIR=bin

API_BIN=$(BIN_DIR)/notification_api
EMAIL_BIN=$(BIN_DIR)/email_worker
SMS_BIN=$(BIN_DIR)/sms_worker

API_SRC=cmd/api/server
EMAIL_SRC=cmd/email_worker
SMS_SRC=cmd/sms_worker

.PHONY: all
all: build

.PHONY: build
build: build-api build-email build-sms

.PHONY: build-api
build-api:
	@echo "Building Notification API..."
	go build -o $(API_BIN) $(API_SRC)/main.go

.PHONY: build-email
build-email:
	@echo "Building Email Worker..."
	go build -o $(EMAIL_BIN) $(EMAIL_SRC)/main.go

.PHONY: build-sms
build-sms:
	@echo "Building SMS Worker..."
	go build -o $(SMS_BIN) $(SMS_SRC)/main.go

.PHONY: run-api
run-api: build-api
	@echo "Running Notification API..."
	$(API_BIN)

.PHONY: run-email
run-email: build-email
	@echo "Running Email Worker..."
	$(EMAIL_BIN)

.PHONY: run-sms
run-sms: build-sms
	@echo "Running SMS Worker..."
	$(SMS_BIN)

.PHONY: clean
clean:
	@echo "Cleaning binaries..."
	rm -rf $(BIN_DIR)

