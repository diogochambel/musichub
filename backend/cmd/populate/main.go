package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dariolbs/PSI/internal/config"
	"github.com/dariolbs/PSI/internal/db"
	"github.com/dariolbs/PSI/internal/models"
	"github.com/dariolbs/PSI/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	cfg := config.Load()

	log.Printf("Connecting to MongoDB at %s (db: %s)", cfg.MongoURI, cfg.DBName)
	mongodb, err := db.NewMongoDB(cfg.MongoURI, cfg.DBName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongodb.Close()

	ctx := context.Background()

	clearCollections(ctx, mongodb)

	userRepo := repository.NewMongoUserRepository(mongodb)
	artistRepo := repository.NewMongoArtistRepository(mongodb)
	songRepo := repository.NewMongoSongRepository(mongodb)
	albumRepo := repository.NewMongoAlbumRepository(mongodb)
	collectionRepo := repository.NewMongoCollectionRepository(mongodb)
	notificationRepo := repository.NewMongoNotificationRepository(mongodb)
	versionRequestRepo := repository.NewMongoVersionRequestRepository(mongodb)
	customListRepo := repository.NewMongoCustomListRepository(mongodb)
	customListItemRepo := repository.NewMongoCustomListItemRepository(mongodb)

	userIDs := seedUsers(ctx, userRepo)
	artistIDs := seedArtists(ctx, artistRepo)
	songIDs := seedSongs(ctx, songRepo, artistIDs)
	albumIDs := seedAlbums(ctx, albumRepo, artistIDs, songIDs)
	seedVersionRequests(ctx, versionRequestRepo, userIDs, albumIDs)
	seedCollections(ctx, collectionRepo, userIDs, albumIDs)
	seedNotifications(ctx, notificationRepo, userIDs)
	seedCustomLists(ctx, customListRepo, customListItemRepo, userIDs, albumIDs)
}

func clearCollections(ctx context.Context, mongodb *db.MongoDB) {
	for _, name := range []string{"users", "artists", "songs", "albums", "user_collections", "version_requests", "notifications", "custom_lists", "custom_list_items"} {
		_, err := mongodb.Collection(name).DeleteMany(ctx, bson.D{})
		if err != nil {
			log.Fatalf("Failed to clear collection %s: %v", name, err)
		}
	}
	log.Println("Cleared all collections")
}

func seedUsers(ctx context.Context, repo *repository.MongoUserRepository) map[string]primitive.ObjectID {
	ids := map[string]primitive.ObjectID{
		"pedro": insertableObjectID(1),
		"maria": insertableObjectID(2),
		"joao":  insertableObjectID(3),
	}

	type userData struct {
		username  string
		email     string
		password  string
		birthDate time.Time
	}

	users := []userData{
		{"pedro", "pedro@example.com", "Password1", time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"maria", "maria@example.com", "Password2", time.Date(2002, 5, 12, 0, 0, 0, 0, time.UTC)},
		{"joao", "joao@example.com", "Password3", time.Date(1994, 9, 20, 0, 0, 0, 0, time.UTC)},
	}

	count := 0
	for _, u := range users {
		salt, err := models.GenerateSalt()
		if err != nil {
			log.Fatalf("Failed to generate salt for user %s: %v", u.username, err)
		}

		user := &models.User{
			ID:        ids[u.username],
			Username:  u.username,
			Email:     u.email,
			Password:  models.HashPassword(u.password, salt),
			Salt:      salt,
			BirthDate: u.birthDate,
		}

		if err := repo.Insert(ctx, user); err != nil {
			log.Printf("Failed to insert user %s: %v", u.username, err)
			continue
		}
		count++
		log.Printf("Inserted user: %s", u.username)
	}
	printSection("Users", count, 3)
	return ids
}

