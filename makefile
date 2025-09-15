BIN_DIR := bin

API_BIN   := $(BIN_DIR)/notification_api
EMAIL_BIN := $(BIN_DIR)/email_worker
SMS_BIN   := $(BIN_DIR)/sms_worker
PUSH_BIN  := $(BIN_DIR)/push_worker

API_SRC   := cmd/notification_api
EMAIL_SRC := cmd/email_worker
SMS_SRC   := cmd/sms_worker
PUSH_SRC  := cmd/push-worker

.PHONY: all
all: build

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

.PHONY: build
build: build-api build-email build-sms 

.PHONY: build-api
build-api: | $(BIN_DIR)
	@echo "Building Notification API..."
	go build -o $(API_BIN) $(API_SRC)/main.go

.PHONY: build-email
build-email: | $(BIN_DIR)
	@echo "Building Email Worker..."
	go build -o $(EMAIL_BIN) $(EMAIL_SRC)/main.go

.PHONY: build-sms
build-sms: | $(BIN_DIR)
	@echo "Building SMS Worker..."
	go build -o $(SMS_BIN) $(SMS_SRC)/main.go

.PHONY: build-push
build-push: | $(BIN_DIR)
	@echo "Building Push Worker..."
	go build -o $(PUSH_BIN) $(PUSH_SRC)/main.go

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

.PHONY: run-push
run-push: build-push
	@echo "Running Push Worker..."
	$(PUSH_BIN)

.PHONY: clean
clean:
	@echo "Cleaning binaries..."
	rm -rf $(BIN_DIR)
