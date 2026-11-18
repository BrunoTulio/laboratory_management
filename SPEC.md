# Spec Doc — Sistema de Gestão e Agendamento de Laboratório Escolar (`laboratory-management`)

## 1. Visão Geral

O **Sistema de Gestão e Agendamento de Laboratório Escolar** é uma aplicação web monolítica desenvolvida em **Go** com **PostgreSQL**, projetada para controlar a reserva e o uso de laboratórios escolares por professores, alunos e direção.

O projeto foi estruturado com fins educacionais em duas fases:
- **Fase 1 (Atual):** Implementação do backend em Go (regras de negócio, persistência, autenticação, autorização, notificações, auditoria) e esqueletos HTML/CSS/JS funcionais.
- **Fase 2 (Planejada):** Desenvolvimento do design visual, identidade, responsividade avançada e calendário interativo pelos alunos, seguido da integração completa com os handlers e templates do backend.

---

## 2. Stack Tecnológica e Infraestrutura

| Camada | Tecnologia / Biblioteca | Observações |
| :--- | :--- | :--- |
| Linguagem | Go `1.27.1` | Módulo `laboratory-management` (`go.mod`) |
| Servidor HTTP | `net/http` (stdlib) | Roteamento via `http.ServeMux` |
| Renderização | `html/template` (stdlib) | Templates em `templates/**/*.html` |
| Banco de Dados | PostgreSQL (`17` no `compose.yaml`) | Driver `github.com/lib/pq v1.12.3` (SQL puro, sem ORM) |
| Segurança / Hash | `golang.org/x/crypto v0.57.0` | `bcrypt` para hash de senhas |
| Configuração | `github.com/joho/godotenv v1.5.1` | Leitura de `.env` e variáveis de ambiente |
| Containerização | Docker / Docker Compose | `Dockerfile` e `compose.yaml` (serviço `db` na porta `5433:5432`) |

### Variáveis de Ambiente (`config/config.go`)

| Variável | Padrão | Descrição |
| :--- | :--- | :--- |
| `APP_ENV` | `development` | Ambiente de execução (`development` / `production`) |
| `APP_PORT` | `8080` | Porta HTTP do servidor |
| `DATABASE_URL` | `postgres://user:password@localhost:5432/laboratory_management?sslmode=disable` | String de conexão PostgreSQL |
| `SESSION_SECRET` | `change-this-secret-key` | Segredo para sessões |
| `CSRF_SECRET` | `change-this-csrf-secret` | Segredo para proteção CSRF |

---

## 3. Arquitetura do Sistema

O sistema adota uma **arquitetura em camadas** organizada por domínio em `internal/`:

```text
Requisição HTTP
   │
   ▼
Cadeia de Middlewares Globais (Router.Use)
   ├── Logging   (registra método, URI e duração)
   ├── Session   (verifica presença do cookie session_token)
   └── CSRF      (exige csrf_token / X-CSRF-Token em métodos não seguros)
   │
   ▼
Middlewares de Rota
   ├── AuthMiddleware (valida sessão e injeta userID, userEmail, userRole no context)
   └── RequireAdmin / RequireRole (RBAC por perfil)
   │
   ▼
Handler (internal/<domínio>/handler.go)
   │
   ▼
Service (internal/<domínio>/service.go)  ──► Regras de negócio, notificações, auditoria
   │
   ▼
Repository (internal/<domínio>/repository.go) ──► Queries SQL (database/sql + lib/pq)
   │
   ▼
PostgreSQL
```

### Estrutura de Diretórios

- `cmd/server/main.go`: Ponto de entrada; carrega `.env`, inicializa conexão com o banco, instancia `web.NewRouter(db)`, registra middlewares globais e inicia `http.ListenAndServe`.
- `config/`: Carregamento de configurações via variáveis de ambiente.
- `internal/database/`: Inicialização do pool PostgreSQL (`MaxOpenConns=25`, `MaxIdleConns=25`, `ConnMaxLifetime=300s`) e executor de migration (`RunMigrations`).
- `internal/middleware/`: Middlewares `Logging`, `Session`, `CSRF`, `AuthMiddleware`, `RequireRole`, `RequireAdmin`, `RequireTeacherOrAdmin`.
- `internal/web/router.go`: Composição de dependências (repositories → services → handlers), carregamento de templates e registro de todas as rotas.
- `internal/auth/`: Autenticação (bcrypt, geração de token de sessão, login, registro, logout, perfil).
- `internal/user/`: Modelo, repositório, serviço e handler de usuários e perfis.
- `internal/laboratory/`: CRUD de laboratórios.
- `internal/class/`: CRUD de turmas.
- `internal/booking/`: Modelo, repositório, serviço de agendamento e `AvailabilityService`.
- `internal/notification/`: Criação, listagem e marcação de leitura de notificações.
- `internal/audit/`: Registro e consulta de logs de auditoria.
- `templates/`: Layouts (`base.html`, `auth.html`), componentes (`navbar.html`, `sidebar.html`, `footer.html`, `form-field.html`) e 18 páginas HTML.
- `static/`: Arquivos CSS (`main.css`, `auth.css`) e JS (`main.js`, `auth.js`).
- `migrations/001_initial_schema.sql`: Schema inicial e seed de dados.
- `tests/`: Testes unitários em `tests/auth/`, `tests/booking/` e `tests/validator/`.

