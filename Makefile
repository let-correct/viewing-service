test:
	go test ./...

push-viewing:
	./scripts/push-to-ecr.sh $$(git rev-parse --short HEAD) viewing

push-all:
	$(MAKE) push-viewing
