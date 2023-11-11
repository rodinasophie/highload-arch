build-server:
	CGO_ENABLED=0 go build -gcflags="all=-N -l" -o bin/social-network -mod vendor main.go

build-client:
	CGO_ENABLED=0 go build -gcflags="all=-N -l" -o bin/social-network-client -mod vendor client/main.go

docker-init:
	docker compose up -d db-leader  && sleep 2;
	docker exec  -it ha-db-leader sh -c "chmod 0600 /root/.pgpass; psql -h localhost -U admin_user -d postgres -c \"drop role if exists replicator; create role replicator with login replication password 'pass';\" ";
	docker exec -d ha-db-leader sh -c "cd /etc/ && ./pg_setup.sh" && sleep 5;
	docker exec -it ha-db-leader sh -c "rm -rf /pgslave; mkdir /pgslave; pg_basebackup -h localhost -D /pgslave -U replicator -w -v -P --wal-method=stream";
	docker compose up  -d db-replica-1 db-replica-2
	docker exec -it ha-db-leader sh -c "psql -U admin_user -f /etc/highload-arch/schema.sql social_net";
	docker exec -it ha-db-leader sh -c "psql -U admin_user -d social_net -c \"COPY users(first_name, second_name, birthdate, city, biography) FROM '/people.csv' DELIMITER U&'\0009' CSV HEADER;\"";
	docker exec -it ha-db-leader sh -c "psql -U admin_user -d social_net -c \" UPDATE users SET biography = 'Empty' WHERE biography IS NULL;\"";

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

load-for-write:
	python3 perf-testing/load.py