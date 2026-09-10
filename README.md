# Sistema de Gestão e Agendamento de Laboratório Escolar

Sistema web completo para gestão e agendamento de laboratório escolar, desenvolvido em Go (Golang) com arquitetura monolítica e server-side rendered.

## 📋 Índice

- [Objetivo do Projeto](#objetivo-do-projeto)
- [Arquitetura](#arquitetura)
- [Tecnologias](#tecnologias)
- [Estrutura do Projeto](#estrutura-do-projeto)
- [Instalação](#instalação)
- [Configuração](#configuração)
- [Banco de Dados](#banco-de-dados)
- [Execução](#execução)
- [Testes](#testes)
- [Rotas](#rotas)
- [Perfis de Usuário](#perfis-de-usuário)
- [Regras de Negócio](#regras-de-negócio)
- [Processo de Desenvolvimento do Frontend](#processo-de-desenvolvimento-do-frontend)
- [Processo de Integração](#processo-de-integração)
- [Segurança](#segurança)

## 🎯 Objetivo do Projeto

O sistema controla o uso de um laboratório escolar, permitindo:

- Cadastro de usuários
- Login e autenticação
- Controle de acesso
- Visualização de disponibilidade
- Solicitação de agendamento
- Controle de horários
- Prevenção de conflitos
- Aprovação de solicitações
- Rejeição de solicitações
- Cancelamento
- Agenda diária, semanal e mensal
- Histórico
- Notificações
- Auditoria
- Gerenciamento de usuários
- Gerenciamento administrativo

## 🏗️ Arquitetura

O sistema segue uma arquitetura em camadas com separação clara de responsabilidades:

```
Navegador
    │
    ▼
Servidor Go
    │
    ├── HTTP (Handlers)
    ├── Middleware
    ├── Services (Regras de Negócio)
    ├── Repositories (Acesso a Dados)
    ├── Templates HTML
    ├── CSS
    ├── JavaScript
    └── Banco de Dados
```

### Camadas do Backend

1. **Handler**: Recebe requisições, valida entrada básica, chama service, prepara resposta, renderiza template
2. **Service**: Regras de negócio, validações, fluxo de agendamento, aprovação, rejeição, cancelamento
3. **Repository**: SQL, consultas, inserções, atualizações, exclusões

## 🛠️ Tecnologias

- **Go 1.27+**: Linguagem principal
- **net/http**: Servidor HTTP padrão do Go
- **html/template**: Renderização de templates HTML
- **PostgreSQL**: Banco de dados relacional
- **SQL**: Consultas e migrations
- **bcrypt**: Hash de senhas
- **HTML/CSS/JavaScript**: Frontend (esqueletos funcionais)

## 📁 Estrutura do Projeto

```
laboratory_management/
│
├── cmd/
│   └── server/
│       └── main.go              # Ponto de entrada da aplicação
│
├── internal/
│   ├── auth/                    # Autenticação
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── model.go
│   │
│   ├── user/                    # Usuários
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── model.go
│   │
│   ├── booking/                 # Agendamentos
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── model.go
│   │   └── availability.go
│   │
│   ├── laboratory/              # Laboratórios
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── model.go
│   │
│   ├── class/                   # Turmas
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── model.go
│   │
│   ├── notification/            # Notificações
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── model.go
│   │
│   ├── audit/                   # Auditoria
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── model.go
│   │
│   ├── database/                # Conexão com banco
│   │   └── database.go
│   │
│   ├── middleware/              # Middleware HTTP
│   │   ├── auth.go
│   │   ├── logging.go
│   │   ├── csrf.go
│   │   └── session.go
│   │
│   └── web/                     # Rotas e renderização
│       └── router.go
│
├── templates/                   # Templates HTML
│   ├── layouts/
│   │   ├── base.html
│   │   └── auth.html
│   ├── components/
│   │   ├── navbar.html
│   │   ├── sidebar.html
│   │   ├── footer.html
│   │   └── form-field.html
│   └── pages/
│       ├── login.html
│       ├── register.html
│       ├── dashboard.html
│       ├── booking-list.html
│       ├── booking-create.html
│       ├── booking-detail.html
│       ├── availability.html
│       ├── notifications.html
│       ├── profile.html
│       ├── agenda.html
│       ├── history.html
│       ├── admin-dashboard.html
│       ├── admin-users.html
│       ├── admin-bookings.html
│       ├── admin-audit.html
│       ├── 403.html
│       ├── 404.html
│       └── 500.html
│
├── static/                      # Arquivos estáticos
│   ├── css/
│   │   ├── main.css
│   │   └── auth.css
│   ├── js/
│   │   ├── main.js
│   │   └── auth.js
│   ├── img/
│   └── fonts/
│
├── migrations/                  # Migrations do banco
│   └── 001_initial_schema.sql
│
├── tests/                       # Testes
│   ├── booking_conflict_test.go
│   ├── auth_test.go
│   └── validator_test.go
│
├── config/                      # Configuração
│   └── config.go
│
├── .env.example                 # Exemplo de variáveis de ambiente
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## 🚀 Instalação

### Pré-requisitos

- Go 1.27 ou superior
- PostgreSQL 12 ou superior
- Git

### Passos

1. **Clone o repositório**
```bash
git clone <repository-url>
cd laboratory_management
```

2. **Instale as dependências**
```bash
go mod download
```

3. **Configure o banco de dados**
```bash
# Crie o banco de dados PostgreSQL
createdb laboratory_management
```

4. **Configure as variáveis de ambiente**
```bash
cp .env.example .env
# Edite o arquivo .env com suas configurações
```

## ⚙️ Configuração

### Variáveis de Ambiente

Configure as seguintes variáveis no arquivo `.env`:

```env
# Application Environment
APP_ENV=development
APP_PORT=8080

# Database
DATABASE_URL=postgres://user:password@localhost:5432/laboratory_management?sslmode=disable

# Session
SESSION_SECRET=your-secret-key-change-this-in-production

# Security
CSRF_SECRET=your-csrf-secret-change-this-in-production
```

### Configuração de Produção

Para produção:

- Altere `APP_ENV=production`
- Use secrets fortes para `SESSION_SECRET` e `CSRF_SECRET`
- Configure SSL/TLS para PostgreSQL
- Use HTTPS no servidor

## 🗄️ Banco de Dados

### Migrations

Execute as migrations para criar as tabelas:

```bash
# O sistema executa migrations automaticamente na inicialização
# Ou execute manualmente se necessário
psql -U user -d laboratory_management -f migrations/001_initial_schema.sql
```

### Estrutura do Banco

O sistema possui as seguintes tabelas:

- `roles` - Perfis de usuário (PROFESSOR, ALUNO, DIRECAO)
- `users` - Usuários do sistema
- `laboratories` - Laboratórios
- `classes` - Turmas
- `bookings` - Agendamentos
- `notifications` - Notificações
- `audit_logs` - Logs de auditoria

### Índices

Índices foram criados para otimizar consultas em:
- Usuários (email, role, status)
- Agendamentos (user, laboratory, class, date, status)
- Notificações (user, read)
- Auditoria (user, action, entity)

## 🏃 Execução

### Desenvolvimento

```bash
# Inicie o servidor
go run cmd/server/main.go
```

O servidor estará disponível em `http://localhost:8080`

### Produção

```bash
# Compile a aplicação
go build -o laboratory-management cmd/server/main.go

# Execute
./laboratory-management
```

## 🧪 Testes

### Executar Testes

```bash
# Executar todos os testes
go test ./...

# Executar testes com coverage
go test -cover ./...

# Executar testes específicos
go test ./tests/booking_conflict_test.go
```

### Testes Implementados

- **Teste de Conflito de Horários**: Valida a lógica de detecção de conflitos
- **Teste de Autenticação**: Valida hash de senhas e geração de tokens
- **Teste de Validação**: Valida funções de validação de entrada

## 🛣️ Rotas

### Autenticação

- `GET /login` - Página de login
- `POST /login` - Processar login
- `GET /cadastro` - Página de cadastro
- `POST /cadastro` - Processar cadastro
- `GET /logout` - Logout

### Dashboard e Perfil

- `GET /dashboard` - Dashboard do usuário
- `GET /perfil` - Perfil do usuário
- `POST /perfil` - Atualizar perfil

### Agendamentos

- `GET /agendamentos` - Lista de agendamentos do usuário
- `GET /agendamentos/novo` - Formulário de novo agendamento
- `POST /agendamentos` - Criar agendamento
- `GET /agendamentos/{id}` - Detalhes do agendamento
- `POST /agendamentos/cancelar` - Cancelar agendamento

### Disponibilidade

- `GET /disponibilidade` - Verificar disponibilidade
- `GET /disponibilidade/verificar` - API de verificação

### Agenda

- `GET /agenda` - Agenda geral
- `GET /agenda/dia` - Agenda diária
- `GET /agenda/semana` - Agenda semanal
- `GET /agenda/mes` - Agenda mensal

### Histórico

- `GET /historico` - Histórico de agendamentos

### Notificações

- `GET /notificacoes` - Lista de notificações
- `GET /notificacoes/nao-lidas` - Contagem de não lidas
- `POST /notificacoes/marcar-lida` - Marcar como lida
- `POST /notificacoes/marcar-todas-lidas` - Marcar todas como lidas

### Administração (Requer perfil DIRECAO)

- `GET /admin` - Dashboard administrativo
- `GET /admin/usuarios` - Lista de usuários
- `GET /admin/usuarios/editar` - Editar usuário
- `GET /admin/agendamentos` - Lista de agendamentos
- `GET /admin/agendamentos/pendentes` - Solicitações pendentes
- `POST /admin/agendamentos/aprovar` - Aprovar agendamento
- `POST /admin/agendamentos/rejeitar` - Rejeitar agendamento
- `POST /admin/agendamentos/cancelar` - Cancelar agendamento
- `GET /admin/historico` - Histórico administrativo
- `GET /admin/auditoria` - Logs de auditoria
- `GET /admin/laboratorios` - Gerenciar laboratórios
- `GET /admin/turmas` - Gerenciar turmas

### Erros

- `GET /403` - Acesso negado
- `GET /404` - Página não encontrada
- `GET /500` - Erro interno do servidor

## 👥 Perfis de Usuário

### PROFESSOR

- Criar cadastro
- Fazer login
- Visualizar dashboard
- Visualizar disponibilidade
- Solicitar laboratório
- Visualizar seus agendamentos
- Visualizar detalhes
- Cancelar quando permitido
- Consultar histórico
- Visualizar notificações
- Consultar agenda

### ALUNO

- Criar cadastro
- Fazer login
- Visualizar dashboard
- Visualizar disponibilidade
- Solicitar laboratório quando permitido
- Visualizar seus agendamentos
- Acompanhar solicitações
- Consultar histórico
- Receber notificações
- Consultar agenda

### DIREÇÃO (Administração)

- Todas as funcionalidades de professor
- Visualizar dashboard administrativo
- Visualizar usuários
- Gerenciar usuários
- Visualizar solicitações
- Aprovar
- Rejeitar
- Cancelar
- Visualizar agenda
- Visualizar histórico
- Visualizar auditoria
- Consultar disponibilidade

## 📋 Regras de Negócio

### Conflito de Horários

O sistema nunca permite que dois agendamentos ocupem o mesmo laboratório em horários conflitantes.

**Exemplos de conflito:**
- 14:00 - 16:00 vs 15:00 - 17:00 (conflito)
- 09:00 - 11:00 vs 10:30 - 12:00 (conflito)

**Exemplos sem conflito:**
- 09:00 - 10:00 vs 10:00 - 11:00 (permitido)

A validação considera sobreposição de intervalos e é executada no backend.

### Fluxo de Solicitação

1. Usuário escolhe data e horário
2. Escolhe turma (opcional)
3. Informa motivo
4. Envia solicitação
5. Status: PENDENTE

### Fluxo de Aprovação

1. Direção analisa solicitação
2. Verifica conflito novamente (antes de aprovar)
3. Se não houver conflito: APROVADO
4. Se houver conflito: não aprova
5. Registra responsável e data
6. Gera notificação

### Fluxo de Rejeição

1. Direção analisa solicitação
2. Define motivo da rejeição
3. Status: REJEITADO
4. Registra responsável, data e motivo
5. Gera notificação

### Cancelamento

Nunca apaga o agendamento. Altera para CANCELADO e registra:
- Quem cancelou
- Quando
- Motivo

## 🎨 Processo de Desenvolvimento do Frontend

### Responsabilidades

**Backend (IA/Desenvolvedores):**
- Go
- Banco de dados
- Regras de negócio
- Autenticação
- Autorização
- Handlers
- Services
- Repositories
- Sessões
- Notificações
- Auditoria

**Frontend (Alunos):**
- HTML
- CSS
- JavaScript
- UI/UX
- Responsividade
- Calendário
- Componentes
- Identidade visual

### Esqueletos HTML Funcionais

O sistema fornece esqueletos HTML funcionais para todas as páginas:

**Estrutura fornecida:**
- Estrutura HTML
- Títulos
- Seções
- Formulários
- Campos
- Botões
- Tabelas
- Áreas para mensagens
- Placeholders
- Dados esperados do backend
- Navegação básica
- Classes CSS organizadas

**Não desenvolvido:**
- Design visual final
- Identidade visual
- Cores
- Tipografia
- Espaçamento
- Responsividade avançada
- Componentes visuais elaborados
- Animações
- UX/UI refinado

## 🔗 Processo de Integração

### Quando os alunos entregarem o frontend:

1. **Análise do frontend existente**
   - Identificar todas as páginas
   - Identificar formulários
   - Identificar campos
   - Identificar botões e ações

2. **Mapeamento**
   - Mapear páginas para rotas
   - Mapear formulários para handlers
   - Conectar dados aos templates
   - Conectar banco
   - Conectar autenticação
   - Conectar autorização
   - Conectar notificações

3. **Integração**
   - Preservar o design dos alunos
   - Adaptar apenas o necessário para funcionalidade
   - Não substituir o frontend dos alunos
   - Trabalhar sobre o frontend existente

### Exemplo de Integração

```
login.html (dos alunos)
    ↓
POST /login
    ↓
AuthHandler
    ↓
AuthService
    ↓
UserRepository
    ↓
Database
```

## 🔒 Segurança

### Implementações de Segurança

- ✅ Hash seguro de senha (bcrypt)
- ✅ Sessões seguras
- ✅ Cookies adequados (HttpOnly, Secure em produção)
- ✅ Proteção CSRF
- ✅ Proteção contra XSS
- ✅ Proteção contra SQL Injection (usando prepared statements)
- ✅ Validação server-side
- ✅ Autorização baseada em perfis
- ✅ Controle de acesso
- ✅ Tratamento seguro de erros
- ✅ Variáveis de ambiente
- ✅ Não exposição de credenciais

### Regras de Segurança

- Nunca armazenar senha em texto puro
- Nunca confiar no frontend para autenticação ou autorização
- Nunca expor credenciais no código
- Sempre validar dados no servidor
- Usar transações para operações críticas

## 📝 Notas Importantes

### Para Alunos

1. **Não altere** a estrutura HTML básica fornecida
2. **Foque** no design visual e na experiência do usuário
3. **Use** as classes CSS fornecidas como base
4. **Adicione** sua identidade visual, cores, tipografia
5. **Melhore** a responsividade e UX
6. **Crie** componentes visuais elaborados
7. **Implemente** o calendário visual
8. **Adicione** animações e transições

### Para Desenvolvedores

1. **Preserve** o design desenvolvido pelos alunos
2. **Adapte** apenas o necessário para funcionalidade
3. **Não substitua** o frontend dos alunos
4. **Teste** a integração cuidadosamente
5. **Documente** as mudanças necessárias

## 🤝 Contribuição

Este projeto foi desenvolvido como um sistema educacional para ensinar:

- Desenvolvimento web em Go
- Arquitetura em camadas
- Design de banco de dados
- Segurança web
- Colaboração entre backend e frontend

## 📄 Licença

Este projeto foi desenvolvido para fins educacionais.

## 📞 Suporte

Para suporte, consulte a documentação ou entre em contato com os instrutores.

---

**Desenvolvido com 🖤 para fins educacionais**