cp /ihouqi/nps-auth/nps-auth.service /etc/systemd/system/nps-auth.service

sudo systemctl daemon-reload
sudo systemctl enable nps-auth
sudo systemctl restart nps-auth

# 查看状态
sudo systemctl status nps-auth

sudo systemctl start nps-auth
sudo systemctl stop nps-auth

ps |grep nps-quth |  grep -v grep