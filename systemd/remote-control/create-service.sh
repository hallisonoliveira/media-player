# Navigate back to project path
cd ../../

# Create a new user exclusively for this service
sudo useradd -r -s /usr/sbin/nologin remote-control-service
sudo usermod -aG input remote-control-service

# Copy the binary file to the linux binaries path
cd cmd/remote-control
go build
sudo cp remote-control /usr/local/bin/remote-control

# Navigate back to project path
cd ../../

# Copy the service file to the systemd service files
sudo cp systemd/remote-control/remote-control.service /etc/systemd/system/remote-control.service

# Navigate back to project path
cd ../../

# Enable and start the new service
sudo systemctl daemon-reexec
sudo systemctl daemon-reload
sudo systemctl enable remote-control.service
sudo systemctl start remote-control.service

# Check status
sudo systemctl status remote-control.service