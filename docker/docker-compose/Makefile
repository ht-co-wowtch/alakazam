.PHONY: platform.rollback platform.migrate platform.seed platform.reset

platform.rollback:
	docker-compose run mig-platform sh -c "wait-for -t 300 mysql:3306 -- make rollback"

platform.migrate:
	docker-compose run mig-platform sh -c "wait-for -t 300 mysql:3306 -- make migrate"

platform.seed:
	docker-compose run mig-platform sh -c "wait-for -t 300 mysql:3306 -- make seed"

platform.reset:
	docker-compose run mig-platform sh -c "wait-for -t 300 mysql:3306 -- make reset"

.PHONY: alakazam.rollback alakazam.migrate alakazam.seed alakazam.reset

alakazam.rollback:
	docker-compose run mig-alakazam sh -c "wait-for -t 300 mysql:3306 -- make rollback"

alakazam.migrate:
	docker-compose run mig-alakazam sh -c "wait-for -t 300 mysql:3306 -- make migrate"

alakazam.seed:
	docker-compose run mig-alakazam sh -c "wait-for -t 300 mysql:3306 -- make seed"

alakazam.reset:
	docker-compose run mig-alakazam sh -c "wait-for -t 300 mysql:3306 -- make reset"