---

## 4. Modelo de Dados (`migrations/001_initial_schema.sql`)

O banco PostgreSQL é composto por **7 tabelas principais**:

```text
roles 1──* users 1──* bookings *──1 laboratories
                      │     │
                      │     *──0..1 classes
                      │
                      1──* notifications

users 1──* audit_logs
```

1. **`roles`**: Perfis de acesso (`id`, `name` UNIQUE, `description`, timestamps).
   - Seed inicial: `PROFESSOR`, `ALUNO`, `DIRECAO`.
2. **`users`**: Usuários (`id`, `name`, `email` UNIQUE, `password_hash`, `role_id` FK, `registration_number`, `phone`, `status` padrão `'ACTIVE'`, timestamps).
   - Status no modelo Go: `ACTIVE`, `INACTIVE`, `BLOCKED`.
3. **`laboratories`**: Laboratórios (`id`, `name`, `description`, `location`, `capacity`, `status` padrão `'ACTIVE'`, timestamps).
   - Seed inicial: `'Laboratório Principal'` (`Bloco A - Sala 101`, capacidade `30`).
4. **`classes`**: Turmas (`id`, `name`, `course`, `grade_year`, `shift`, `status` padrão `'ACTIVE'`, timestamps).
5. **`bookings`**: Agendamentos (`id`, `user_id` FK, `laboratory_id` FK, `class_id` FK opcional, `booking_date`, `start_time`, `end_time`, `shift`, `reason`, `observation`, `status` padrão `'PENDING'`, campos de aprovação `approved_by`/`approved_at`, rejeição `rejected_by`/`rejected_at`/`rejection_reason`, cancelamento `cancelled_by`/`cancelled_at`/`cancellation_reason`, timestamps).
   - Status no modelo Go: `PENDING`, `APPROVED`, `REJECTED`, `CANCELLED`, `CONCLUIDO`.
6. **`notifications`**: Notificações por usuário (`id`, `user_id` FK, `booking_id` FK opcional, `type`, `title`, `message`, `is_read` padrão `FALSE`, `created_at`).
7. **`audit_logs`**: Trilha de auditoria (`id`, `user_id` FK opcional, `action`, `entity_type`, `entity_id`, `details`, `ip_address`, `user_agent`, `created_at`).

Índices otimizados cobrem `users(email, role_id, status)`, `bookings(user_id, laboratory_id, class_id, booking_date, status, (booking_date, start_time, end_time))`, `notifications(user_id, is_read)` e `audit_logs(user_id, action, (entity_type, entity_id))`.

---

## 5. Perfis de Usuário e Controle de Acesso

| Funcionalidade | `ALUNO` | `PROFESSOR` | `DIRECAO` |
| :--- | :---: | :---: | :---: |
| Cadastro, Login, Logout, Perfil | ✓ | ✓ | ✓ |
| Dashboard do usuário | ✓ | ✓ | ✓ |
| Consultar disponibilidade e agenda | ✓ | ✓ | ✓ |
| Solicitar agendamento | ✓ (quando permitido) | ✓ | ✓ |
| Visualizar próprios agendamentos e histórico | ✓ | ✓ | ✓ |
| Visualizar notificações e marcar como lidas | ✓ | ✓ | ✓ |
| Dashboard administrativo (`/admin`) | — | — | ✓ |
| Aprovar / Rejeitar / Cancelar qualquer agendamento | — | — | ✓ |
| Gerenciar usuários, laboratórios e turmas | — | — | ✓ |
| Visualizar logs de auditoria | — | — | ✓ |

---

## 6. Regras de Negócio

### 6.1. Ciclo de Vida do Agendamento

```text
[Criação] ──► PENDING ──┬──(Aprovação + re-verificação de conflito)──► APPROVED ──► CONCLUIDO
                        ├──(Rejeição + motivo)──────────────────────► REJECTED
                        └──(Cancelamento + motivo, sem DELETE)──────► CANCELLED
```

- **Criação (`BookingService.CreateBooking`):**
  1. Valida `LaboratoryID`, `BookingDate`, `StartTime < EndTime` e `Reason`.
  2. Verifica conflito de horário no laboratório e data informados contra agendamentos `APPROVED` e `PENDING`.
  3. Persiste o agendamento com status `PENDING`.
  4. Registra log de auditoria (`CREATE_BOOKING`) e notificação para o solicitante.