func seedArtists(ctx context.Context, repo *repository.MongoArtistRepository) map[string]primitive.ObjectID {
	ids := map[string]primitive.ObjectID{
		"papa_roach":    insertableObjectID(10),
		"ed_sheeran":    insertableObjectID(11),
		"lady_gaga":     insertableObjectID(12),
		"jacob_shaddix": insertableObjectID(13),
		"jerry_horton":  insertableObjectID(14),
	}

	artists := []models.Artist{
		{
			ID: ids["papa_roach"], ISNI: "0000000123456789", Name: "Papa Roach",
			StartYear: 1993, Type: models.ArtistTypeGroup,
			MemberIDs: []primitive.ObjectID{ids["jacob_shaddix"], ids["jerry_horton"]},
		},
		{
			ID: ids["ed_sheeran"], ISNI: "0000000123456790", Name: "Ed Sheeran",
			StartYear: 2004, Type: models.ArtistTypeSolo,
		},
		{
			ID: ids["lady_gaga"], ISNI: "0000000123456791", Name: "Lady Gaga",
			StartYear: 2001, Type: models.ArtistTypeSolo,
		},
		{
			ID: ids["jacob_shaddix"], ISNI: "0000000123456792", Name: "Jacob Shaddix",
			StartYear: 1993, Type: models.ArtistTypeSolo,
		},
		{
			ID: ids["jerry_horton"], ISNI: "0000000123456793", Name: "Jerry Horton",
			StartYear: 1993, Type: models.ArtistTypeSolo,
		},
	}

	count := 0
	for _, a := range artists {
		if err := repo.Insert(ctx, &a); err != nil {
			log.Printf("Failed to insert artist %s: %v", a.Name, err)
			continue
		}
		count++
		log.Printf("Inserted artist: %s (%s)", a.Name, a.Type)
	}
	printSection("Artists", count, 5)
	return ids
}

func seedSongs(ctx context.Context, repo *repository.MongoSongRepository, artistIDs map[string]primitive.ObjectID) map[string]primitive.ObjectID {
	ids := map[string]primitive.ObjectID{
		"last_resort":            insertableObjectID(20),
		"broken_home":            insertableObjectID(21),
		"between_angels_insects": insertableObjectID(22),
		"infest_title":           insertableObjectID(23),
		"shape_of_you":           insertableObjectID(24),
		"perfect":                insertableObjectID(25),
		"bad_romance":            insertableObjectID(26),
		"poker_face":             insertableObjectID(27),
		"born_survivor":          insertableObjectID(28),
		"the_a_team":             insertableObjectID(40),
		"lego_house":             insertableObjectID(41),
		"give_me_love":           insertableObjectID(42),
		"castle_on_hill":         insertableObjectID(43),
		"photograph":             insertableObjectID(44),
		"thinking_out_loud":      insertableObjectID(45),
		"galway_girl":            insertableObjectID(46),
		"happier":                insertableObjectID(47),
		"overpass_graffiti":      insertableObjectID(48),
		"shivers":                insertableObjectID(49),
		"bad_habits":             insertableObjectID(50),
		"eyes_closed":            insertableObjectID(51),
		"azizam":                 insertableObjectID(52),
	}

	type songData struct {
		song         models.Song
		artistIDKeys []string
	}

	songs := []songData{
		{models.Song{ID: ids["last_resort"], ISRC: "USUM72000001", Title: "Last Resort", DurationSeconds: 240}, []string{"papa_roach"}},
		{models.Song{ID: ids["broken_home"], ISRC: "USUM72000002", Title: "Broken Home", DurationSeconds: 228}, []string{"papa_roach"}},
		{models.Song{ID: ids["between_angels_insects"], ISRC: "USUM72000003", Title: "Between Angels and Insects", DurationSeconds: 216}, []string{"papa_roach"}},
		{models.Song{ID: ids["infest_title"], ISRC: "USUM72000004", Title: "Infest", DurationSeconds: 252}, []string{"papa_roach"}},
		{models.Song{ID: ids["shape_of_you"], ISRC: "GBDJH72000001", Title: "Shape of You", DurationSeconds: 234}, []string{"ed_sheeran"}},
		{models.Song{ID: ids["perfect"], ISRC: "GBDJH72000002", Title: "Perfect", DurationSeconds: 263}, []string{"ed_sheeran"}},
		{models.Song{ID: ids["bad_romance"], ISRC: "USUM70900001", Title: "Bad Romance", DurationSeconds: 295}, []string{"lady_gaga"}},
		{models.Song{ID: ids["poker_face"], ISRC: "USUM70900002", Title: "Poker Face", DurationSeconds: 237}, []string{"lady_gaga"}},
		{models.Song{ID: ids["born_survivor"], ISRC: "USUM72000005", Title: "Born Survivor", DurationSeconds: 198}, []string{"papa_roach"}},

		{models.Song{ID: ids["the_a_team"], ISRC: "GBDJH11000001", Title: "The A Team", DurationSeconds: 256}, []string{"ed_sheeran"}},
		{models.Song{ID: ids["lego_house"], ISRC: "GBDJH11000002", Title: "Lego House", DurationSeconds: 221}, []string{"ed_sheeran"}},
		{models.Song{ID: ids["give_me_love"], ISRC: "GBDJH11000003", Title: "Give Me Love", DurationSeconds: 263}, []string{"ed_sheeran"}},
		{models.Song{ID: ids["castle_on_hill"], ISRC: "GBDJH17000001", Title: "Castle on the Hill", DurationSeconds: 263}, []string{"ed_sheeran"}},
		{models.Song{ID: ids["photograph"], ISRC: "GBDJH14000001", Title: "Photograph", DurationSeconds: 258}, []string{"ed_sheeran"}},
		{models.Song{ID: ids["thinking_out_loud"], ISRC: "GBDJH14000002", Title: "Thinking Out Loud", DurationSeconds: 281}, []string{"ed_sheeran"}},
		{models.Song{ID: ids["galway_girl"], ISRC: "GBDJH17000002", Title: "Galway Girl", DurationSeconds: 182}, []string{"ed_sheeran"}},
		{models.Song{ID: ids["happier"], ISRC: "GBDJH17000003", Title: "Happier", DurationSeconds: 223}, []string{"ed_sheeran"}},
		{models.Song{ID: ids["overpass_graffiti"], ISRC: "GBDJH21000001", Title: "Overpass Graffiti", DurationSeconds: 238}, []string{"ed_sheeran"}},
		{models.Song{ID: ids["shivers"], ISRC: "GBDJH21000002", Title: "Shivers", DurationSeconds: 222}, []string{"ed_sheeran"}},
		{models.Song{ID: ids["bad_habits"], ISRC: "GBDJH21000003", Title: "Bad Habits", DurationSeconds: 231}, []string{"ed_sheeran"}},
		{models.Song{ID: ids["eyes_closed"], ISRC: "GBDJH23000001", Title: "Eyes Closed", DurationSeconds: 220}, []string{"ed_sheeran"}},
		{models.Song{ID: ids["azizam"], ISRC: "GBDJH25000001", Title: "Azizam", DurationSeconds: 210}, []string{"ed_sheeran"}},
	}

	count := 0
	for _, s := range songs {
		s.song.ArtistIDs = resolveIDs(s.artistIDKeys, artistIDs)
		if err := repo.Insert(ctx, &s.song); err != nil {
			log.Printf("Failed to insert song %s: %v", s.song.Title, err)
			continue
		}
		count++
		log.Printf("Inserted song: %s", s.song.Title)
	}
	printSection("Songs", count, 21)
	return ids
}

