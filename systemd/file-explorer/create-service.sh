# Navigate back to project path
cd ../../

# Create a new user exclusively for this service
sudo useradd -r -s /usr/sbin/nologin file-explorer-service

# Create the media directory witch will be the root path
sudo mkdir -p /opt/file-explorer-root
sudo chown file-explorer-service:file-explorer-service /opt/file-explorer-root
sudo chmod 777 /opt/file-explorer-root

# Copy the binary file to the linux binaries path
cd cmd/file-explorer
go build
sudo cp file-explorer /usr/local/bin/file-explorer

# Navigate back to project path
cd ../../

# Copy the service file to the systemd service files
sudo cp systemd/file-explorer/file-explorer.service /etc/systemd/system/file-explorer.service

# Navigate back to project path
cd ../../

# Enable and start the new service
sudo systemctl daemon-reexec
sudo systemctl daemon-reload
sudo systemctl enable file-explorer.service
sudo systemctl start file-explorer.service

# Check status
sudo systemctl status file-explorer.service