# Navigate back to project path
cd ../../

# Create a new user exclusively for this service
sudo useradd -r -s /usr/sbin/nologin display-service
sudo usermod -aG i2c display-service

# Copy the binary file to the linux binaries path
cd cmd/display
go build
sudo cp display /usr/local/bin/display

# Navigate back to project path
cd ../../

# Copy the service file to the systemd service files
sudo cp systemd/display/display.service /etc/systemd/system/display.service

# Navigate back to project path
cd ../../

# Enable and start the new service
sudo systemctl daemon-reexec
sudo systemctl daemon-reload
sudo systemctl enable display.service
sudo systemctl start display.service

# Check status
sudo systemctl status display.service