func seedAlbums(ctx context.Context, repo *repository.MongoAlbumRepository, artistIDs map[string]primitive.ObjectID, songIDs map[string]primitive.ObjectID) []primitive.ObjectID {
	albums := []models.Album{
		{
			ID: insertableObjectID(30), MBID: "a3b4c5d6-1111-2222-3333-444455556666",
			Title: "Infest", ReleaseYear: 2000, Type: models.AlbumTypeLP,
			ArtistID: pObjectID(artistIDs["papa_roach"]),
			Tracks: []models.TrackEntry{
				{TrackNumber: 1, SongID: songIDs["infest_title"]},
				{TrackNumber: 2, SongID: songIDs["last_resort"]},
				{TrackNumber: 3, SongID: songIDs["broken_home"]},
				{TrackNumber: 4, SongID: songIDs["between_angels_insects"]},
				{TrackNumber: 5, SongID: songIDs["born_survivor"]},
			},
			Releases: []models.AlbumRelease{
				{Support: models.SupportTypeCD, VersionName: "Standard Edition", EAN13: "5901234123457"},
				{Support: models.SupportTypeVinyl, VersionName: "Red Vinyl", EAN13: "5901234123464"},
				{Support: models.SupportTypeCassette, VersionName: "Standard Cassette", EAN13: "5901234123471"},
			},
		},
		{
			ID: insertableObjectID(31), MBID: "a3b4c5d6-1111-2222-3333-444455556667",
			Title: "Now 4", ReleaseYear: 1999, Type: models.AlbumTypeLP,
			ArtistID: nil,
			Tracks: []models.TrackEntry{
				{TrackNumber: 1, SongID: songIDs["last_resort"]},
				{TrackNumber: 2, SongID: songIDs["bad_romance"]},
				{TrackNumber: 3, SongID: songIDs["shape_of_you"]},
				{TrackNumber: 4, SongID: songIDs["poker_face"]},
			},
			Releases: []models.AlbumRelease{
				{Support: models.SupportTypeCD, VersionName: "Standard Edition", EAN13: "6901234123458"},
			},
		},
		{
			ID: insertableObjectID(32), MBID: "a3b4c5d6-1111-2222-3333-444455556668",
			Title: "Divide", ReleaseYear: 2025, Type: models.AlbumTypeLP,
			ArtistID: pObjectID(artistIDs["ed_sheeran"]),
			ImageURL: pString("https://upload.wikimedia.org/wikipedia/en/4/45/Divide_cover.png"),
			Tracks: []models.TrackEntry{
				{TrackNumber: 1, SongID: songIDs["shape_of_you"]},
				{TrackNumber: 2, SongID: songIDs["perfect"]},
			},
			Releases: []models.AlbumRelease{
				{Support: models.SupportTypeCD, VersionName: "Standard Edition", EAN13: "7901234123459"},
				{Support: models.SupportTypeVinyl, VersionName: "Blue Vinyl", EAN13: "7901234123466"},
			},
		},
		{
			ID: insertableObjectID(33), MBID: "a3b4c5d6-1111-2222-3333-444455556669",
			Title: "The Fame Monster", ReleaseYear: 2026, Type: models.AlbumTypeEP,
			ArtistID: pObjectID(artistIDs["lady_gaga"]),
			Tracks: []models.TrackEntry{
				{TrackNumber: 1, SongID: songIDs["bad_romance"]},
				{TrackNumber: 2, SongID: songIDs["poker_face"]},
			},
			Releases: []models.AlbumRelease{
				{Support: models.SupportTypeCD, VersionName: "Deluxe Edition", EAN13: "8901234123450"},
			},
		},
		// Sim esses são mesmo os nomes dos albuns do Ed
		{
			ID: insertableObjectID(50), MBID: "gbdjh-0000000000011",
			Title: "+", ReleaseYear: 2011, Type: models.AlbumTypeLP,
			ArtistID: pObjectID(artistIDs["ed_sheeran"]),
			Tracks: []models.TrackEntry{
				{TrackNumber: 1, SongID: songIDs["the_a_team"]},
				{TrackNumber: 2, SongID: songIDs["lego_house"]},
				{TrackNumber: 3, SongID: songIDs["give_me_love"]},
			},
			Releases: []models.AlbumRelease{
				{Support: models.SupportTypeCD, VersionName: "Standard Edition", EAN13: "1012345678901"},
			},
		},
		{
			ID: insertableObjectID(51), MBID: "gbdjh-0000000000012",
			Title: "x", ReleaseYear: 2014, Type: models.AlbumTypeLP,
			ArtistID: pObjectID(artistIDs["ed_sheeran"]),
			ImageURL: pString("https://upload.wikimedia.org/wikipedia/en/a/ad/X_cover.png"),
			Tracks: []models.TrackEntry{
				{TrackNumber: 1, SongID: songIDs["photograph"]},
				{TrackNumber: 2, SongID: songIDs["thinking_out_loud"]},
			},
			Releases: []models.AlbumRelease{
				{Support: models.SupportTypeCD, VersionName: "Standard Edition", EAN13: "1012345678902"},
				{Support: models.SupportTypeVinyl, VersionName: "White Vinyl", EAN13: "1012345678903"},
			},
		},
		{
			ID: insertableObjectID(52), MBID: "gbdjh-0000000000017",
			Title: "÷", ReleaseYear: 2017, Type: models.AlbumTypeLP,
			ArtistID: pObjectID(artistIDs["ed_sheeran"]),
			Tracks: []models.TrackEntry{
				{TrackNumber: 1, SongID: songIDs["castle_on_hill"]},
				{TrackNumber: 2, SongID: songIDs["shape_of_you"]},
				{TrackNumber: 3, SongID: songIDs["galway_girl"]},
				{TrackNumber: 4, SongID: songIDs["happier"]},
			},
			Releases: []models.AlbumRelease{
				{Support: models.SupportTypeCD, VersionName: "Standard Edition", EAN13: "1012345678910"},
			},
		},
		{
			ID: insertableObjectID(53), MBID: "gbdjh-0000000000021",
			Title: "=", ReleaseYear: 2021, Type: models.AlbumTypeLP,
			ArtistID: pObjectID(artistIDs["ed_sheeran"]),
			Tracks: []models.TrackEntry{
				{TrackNumber: 1, SongID: songIDs["bad_habits"]},
				{TrackNumber: 2, SongID: songIDs["shivers"]},
				{TrackNumber: 3, SongID: songIDs["overpass_graffiti"]},
			},
			Releases: []models.AlbumRelease{
				{Support: models.SupportTypeCD, VersionName: "Standard Edition", EAN13: "1012345678921"},
				{Support: models.SupportTypeVinyl, VersionName: "Transparent Vinyl", EAN13: "1012345678922"},
			},
		},
		{
			ID: insertableObjectID(54), MBID: "gbdjh-0000000000023",
			Title: "Subtract", ReleaseYear: 2023, Type: models.AlbumTypeLP,
			ArtistID: pObjectID(artistIDs["ed_sheeran"]),
			Tracks: []models.TrackEntry{
				{TrackNumber: 1, SongID: songIDs["eyes_closed"]},
			},
			Releases: []models.AlbumRelease{
				{Support: models.SupportTypeCD, VersionName: "Standard Edition", EAN13: "1012345678931"},
			},
		},
		{
			ID: insertableObjectID(55), MBID: "gbdjh-0000000000025",
			Title: "Play", ReleaseYear: 2025, Type: models.AlbumTypeLP,
			ArtistID: pObjectID(artistIDs["ed_sheeran"]),
			Tracks: []models.TrackEntry{
				{TrackNumber: 1, SongID: songIDs["azizam"]},
			},
			Releases: []models.AlbumRelease{
				{Support: models.SupportTypeCD, VersionName: "Standard Edition", EAN13: "1012345678941"},
			},
		},
	}

	count := 0
	for _, a := range albums {
		if err := repo.Insert(ctx, &a); err != nil {
			log.Printf("Failed to insert album %s: %v", a.Title, err)
			continue
		}
		count++
		log.Printf("Inserted album: %s (%s)", a.Title, a.Type)
	}
	printSection("Albums", count, 10)

	albumIDs := make([]primitive.ObjectID, 0, len(albums))
	for _, a := range albums {
		albumIDs = append(albumIDs, a.ID)
	}
	return albumIDs
}