- **Aprovação (`BookingService.ApproveBooking`):**
  1. Exige status atual `PENDING`.
  2. Re-executa a verificação de conflito excluindo o próprio `bookingID` para evitar condições de corrida.
  3. Atualiza para `APPROVED` com `approved_by` e `approved_at`.
  4. Registra auditoria e notifica o solicitante.
- **Rejeição (`BookingService.RejectBooking`):**
  1. Exige status atual `PENDING`.
  2. Atualiza para `REJECTED` com `rejected_by`, `rejected_at` e `rejection_reason`.
  3. Registra auditoria e notifica o solicitante.
- **Cancelamento (`BookingService.CancelBooking`):**
  1. Nunca exclui fisicamente o registro (`DELETE`); altera o status para `CANCELLED`, gravando `cancelled_by`, `cancelled_at` e `cancellation_reason`.
  2. Registra auditoria e notificação.

### 6.2. Detecção de Conflito de Horários (`TimeRange.HasConflict`)

Dois intervalos `[startA, endA]` e `[startB, endB]` no mesmo laboratório e data **não** conflitam apenas se um terminar antes ou exatamente quando o outro começa (`endA <= startB || endB <= startA`). Qualquer sobreposição parcial ou total gera conflito.

### 6.3. Disponibilidade (`AvailabilityService`)

- Gera grade padrão horária das `08:00` às `12:00` e `13:00` às `18:00` (9 slots de 1 hora).
- Classifica cada slot como `AVAILABLE`, `OCCUPIED` (quando há reserva `APPROVED`) ou `PENDING` (quando há reserva `PENDING`).

---

## 7. Mapeamento de Rotas HTTP (`internal/web/router.go`)

### Públicas
- `GET /static/*` — Arquivos estáticos
- `GET|POST /login` — Autenticação
- `GET|POST /cadastro` — Registro de usuário
- `GET /logout` — Encerramento de sessão
- `GET /403`, `GET /404`, `GET /500` — Páginas de erro (e `/` redireciona para `/login`)

### Protegidas (`AuthMiddleware`)
- `GET /dashboard` — Dashboard do usuário
- `GET /perfil` — Perfil do usuário
- `GET /agendamentos` — Lista de agendamentos do usuário
- `GET|POST /agendamentos/novo` — Criação de agendamento
- `GET /agendamentos/` — Detalhes do agendamento (via query `?id=`)
- `GET /disponibilidade` — Consulta de disponibilidade
- `GET /notificacoes`, `GET /notificacoes/nao-lidas`, `POST /notificacoes/marcar-lida`, `POST /notificacoes/marcar-todas-lidas`
- `GET /agenda`, `/agenda/dia`, `/agenda/semana`, `/agenda/mes`
- `GET /historico`

### Administrativas (`RequireAdmin` — perfil `DIRECAO`)
- `GET /admin` — Dashboard administrativo
- `GET /admin/usuarios`, `GET|POST /admin/usuarios/editar`
- `GET /admin/agendamentos`, `GET /admin/agendamentos/pendentes`
- `POST /admin/agendamentos/aprovar`, `POST /admin/agendamentos/rejeitar`, `POST /admin/agendamentos/cancelar`
- `GET /admin/historico`
- `GET /admin/auditoria`
- `GET /admin/laboratorios`, `/admin/laboratorios/criar`, `/admin/laboratorios/editar`
- `GET /admin/turmas`, `/admin/turmas/criar`, `/admin/turmas/editar`

---

## 8. Testes Automatizados

Os testes ficam em `tests/`:
- `tests/booking/booking_conflict_test.go`: Testa `TimeRange.HasConflict` (limites exatos, sobreposição parcial, inclusão total, casos de exemplo dos requisitos) e constantes de status/turno.
- `tests/auth/auth_test.go`: Testes de hash de senha e tokens de autenticação.
- `tests/validator/validator_test.go`: Testes de validação de entrada.

---

## 9. Estado Atual, Lacunas Técnicas e Próximos Passos

### Corrigido na Base Atual
- **Sessão e Autenticação (`internal/middleware/auth.go`, `internal/auth/service.go`):** Implementados `SignSessionToken` e `VerifySessionToken` com HMAC-SHA256 e expiração, além de `ValidateSession`.
- **Contexto de Usuário nos Handlers (`internal/auth/handler.go`, `internal/booking/handler.go`):** Substituídos os placeholders locais por `middleware.GetUserIDFromContext` e `middleware.GetUserRoleFromContext`.
- **Query de Conflito SQL (`internal/booking/repository.go`):** Corrigido `CheckConflict` para tratar `excludeID == 0` via `($3 = 0 OR id != $3)` e comparar sobreposição de intervalo com `start_time < $5 AND end_time > $4`.
- **Execução de Migrations no Startup (`cmd/server/main.go`):** Adicionada chamada a `database.RunMigrations()` após `database.Initialize()`.

