# Navigate back to project path
cd ../../

# Create a new user exclusively for this service
sudo useradd -r -s /usr/sbin/nologin display-service

# Copy the binary file to the linux binaries path
cd cmd/datetime
go build
sudo cp datetime /usr/local/bin/datetime

# Navigate back to project path
cd ../../

# Copy the service file to the systemd service files
sudo cp systemd/datetime/datetime.service /etc/systemd/system/datetime.service

# Navigate back to project path
cd ../../

# Enable and start the new service
sudo systemctl daemon-reexec
sudo systemctl daemon-reload
sudo systemctl enable datetime.service
sudo systemctl start datetime.service

# Check status
sudo systemctl status datetime.service