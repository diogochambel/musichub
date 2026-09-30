# Backend

## Estrutura de Diretórios

```
backend/
├── cmd/
│   ├── server/main.go              # Ponto de entrada (config, DB, rotas, CORS, graceful shutdown)
│   └── populate/main.go            # Script de seed/população da base de dados (clear collections + insert dados de teste)
├── internal/
│   ├── config/config.go            # Configuração por variáveis de ambiente (MongoURI, DBName, ServerPort, FrontendURL)
│   ├── db/mongo.go                 # Conexão MongoDB (Connect, Ping, helper Collection)
│   ├── handlers/                   # HTTP handlers — adicionar novos handlers aqui
│   │   ├── health.go               # /health, /health/db
│   │   ├── user.go                 # CRUD /users + artista favorito (define UserRepository, ArtistFinder)
│   │   ├── auth.go                 # POST /api/auth/login, POST /api/auth/register
│   │   ├── home.go                 # /api/albums/featured, /api/artists/featured, /api/albums/recent
│   │   ├── artist.go               # GET /api/artists/{id}, GET /api/artists/search
│   │   ├── album.go                # GET /api/albums/{id}, GET /api/albums/search
│   │   ├── song.go                 # GET /api/songs/{id}
│   │   ├── collection.go           # GET/POST /api/collections, DELETE /api/collections/{id}
│   │   ├── request.go              # GET/POST /api/requests, POST /api/requests/{id}/respond + notification endpoints
│   │   └── custom_list.go          # GET/POST /api/lists, DELETE /api/lists/{id}, GET/POST /api/lists/{id}/items, DELETE /api/lists/{id}/items/{albumId}
│   ├── models/                     # Modelos de dados e validação — adicionar novos structs aqui
│   │   ├── user.go                 # Struct User (BirthDate + FavoriteArtistID), CreateUserRequest, UpdateUserRequest, SetFavoriteArtistRequest
│   │   ├── artist.go               # Struct Artist, CreateArtistRequest, UpdateArtistRequest
│   │   ├── album.go                # Struct Album (ImageURL opcional), TrackEntry, AlbumRelease, CreateAlbumRequest, UpdateAlbumRequest
│   │   ├── song.go                 # Struct Song, CreateSongRequest, UpdateSongRequest
│   │   ├── collection.go           # Struct CollectionItem, CreateCollectionRequest
│   │   ├── version_request.go      # Struct VersionRequest, RequestStatus enum ("em análise" | "aceite" | "recusado")
│   │   ├── notification.go         # Struct Notification (ligada a um VersionRequest)
│   │   ├── custom_list.go          # Struct CustomList, CreateCustomListRequest
│   │   └── custom_list_item.go     # Struct CustomListItem
│   ├── repository/                 # Acesso à base de dados — adicionar novas implementações de repo aqui
│   │   ├── user.go                 # MongoUserRepository (CRUD + indexes únicos + SetFavoriteArtistID + RemoveFavoriteArtistID)
│   │   ├── artist.go               # MongoArtistRepository (CRUD + SearchByName + FindRandom + FindByISNI)
│   │   ├── album.go                # MongoAlbumRepository (CRUD + FindByArtistID ordenado por release_year desc + FindRandom + FindRecent + FindByMBID + FindAll)
│   │   ├── song.go                 # MongoSongRepository (CRUD + FindByArtistID + FindRandom + FindByISRC)
│   │   ├── collection.go           # MongoCollectionRepository (Insert, FindByUserID, FindByUserAndAlbumAndEAN, Delete)
│   │   ├── version_request.go      # MongoVersionRequestRepository (Insert, FindByID, FindByUserID, FindByUserIDAndStatus, UpdateStatus)
│   │   ├── notification.go         # MongoNotificationRepository (Insert, FindByUserID, CountUnread, MarkAllRead, MarkRead, Delete)
│   │   ├── custom_list.go          # MongoCustomListRepository (Insert, FindByID, FindByUserID, FindByUserAndName, Delete, Update)
│   │   └── custom_list_item.go     # MongoCustomListItemRepository (Insert, FindByListID, CountByListID, DeleteByListID, DeleteByListIDAndAlbumID)
│   └── middleware/                  # Middleware HTTP
│       ├── auth.go                 # CreateToken, VerifyToken, RequireAuth middleware (HMAC-SHA256, hmac.Equal)
│       └── cors.go                 # CORS middleware (origens permitidas via FRONTEND_URL)
├── go.mod
└── go.sum
```

### Adicionar um novo recurso

1. Definir o modelo em `internal/models/`
2. Criar um repositório em `internal/repository/`
3. Adicionar handlers em `internal/handlers/`
4. Registar rotas em `cmd/server/main.go`

## Rotas da API

### Públicas (sem autenticação)

| Método | Rota | Handler | Descrição |
|--------|------|---------|-----------|
| GET | `/health` | HealthHandler.Health | Estado do backend |
| GET | `/health/db` | HealthHandler.DBHealth | Estado da ligação MongoDB |
| POST | `/api/auth/login` | AuthHandler.Login | Autenticação |
| POST | `/api/auth/register` | AuthHandler.Register | Registo de novo utilizador |
| GET | `/api/albums/featured` | HomeHandler.FeaturedAlbums | Álbuns em destaque |
| GET | `/api/artists/featured` | HomeHandler.FeaturedArtists | Artistas em destaque |
| GET | `/api/albums/recent` | HomeHandler.RecentReleases | Lançamentos recentes |
| GET | `/api/albums/search` | AlbumHandler.SearchAlbums | Pesquisa parcial por título de álbum |
| GET | `/api/artists/{id}` | ArtistHandler.GetArtist | Detalhes do artista |
| GET | `/api/artists/search` | ArtistHandler.SearchArtists | Pesquisa parcial por nome de artista |
| GET | `/api/albums/{id}` | AlbumHandler.GetAlbum | Detalhes do álbum |
| GET | `/api/songs/{id}` | SongHandler.GetSong | Detalhes da música |

