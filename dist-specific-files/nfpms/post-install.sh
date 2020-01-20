echo "...Changing to executable"
chmod +x /opt/goEDMS/goEDMS
echo "...Changing permissions"
chown -R goEDMS:goEDMS /opt/goEDMS
echo "...Enabling systemd service"
systemctl enable goEDMS.service
echo "...Starting goEDMS service"
systemctl start goEDMS.service