### Pendências para Produção e Fase 2
1. **Validação Estrita de CSRF (`internal/middleware/csrf.go`):** O middleware `CSRF` verifica apenas se o token está não-vazio, sem validar contra `CSRFMiddleware.ValidateToken`.
2. **Renderização de Templates vs. JSON:** Handlers de domínio (`auth`, `user`, `booking`, `audit`) ainda respondem JSON nos métodos `render*`, e `router.go` usa dados mockados/zerados em alguns handlers.
3. **Disciplinas no Banco:** A tela `Cadastros` do protótipo inclui `Professores`, `Turmas` e `Disciplinas` — requer entidade/tabela `subjects` (além de `users` e `classes`).

---

## 10. Especificação das Telas (Protótipo CETEP Bacia do Paramirim)

### Identidade Visual Global
- **Tema:** Claro, fundo off-white/creme (`#F7F7F2`), cards brancos (`#FFFFFF`) com borda sutil (`#E5E7EB`) e cantos arredondados (`14px`–`16px`).
- **Cor primária:** Verde-floresta escuro (`#0B5E42`) para botões principais, ícone do cabeçalho e eyebrows; verde-menta claro (`#E6F4EE`) para pills e aba ativa.
- **Tipografia:** Títulos principais em **serifa** (`Georgia, 'Times New Roman', serif`); rótulos, números e textos de apoio em **sans-serif** (`system-ui, -apple-system, sans-serif`).
- **Cabeçalho:** Marca `CETEP · BACIA DO PARAMIRIM` / `Laboratório de informática` à esquerda e botão `+ Novo agendamento` à direita.
- **Navegação por abas:** `Painel` (`/dashboard`), `Agenda` (`/agenda`), `Cadastros` (`/cadastros`).
- **Rodapé:** `CETEP Bacia do Paramirim · Organização do laboratório de informática`.

### Tela 1 — `Painel` (`VISÃO GERAL`)
- **Título:** `Um laboratório, muitas possibilidades.` | **Subtítulo:** `Acompanhe a utilização e organize suas próximas aulas.`
- **Badge à direita:** `Dados editáveis pelos cadastros e agendamentos`.
- **KPIs (3 colunas):** `Reservas de hoje`, `Horários livres hoje` (13 no total quando vazio), `Próximas reservas`.
- **Grid inferior (2 colunas):**
  - `Próximos agendamentos` (lista ou estado vazio: *"Nenhum agendamento futuro. Cadastre os dados e crie a primeira reserva."*).
  - `Resumo do dia` (data atual + contagem de horários livres por turno):
    - **Manhã (5 slots):** `07:30–08:15`, `08:15–09:00`, `09:15–10:00`, `10:00–10:45`, `10:45–11:00`
    - **Tarde (5 slots):** `13:00–14:00`, `14:00–15:00`, `15:00–16:00`, `16:00–17:00`, `17:00–18:00`
    - **Noite (3 slots):** `19:00–20:00`, `20:00–21:00`, `21:00–22:00`

### Tela 2 — `Agenda` (`COORDENAÇÃO E PROFESSORES`)
- **Título:** `Agenda do laboratório` | **Subtítulo:** `Consulte todas as reservas e encontre períodos disponíveis.`
- **Seletor de visão:** `Dia` | `Semana`.
- **Card de filtros:** Navegação `< Data >` + botão `Ir para hoje`; filtros `Data`, `Período`, `Professor`, `Turma`, `Disciplina` e ação `Limpar filtros`.
- **Grade por turnos (3 colunas pastel):**
  - `Manhã · 07h–12h` (fundo verde pastel `#EBF5E5`)
  - `Tarde · 13h–18h` (fundo âmbar/pêssego pastel `#FDF0DE`)
  - `Noite · 19h–22h` (fundo lavanda pastel `#EAEAF7`)
  - Cada coluna exibe `X reserva(s) · Y horário(s) livre(s)` e chips pontilhados para os slots livres.

### Tela 3 — `Cadastros` (`ORGANIZAÇÃO ESCOLAR`)
- **Título:** `Cadastros` | **Subtítulo:** `Cadastre os nomes que serão usados nos agendamentos. Os dados abaixo são editáveis.`
- **Grid 3 colunas:**
  - **Professores:** campo `Nome do professor` (`Ex.: Ana Souza`) + botão `+` + lista editável.
  - **Turmas:** campo `Identificação da turma` (`Ex.: 2º ano · Informática`) + botão `+` + lista editável.
  - **Disciplinas:** campo `Nome da disciplina` (`Ex.: Programação`) + botão `+` + lista editável.
