import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { Subscription } from 'rxjs';
import { MusicService } from '../../services/music.service';
import { ErrorPopupService } from '../../services/error-popup.service';

@Component({
  selector: 'app-artist-albums',
  imports: [RouterLink],
  templateUrl: './artist-albums.component.html',
  styleUrls: ['./artist-albums.component.css'],
})
export class ArtistAlbumsComponent implements OnInit, OnDestroy {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private musicService = inject(MusicService);
  private errorPopup = inject(ErrorPopupService);
  private routeSub!: Subscription;

  artist: any = null;
  albums: any[] = [];

  ngOnInit() {
    this.routeSub = this.route.paramMap.subscribe((params) => {
      const id = params.get('id');
      if (id) {
        this.musicService.getArtist(id).subscribe({
          next: (data) => {
            this.artist = data;
            this.albums = data.albums || [];
          },
          error: () => this.errorPopup.showError('Failed to load artist'),
        });
      }
    });
  }

  ngOnDestroy() {
    this.routeSub?.unsubscribe();
  }
}
