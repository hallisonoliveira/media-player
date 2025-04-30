# Raspberry Pi MP3 Media Player

Este projeto é um sistema modular baseado em microserviços que transforma um Raspberry Pi em um **media player para arquivos MP3**, com foco em uma interface inspirada em **CD players clássicos**. Cada serviço é isolado, se comunica via **Redis Pub/Sub** e é gerenciado individualmente através do **systemd**.

## Visão Geral dos Serviços

O sistema é composto por diversos serviços independentes, cada um com uma responsabilidade única:

| Serviço        | Descrição                                                                 |
|----------------|---------------------------------------------------------------------------|
| `datetime`     | Publica a data e hora atual periodicamente em um canal Redis.             |
| `display`      | Mostra informações no display LCD 20x2 via I²C, com prioridade configurável. |
| `ir-remote`    | Escuta comandos de um controle remoto infravermelho e os publica no Redis.|
| `panel`        | Lê os botões físicos (GPIO) do painel e publica comandos no Redis.        |
| `player`       | Reproduz arquivos MP3 e publica o status e metadados das faixas.          |

Cada serviço executa de forma isolada como um **serviço systemd**, permitindo controle individual com `systemctl`.

## Comunicação entre Serviços

A comunicação entre os serviços é feita através de canais Redis utilizando o padrão **Pub/Sub**:

- O serviço `datetime` publica mensagens no canal `datetime`.
- O `display` assina múltiplos canais e exibe as mensagens com base em prioridades (ex: mostrar a música atual ao invés da hora).
- O `player` publica status da música em execução (tempo, artista, título).
- `panel` e `ir-remote` publicam comandos como `play`, `pause`, `next`, `prev` em canais específicos.
- O `player` escuta esses comandos para controlar a reprodução.

## Requisitos

- Raspberry Pi (qualquer modelo com GPIO e I²C)
- Display LCD 20x2 com módulo I²C
- Controle remoto infravermelho e receptor IR
- Redis instalado e em execução
- Go 1.21 ou superior
- `mpg123` ou outro player de linha de comando para reprodução MP3

## Instalação

### 1. Instalar o Redis
Atualizar o sistema:
```bash
sudo apt update
```

Instalação:
```bah
sudo apt install redis-server
```
Configuração no _systemd_ para que o Redis seja inicializado juntamente com o sistema:
```bash
sudo systemctl enable redis-server
```

### 2. Instalar o Go (golang)
Baixar o pacote _tar_ referente à versão e plataforma:
```bash
wget https://go.dev/dl/go1.22.3.linux-armv6l.tar.gz
```

Instalar:
```bash
sudo tar -C /usr/local -xzf go1.22.3.linux-armv6l.tar.gz
```

É necessário configurar a variável PATH para que o Go seja reconhecido como um comando do sistema e demais configurações de ambiente.
Editar o arquivo `profile`
```bash
vim ~/.profile
```
Adicionar as seguintes variáveis de ambiente no final do arquivo:
```bash
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

Para validar se a instalação e configuração foram feitas corretamente:
```bash
go version
```

## Licença
Este projeto está licenciado sob a [MIT License](https://opensource.org/licenses/MIT).
