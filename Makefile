create-migrations:
	find . -type f -name "schema.sql" -exec cat {} \; > combined.sql
	python create-migrations.py $(TITLE)

generate-models:
	sqlc generate

apply-migrations:
	python migrate.py 