### Protegidas (requerem Bearer token via `RequireAuth` middleware)

| Método | Rota | Handler | Descrição |
|--------|------|---------|-----------|
| POST | `/users` | UserHandler.CreateUser | Criar utilizador |
| GET | `/users` | UserHandler.ListUsers | Listar utilizadores |
| GET | `/users/{id}` | UserHandler.GetUser | Obter utilizador por ID |
| PUT | `/users/{id}` | UserHandler.UpdateUser | Atualizar utilizador |
| DELETE | `/users/{id}` | UserHandler.DeleteUser | Eliminar utilizador |
| PUT | `/users/{id}/favorite-artist` | UserHandler.SetFavoriteArtist | Definir artista favorito |
| DELETE | `/users/{id}/favorite-artist` | UserHandler.RemoveFavoriteArtist | Remover artista favorito |
| GET | `/api/collections` | CollectionHandler.HandleCollections | Listar coleção do utilizador |
| POST | `/api/collections` | CollectionHandler.HandleCollections | Adicionar versão à coleção |
| DELETE | `/api/collections/{id}` | CollectionHandler.HandleCollectionItem | Remover item da coleção |
| GET | `/api/requests` | RequestHandler.ListRequests | Listar pedidos de versão |
| POST | `/api/requests` | RequestHandler.CreateRequest | Submeter novo pedido de versão |
| POST | `/api/requests/{id}/respond` | RequestHandler.RespondToRequest | Responder a pedido (aceitar/rejeitar) + criar notificação |
| GET | `/api/notifications` | RequestHandler.ListNotifications | Listar notificações do utilizador |
| GET | `/api/notifications/unread-count` | RequestHandler.CountUnreadNotifications | Contar notificações não lidas |
| POST | `/api/notifications/mark-all-read` | RequestHandler.MarkAllNotificationsRead | Marcar todas como lidas |
| POST | `/api/notifications/{id}/read` | RequestHandler.MarkNotificationRead | Marcar notificação como lida |
| GET | `/api/lists` | CustomListHandler.HandleCustomLists | Listar listas personalizadas do utilizador |
| POST | `/api/lists` | CustomListHandler.HandleCustomLists | Criar lista personalizada |
| DELETE | `/api/lists/{id}` | CustomListHandler.HandleCustomListItem | Eliminar lista personalizada |
| GET | `/api/lists/{id}/items` | CustomListHandler.HandleCustomListItem | Listar álbuns de uma lista |
| POST | `/api/lists/{id}/items` | CustomListHandler.HandleCustomListItem | Adicionar álbum a uma lista |
| DELETE | `/api/lists/{id}/items/{albumId}` | CustomListHandler.HandleCustomListItem | Remover álbum de uma lista |

## Artista Favorito

O campo `FavoriteArtistID` no modelo `User` é um ponteiro (`*primitive.ObjectID`), permitindo distinguir entre "sem favorito" (`nil`) e "favorito definido".

### PUT /users/{id}/favorite-artist

**Body:**
```json
{ "artist_id": "661234567890abcdef123456" }
```

**Respostas:**
- `200 OK` — Artista favorito definido com sucesso, devolve o utilizador atualizado
- `400 Bad Request` — `artist_id` inválido ou formato incorreto
- `404 Not Found` — Utilizador ou artista não encontrado
- `409 Conflict` — Utilizador já tem outro artista favorito

### DELETE /users/{id}/favorite-artist

**Respostas:**
- `200 OK` — Artista favorito removido, devolve o utilizador atualizado
- `400 Bad Request` — ID do utilizador inválido
- `404 Not Found` — Utilizador não encontrado

### GET /users/{id}

O campo `favorite_artist_id` aparece no JSON apenas quando definido (tag `omitempty`).

## Coleções (US10–US11)

Cada item da coleção é único por combinação de `user_id + album_id + ean_13` (índice composto unique no MongoDB).

## Listas Personalizadas (US15–US18)

Cada lista personalizada é única por combinação de `user_id + name` (índice composto unique no MongoDB, limite de 100 carateres para o nome).

### Operações disponíveis

| Operação | Rota | Descrição |
|----------|------|-----------|
| Criar lista | `POST /api/lists` | Cria uma nova lista com nome único por utilizador |
| Listar listas | `GET /api/lists` | Devolve listas com contagem de álbuns e data de modificação |
| Eliminar lista | `DELETE /api/lists/{id}` | Remove a lista e todos os seus items |
| Listar items | `GET /api/lists/{id}/items` | Devolve álbuns da lista com título e data de adição |
| Adicionar álbum | `POST /api/lists/{id}/items` | Adiciona álbum à lista (previne duplicados, atualiza timestamp) |
| Remover álbum | `DELETE /api/lists/{id}/items/{albumId}` | Remove álbum da lista |

## Pedidos de Versão (US12–US14)

Quando um pedido é respondido com `"aceite"`, a nova versão é automaticamente adicionada ao array `releases` do álbum correspondente e uma notificação é criada para o utilizador que submeteu o pedido.
