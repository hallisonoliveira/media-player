- Copiar o arquivo de mapeamento para o diretório do ir-keytable
sudo cp hp_rc1762302-00_rc6.toml /lib/udev/rc_keymaps/hp_rc1762302-00_rc6.toml

- Criar o arquivo do systemd
sudo vim /etc/systemd/system/ir-keytable.service

- Habilitar o serviço para ser executado na inicialização do linux
sudo systemctl enable ir-keytable
