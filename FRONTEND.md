# Frontend

## Estrutura de Diretórios

```
frontend/src/
├── main.ts                         # Bootstrap da aplicação (Angular 17 standalone)
├── styles.css                      # Reset CSS global + variáveis CSS para tema light/dark
├── app/
│   ├── app.component.ts/html/css   # Shell raiz (<app-header> + <router-outlet> + <app-error-popup>)
│   ├── app.config.ts               # Providers (HttpClient com errorInterceptor, Router)
│   ├── app.routes.ts               # Definição de rotas (lazy-loaded)
│   ├── interceptors/
│   │   ├── auth.interceptor.ts     # Adiciona header Authorization: Bearer <token>
│   │   └── error.interceptor.ts    # Interceptor HTTP global → ErrorPopupService
│   ├── guards/
│   │   └── auth.guard.ts           # Route guard — redireciona para /login se não autenticado
│   ├── services/                   # Serviços Angular — chamadas à API
│   │   ├── auth.service.ts         # POST /api/auth/login, POST /api/auth/register, handleLoginSuccess, decodeToken
│   │   ├── home.service.ts         # GET /api/albums/featured, /api/artists/featured, /api/albums/recent
│   │   ├── music.service.ts        # GET /api/artists/:id, /api/albums/:id, /api/songs/:id, GET /api/artists/search, GET /api/albums/search, GET /api/albums
│   │   ├── user.service.ts         # GET /users/:id, PUT /users/:id/favorite-artist, DELETE /users/:id/favorite-artist
│   │   ├── profile.service.ts      # GET/PUT /users/:id (perfil do utilizador autenticado)
│   │   ├── collection.service.ts   # GET /api/collections, POST /api/collections, DELETE /api/collections/:id
│   │   ├── request.service.ts      # GET /api/requests, POST /api/requests, POST /api/requests/:id/respond
│   │   ├── notification.service.ts # GET /api/notifications, GET /api/notifications/unread-count, POST mark-all-read, POST /:id/read
│   │   ├── custom_list.service.ts  # GET/POST/DELETE /api/lists, GET/POST/DELETE /api/lists/{id}/items
│   │   ├── theme.service.ts        # Gestão do tema light/dark (persiste em localStorage, atribui [data-theme])
│   │   └── error-popup.service.ts  # Estado do popup de erro (signal-based: visible + message)
│   └── components/                 # Componentes de funcionalidade — adicionar novos componentes aqui
│       ├── header/              # Barra de navegação sticky (brand, nav links, theme toggle, notification bell com unread badge, username + logout)
│       ├── home/                # Página inicial (/, hero + featured albums/artists/recent releases)
│       ├── dashboard/           # Dashboard (/dashboard, barra de pesquisa + cards de navegação)
│       ├── artist-detail/       # Detalhes do artista (/artists/:id, membros + álbuns com capa + botão favorito + link "See all")
│       ├── artist-albums/       # Lista completa de álbuns do artista (/artists/:id/albums, ordenados por ano)
│       ├── album-detail/        # Detalhes do álbum (/albums/:id, capa opcional, tracklist, releases, add-to-collection, add-to-list)
│       ├── song-detail/         # Detalhes da música (/songs/:id)
│       ├── collection/          # Coleção do utilizador (/collection, tabela ordenável desktop + cards mobile, remover com confirmação)
│       ├── lists/               # Listas personalizadas (/lists, formulário inline de criação + cards com contagem, ordenação, remoção)
│       ├── list_detail/         # Detalhe de lista (/lists/:id, álbuns com ordenação, links, remoção)
│       ├── requests/            # Pedidos de versão (/requests, formulário com pesquisa de álbuns + filtro por estado)
│       ├── notifications/       # Notificações (/notifications, cartões aceite/recusado, marcar como lidas)
│       ├── health-check/        # Monitorização do sistema (/health, polls ao backend + lista de rotas)
│       ├── login/               # Formulário de login (/login, redireciona para /dashboard)
│       ├── register/            # Formulário de registo (/register, validação client-side)
│       ├── profile/             # Perfil do utilizador (dados do utilizador autenticado)
│       └── error-popup/         # Modal overlay de erro (warning SVG, animações enter/leave)
└── environments/
    ├── environment.ts              # Configuração de produção (apiUrl)
    └── environment.development.ts   # Configuração de desenvolvimento (substituído via angular.json)
```

### Adicionar uma nova funcionalidade

1. Criar uma pasta de componente em `src/app/components/`
2. Adicionar uma rota lazy em `app.routes.ts`
3. Adicionar um serviço Angular em `src/app/services/` para chamadas à API
4. Usar o `ErrorPopupService` para notificações de erro consistentes

### Padrões de design

- **Serviços**: Um `Injectable` por domínio usando `HttpClient` + `environment.apiUrl`
- **Autenticação**: `AuthService` (login/registo/token) + `auth.interceptor.ts` (Bearer header) + `auth.guard.ts` (proteção de rotas)
- **Erro**: `ErrorPopupService` (signal-based) + `error.interceptor.ts` (intercepta erros HTTP automáticos)
- **Tema**: `ThemeService` (light/dark mode, persiste em `localStorage`, aplica atributo `data-theme` no `<html>`)
- **Estilo**: CSS 100% custom, tema purple gradient com suporte a dark mode, sem frameworks externos
- **Ícones**: Inline SVGs (sem biblioteca de ícones)
- **Componentes**: Standalone (Angular 17+, sem NgModules)
- **Imagens**: Capas de álbuns opcionais via `image_url` (mostra placeholder CSS quando ausente)
