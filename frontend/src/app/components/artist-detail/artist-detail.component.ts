import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { Subscription } from 'rxjs';
import { MusicService } from '../../services/music.service';
import { ErrorPopupService } from '../../services/error-popup.service';
import { UserService } from '../../services/user.service';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-artist-detail',
  imports: [RouterLink],
  templateUrl: './artist-detail.component.html',
  styleUrls: ['./artist-detail.component.css'],
})
export class ArtistDetailComponent implements OnInit, OnDestroy {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private musicService = inject(MusicService);
  private errorPopup = inject(ErrorPopupService);
  private userService = inject(UserService);
  private authService = inject(AuthService);
  private routeSub!: Subscription;

  artist: any = null;
  isFavorite = false;
  favoriteArtistId: string | null = null;
  animating = false;
  favoriteSuccessMessage = '';
  favoriteMessageType: 'added' | 'removed' | '' = '';

  get displayedAlbums() {
    if (!this.artist?.albums) return [];
    return this.artist.albums.slice(0, 4);
  }

  get hasMoreAlbums() {
    if (!this.artist?.albums) return false;
    return this.artist.albums.length > 4;
  }

  ngOnInit() {
    this.routeSub = this.route.paramMap.subscribe((params) => {
      const id = params.get('id');
      if (id) {
        this.artist = null;
        this.musicService.getArtist(id).subscribe({
          next: (data) => {
            this.artist = data;
            this.checkFavorite();
          },
          error: () => this.errorPopup.showError('Failed to load artist'),
        });
      }
    });
  }

  ngOnDestroy() {
    this.routeSub?.unsubscribe();
  }

  private checkFavorite() {
    const uid = this.authService.getCurrentUserId();
    if (!uid) {
      this.isFavorite = false;
      this.favoriteArtistId = null;
      return;
    }
    this.userService.getUser(uid).subscribe({
      next: (user) => {
        this.favoriteArtistId = user.favorite_artist_id || null;
        this.isFavorite = this.favoriteArtistId === this.artist?.id;
      },
      error: () => {
        this.favoriteArtistId = null;
        this.isFavorite = false;
      },
    });
  }

  typeLabel(type: string): string {
    if (type === 'solo') return 'Solo Artist';
    if (type === 'group') return 'Group';
    return type;
  }

  toggleFavorite() {
    if (!this.authService.getToken()) {
      this.errorPopup.showError('You must be logged in to favorite an artist');
      return;
    }

    const uid = this.authService.getCurrentUserId();
    const aid = this.artist.id;

    if (!uid || !aid) {
      this.errorPopup.showError('An error occurred');
      return;
    }

    this.animating = true;
    this.favoriteSuccessMessage = '';
    this.favoriteMessageType = '';

    if (this.isFavorite) {
      this.userService.removeFavoriteArtist(uid).subscribe({
        next: () => {
          this.isFavorite = false;
          this.favoriteArtistId = null;
          this.animating = false;
          this.favoriteSuccessMessage = 'Artist removed from favorites.';
          this.favoriteMessageType = 'removed';
        },
        error: () => {
          this.animating = false;
          this.errorPopup.showError('Failed to remove favorite');
        },
      });
    } else {
      if (this.favoriteArtistId && this.favoriteArtistId !== aid) {
        this.animating = false;
        this.errorPopup.showError(
          'You already have a different favorite artist. Remove it first before adding a new one.',
        );
        return;
      }

      this.userService.setFavoriteArtist(uid, aid).subscribe({
        next: () => {
          this.isFavorite = true;
          this.favoriteArtistId = aid;
          this.animating = false;
          this.favoriteSuccessMessage = 'Artist added to favorites!';
          this.favoriteMessageType = 'added';
        },
        error: () => {
          this.animating = false;
          this.errorPopup.showError('Failed to set favorite');
        },
      });
    }
  }

  onAnimationDone() {
    this.animating = false;
  }
}