func seedVersionRequests(ctx context.Context, repo *repository.MongoVersionRequestRepository, userIDs map[string]primitive.ObjectID, albumIDs []primitive.ObjectID) {
	requests := []models.VersionRequest{
		{ID: insertableObjectID(300), UserID: userIDs["pedro"], AlbumID: albumIDs[0], EAN13: "9999999999999", Support: models.SupportTypeVinyl, VersionName: "Picture Disc", Status: models.RequestStatusAccepted, RequestedAt: time.Now().Add(-48 * time.Hour)},
		{ID: insertableObjectID(301), UserID: userIDs["maria"], AlbumID: albumIDs[2], EAN13: "8888888888888", Support: models.SupportTypeCassette, VersionName: "Limited Cassette", Status: models.RequestStatusRejected, RequestedAt: time.Now().Add(-24 * time.Hour)},
		{ID: insertableObjectID(302), UserID: userIDs["joao"], AlbumID: albumIDs[3], EAN13: "7777777777777", Support: models.SupportTypeVinyl, VersionName: "Colored Vinyl", Status: models.RequestStatusAccepted, RequestedAt: time.Now().Add(-72 * time.Hour)},
	}

	count := 0
	for _, r := range requests {
		if err := repo.Insert(ctx, &r); err != nil {
			log.Printf("Failed to insert version request for user=%s: %v", r.UserID.Hex(), err)
			continue
		}
		count++
		log.Printf("Inserted version request for user=%s: %s — %s", r.UserID.Hex()[:8], r.AlbumID.Hex()[:8], r.Status)
	}
	printSection("Version Requests", count, 3)
}

