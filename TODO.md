## User Stories (Sprint 3)

### US15 — Criar Lista Personalizada
- [x] **Backend:** Modelo `CustomList`, repositório, handler `POST /api/lists` (nome único por utilizador, limite 100 carateres), índices MongoDB
- [x] **Frontend:** `ListService`, modal/formulário para criar lista com validação de nome e limite de carateres

### US16 — Visualizar e Gerir Listas Personalizadas
- [x] **Backend:** Handler `GET /api/lists` (com contagem de álbuns e data de última modificação), `DELETE /api/lists/:id`, ordenação por qualquer campo
- [x] **Frontend:** Página `/lists` com cards (nome, nº de álbuns, última modificação), ordenação, estado vazio, remoção com confirmação

### US17 — Adicionar Álbum a uma Lista
- [x] **Backend:** Modelo `CustomListItem`, handler `POST /api/lists/:id/items` (prevenir duplicados), atualização da data de modificação da lista
- [x] **Frontend:** Modal na página do álbum com seletor de listas, mensagens de sucesso/erro, validação de duplicados inline

### US18 — Visualizar Conteúdo de uma Lista
- [x] **Backend:** Handler `GET /api/lists/:id/items` (com títulos dos álbuns e data de adição), `DELETE /api/lists/:id/items/:albumId`
- [x] **Frontend:** Página `/lists/:id` com nome da lista, álbuns (título, data de adição), ordenação, links para os álbuns, estado vazio, remoção de álbuns

### US19 — Dashboard do Utilizador
- [x] **Backend:** Endpoints necessários (coleção, listas, pedidos, pesquisa) estáveis
- [x] **Frontend:** `DashboardComponent` com barra de pesquisa e cards de navegação; rota `/dashboard` com `authGuard`

## Distribuição de Tarefas

| Membro | Responsabilidades |
|--------|-------------------|
| **Dário** | **US15** — Criar Lista Personalizada (backend + frontend); **Refinamentos** — botão uniforme, responsividade coleção, header, merge conflicts, interceptor, documentação |
| **Diogo Chambel** | **US16** — Visualizar e Gerir Listas (backend: listagem, ordenação, remoção + frontend: página `/lists` com ordenação, empty state) |
| **Daniel Horta** | **US17 + US18** — Conteúdo das Listas (backend: itens, adicionar/remover álbuns, duplicados + frontend: modal no álbum, página `/lists/:id` com detalhes); **US14** — Notificações (backend + frontend) |
| **Rafael** | **US19** — Dashboard (frontend: `DashboardComponent`, barra de pesquisa, cards de navegação, rotas) |
