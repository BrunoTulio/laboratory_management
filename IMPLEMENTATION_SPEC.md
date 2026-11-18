# Spec de Implementação — Conclusão do Sistema (CETEP Laboratório)

Este documento especifica as entregas restantes para tornar o fluxo principal do sistema completo no navegador, seguro e coberto por testes automatizados.

---

## 1. Autenticação HTML (`/login` e `/cadastro`) e Contexto Real
- **Rotas:** `GET|POST /login`, `GET|POST /cadastro`
- **Requisitos:**
  - `GET /login` e `GET /cadastro` devem renderizar páginas HTML (`login.html` / `register.html`) com o layout de autenticação (`auth.html` / `auth.css`) e `CSRFToken`.
  - `POST /login` e `POST /cadastro` devem aceitar tanto `application/x-www-form-urlencoded` (formulário HTML padrão) quanto `application/json` (API).
  - Em `POST /login` via formulário HTML: autenticar via `authService.Login`, definir cookie `session_token` assinado com HMAC-SHA256 e redirecionar (`303 See Other`) para `/dashboard`. Em erro, re-renderizar `login.html` com mensagem de erro.
  - Em `POST /cadastro` via formulário HTML: validar campos, hashear senha, criar usuário (`ALUNO` ou `PROFESSOR` conforme selecionado, bloqueando `DIRECAO`), e redirecionar para `/login`.
  - `getUserFromContext` em `internal/web/router.go` deve usar `middleware.GetUserIDFromContext`, `middleware.GetUserEmailFromContext` e `middleware.GetUserRoleFromContext`.

## 2. Visão Semanal da Agenda (`/agenda/semana`)
- **Rota:** `GET /agenda/semana`
- **Requisitos:**
  - Calcular o intervalo da semana (segunda a sábado) contendo a data alvo (`?date=YYYY-MM-DD`, padrão hoje).
  - Para cada dia da semana, calcular o resumo dos 3 turnos (`Manhã`, `Tarde`, `Noite`) com contagem de reservas e horários livres, aplicando os filtros ativos (`shift`, `teacher`, `class_id`, `subject`).
  - Navegação para semana anterior (`-7 dias`), próxima semana (`+7 dias`) e `Ir para hoje`.
  - Renderizar a visão semanal em `templates/pages/agenda.html` mantendo a alternância `Dia | Semana`.

## 3. Edição em `Cadastros` (`/cadastros`) e Filtro de Turma na Agenda
- **Repositório (`internal/registry/repository.go`):**
  - Adicionar `UpdateTeacher(id int, name string) error`, `UpdateClass(id int, name string) error` e `UpdateSubject(id int, name string) error`.
- **Handler (`internal/web/router.go` — `cadastrosHandler`):**
  - Suportar `action=update` no `POST /cadastros` além de criação e remoção (`action=delete`).
- **Templates:**
  - Em `templates/pages/cadastros.html`: permitir editar inline o nome de Professores, Turmas e Disciplinas.
  - Em `templates/pages/agenda.html`: marcar `selected` no filtro `Turma` (`filter_class`) quando selecionado.

## 4. Validação Estrita de CSRF (`internal/middleware/csrf.go`)
- **Requisitos:**
  - Substituir a checagem apenas de não-vazio por validação real contra uma instância compartilhada de `CSRFMiddleware` (`GenerateToken` / `ValidateToken`).
  - Métodos seguros (`GET`, `HEAD`, `OPTIONS`, `TRACE`) seguem liberados.
  - Métodos não seguros (`POST`, `PUT`, `DELETE`, `PATCH`) sem token ou com token inválido/desconhecido devem retornar `403 Forbidden`.
  - Templates que enviam `POST` devem receber um token gerado via `middleware.GenerateCSRFToken()`.

## 5. Plano de Validação por Testes (`tests/`)
A tarefa só será considerada concluída após todos os testes abaixo passarem via `asdf exec go test ./...` e `asdf exec go build ./...`:
1. **`tests/middleware/middleware_test.go`:**
   - `SignSessionToken` e `VerifySessionToken`: token válido, token adulterado, token expirado, formato inválido.
   - `AuthMiddleware`: redireciona para `/login` sem cookie ou com cookie inválido; injeta `userID`, `userEmail` e `userRole` no context quando válido.
   - `CSRF`: permite `GET`; bloqueia `POST` sem token (`403`); bloqueia `POST` com token forjado (`403`); permite `POST` com token gerado por `GenerateCSRFToken()`.
2. **`tests/web/schedule_test.go`:**
   - `buildShiftSummaries`: valida contagem inicial de `13` slots livres (`5` manhã, `5` tarde, `3` noite), marcação de slot ocupado quando há reserva sobreposta, e filtros por `shift`, `teacher`, `classID` e `subject`.
   - `buildWeekDays`: valida geração dos 6 dias úteis (segunda a sábado) a partir de qualquer data da semana.
   - `formatDatePTBR`: valida formatação em português (`ter., 17 de nov. de 2026`).
3. **Regressão das suítes existentes:**
   - `tests/auth`, `tests/booking`, `tests/validator` continuam passando 100%.