func seedCollections(ctx context.Context, repo *repository.MongoCollectionRepository, userIDs map[string]primitive.ObjectID, albumIDs []primitive.ObjectID) {
	items := []models.CollectionItem{
		{ID: insertableObjectID(100), UserID: userIDs["pedro"], AlbumID: albumIDs[0], EAN13: "5901234123457", Support: models.SupportTypeCD, VersionName: "Standard Edition", AddedAt: time.Now()},
		{ID: insertableObjectID(101), UserID: userIDs["pedro"], AlbumID: albumIDs[2], EAN13: "7901234123466", Support: models.SupportTypeVinyl, VersionName: "Blue Vinyl", AddedAt: time.Now()},
		{ID: insertableObjectID(102), UserID: userIDs["maria"], AlbumID: albumIDs[0], EAN13: "5901234123464", Support: models.SupportTypeVinyl, VersionName: "Red Vinyl", AddedAt: time.Now()},
		{ID: insertableObjectID(103), UserID: userIDs["joao"], AlbumID: albumIDs[3], EAN13: "8901234123450", Support: models.SupportTypeCD, VersionName: "Deluxe Edition", AddedAt: time.Now()},
	}

	count := 0
	for _, item := range items {
		if err := repo.Insert(ctx, &item); err != nil {
			log.Printf("Failed to insert collection item (user=%s, ean=%s): %v", item.UserID.Hex(), item.EAN13, err)
			continue
		}
		count++
		log.Printf("Inserted collection item for user=%s album=%s", item.UserID.Hex()[:8], item.AlbumID.Hex()[:8])
	}
	printSection("Collection Items", count, 4)
}

