# DIGIO GAMES

Monolito Go para emulação SNES com cloud saves no PostgreSQL (Neon).

## Setup

1. **Banco de dados**
   - Crie um banco no [Neon](https://neon.tech)
   - Copie a connection string

2. **Configure o ambiente**
   ```bash
   cp .env.example .env
   ```
   Edite `.env` e adicione sua `DATABASE_URL`

3. **Adicione ROMs**
   - Coloque arquivos `.sfc` ou `.smc` na pasta `roms/`
   - Exemplo: `roms/SuperMario.sfc`

4. **Execute**
   ```bash
   go run main.go
   ```
   Ou compile:
   ```bash
   go build -o digio-games.exe .
   .\digio-games.exe
   ```

5. **Acesse**
   ```
   http://localhost:8080/digio.com.br/
   ```

## Rotas

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/digio.com.br/` | Biblioteca de ROMs |
| GET | `/digio.com.br/play/{rom}` | Player EmulatorJS |
| GET | `/digio.com.br/api/roms` | Lista ROMs (JSON) |
| GET | `/digio.com.br/api/roms/{name}` | Download ROM |
| GET | `/digio.com.br/api/saves/{rom}` | Lista saves |
| GET | `/digio.com.br/api/saves/{rom}/data` | Download save |
| POST | `/digio.com.br/api/saves/{rom}` | Upload save |

## Features

- EmulatorJS (SNES) via CDN
- Auto-save SRAM a cada 30s
- 3 slots de save state manual
- Sync automático com PostgreSQL (Neon)
- UI minimalista com ícones SVG
- SweetAlert2 + Toastify

## Estrutura

```
digio-games/
├── main.go                      → Entry point + router
├── internal/
│   ├── db/db.go                → PostgreSQL (pgx)
│   └── handler/handler.go      → HTTP handlers
├── web/
│   ├── embed.go                → Embed assets
│   ├── templates/              → HTML templates
│   └── static/                 → CSS + JS
└── roms/                       → Coloque .sfc/.smc aqui
```

## Variáveis de ambiente

```env
DATABASE_URL=postgresql://user:pass@ep-xxx.neon.tech/db?sslmode=require
PORT=8080
ROMS_DIR=roms
```
