# Build the image
docker build -t txanalyser-db-sqlite .

# Run the image
docker run -it \
	-v "$(pwd)/data:/app/data" \
	-v "$(pwd)/scripts:/app/scripts" \
	txanalyser-db-sqlite