func seedNotifications(ctx context.Context, repo *repository.MongoNotificationRepository, userIDs map[string]primitive.ObjectID) {
	notes := []models.Notification{
		{ID: insertableObjectID(200), UserID: userIDs["pedro"], RequestID: insertableObjectID(300), AlbumTitle: "Infest", Response: "aceite", Read: false, CreatedAt: time.Now()},
		{ID: insertableObjectID(201), UserID: userIDs["maria"], RequestID: insertableObjectID(301), AlbumTitle: "Divide", Response: "recusado", Read: true, CreatedAt: time.Now()},
		{ID: insertableObjectID(202), UserID: userIDs["joao"], RequestID: insertableObjectID(302), AlbumTitle: "The Fame Monster", Response: "aceite", Read: false, CreatedAt: time.Now()},
	}

	count := 0
	for _, n := range notes {
		if err := repo.Insert(ctx, &n); err != nil {
			log.Printf("Failed to insert notification for user=%s: %v", n.UserID.Hex(), err)
			continue
		}
		count++
		log.Printf("Inserted notification for user=%s: %s — %s", n.UserID.Hex()[:8], n.AlbumTitle, n.Response)
	}
	printSection("Notifications", count, 3)
}

func seedCustomLists(ctx context.Context, listRepo *repository.MongoCustomListRepository, itemRepo *repository.MongoCustomListItemRepository, userIDs map[string]primitive.ObjectID, albumIDs []primitive.ObjectID) {
	lists := []models.CustomList{
		{ID: insertableObjectID(400), UserID: userIDs["pedro"], Name: "Favorites", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: insertableObjectID(401), UserID: userIDs["maria"], Name: "Morning Vibes", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: insertableObjectID(402), UserID: userIDs["maria"], Name: "Road Trip", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	listCount := 0
	for _, l := range lists {
		if err := listRepo.Insert(ctx, &l); err != nil {
			log.Printf("Failed to insert list %s: %v", l.Name, err)
			continue
		}
		listCount++
		log.Printf("Inserted list: %s (user=%s)", l.Name, l.UserID.Hex()[:8])
	}
	printSection("Custom Lists", listCount, 3)

	items := []models.CustomListItem{
		{ID: insertableObjectID(410), ListID: insertableObjectID(401), AlbumID: albumIDs[2], AddedAt: time.Now()},
		{ID: insertableObjectID(411), ListID: insertableObjectID(401), AlbumID: albumIDs[3], AddedAt: time.Now()},
		{ID: insertableObjectID(412), ListID: insertableObjectID(401), AlbumID: albumIDs[5], AddedAt: time.Now()},
	}

	itemCount := 0
	for _, item := range items {
		if err := itemRepo.Insert(ctx, &item); err != nil {
			log.Printf("Failed to insert list item: %v", err)
			continue
		}
		itemCount++
		log.Printf("Inserted list item: album=%s into list=Morning Vibes", item.AlbumID.Hex()[:8])
	}
	printSection("List Items", itemCount, 3)
}

func pObjectID(id primitive.ObjectID) *primitive.ObjectID {
	return &id
}

func pString(s string) *string {
	return &s
}

func resolveIDs(keys []string, idMap map[string]primitive.ObjectID) []primitive.ObjectID {
	result := make([]primitive.ObjectID, len(keys))
	for i, k := range keys {
		result[i] = idMap[k]
	}
	return result
}

func insertableObjectID(n int) primitive.ObjectID {
	var id [12]byte
	id[0] = byte(n >> 24)
	id[1] = byte(n >> 16)
	id[2] = byte(n >> 8)
	id[3] = byte(n)
	for i := 4; i < 12; i++ {
		id[i] = 0x01
	}
	return primitive.ObjectID(id)
}

func printSection(name string, inserted, total int) {
	fmt.Printf("  %-10s %d/%d inserted\n", name+":", inserted, total)
}
