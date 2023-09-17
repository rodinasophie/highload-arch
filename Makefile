build-proj:
	CGO_ENABLED=0 go build -o bin/social-network -mod vendor main.go

docker-init:
	docker compose up -d && sleep 2
	docker exec -it highload-arch-db sh -c "psql -U admin_user -f /etc/highload-arch/schema.sql social_net";

docker-reset:
	docker compose down

