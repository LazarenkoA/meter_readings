
`docker build -t meter_readings -f docker/Dockerfile .`
`docker run -d --name meter_readings --restart always -v ./data:/app/data meter_readings`