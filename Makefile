build-server:
	CGO_ENABLED=0 go build -gcflags="all=-N -l" -o bin/social-network -mod vendor main.go

build-client:
	CGO_ENABLED=0 go build -gcflags="all=-N -l" -o bin/social-network-client -mod vendor client/main.go

docker-init:
	docker compose up -d && sleep 2
	docker exec -it highload-arch-db sh -c "psql -U admin_user -f /etc/highload-arch/schema.sql social_net";

docker-run:
	docker exec -d highload-arch-backend sh -c "./bin/social-network";

docker-reset:
	docker compose down

generate:
	python3 data-generator/generate.py

select-prefix:
	python3 perf-testing/prefix_selection.py


perf-test:
	locust -f perf-testing/main.py  --host=http://localhost:8083 
