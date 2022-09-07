PHONY: generate

generate:
	openapi-generator generate \
		--input-spec glinet-openapi.yml \
		--generator-name go \
		--additional-properties=gitUserId=ryanrishi \
		--additional-properties=gitRepoId=glinet-client-go \
		--additional-properties=packageName=glinet

