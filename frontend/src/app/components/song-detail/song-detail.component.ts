import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { Subscription } from 'rxjs';
import { MusicService } from '../../services/music.service';
import { ErrorPopupService } from '../../services/error-popup.service';

@Component({
  selector: 'app-song-detail',
  imports: [RouterLink],
  templateUrl: './song-detail.component.html',
  styleUrls: ['./song-detail.component.css'],
})
export class SongDetailComponent implements OnInit, OnDestroy {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private musicService = inject(MusicService);
  private errorPopup = inject(ErrorPopupService);
  private routeSub!: Subscription;

  song: any = null;

  ngOnInit() {
    this.routeSub = this.route.paramMap.subscribe((params) => {
      const id = params.get('id');
      if (id) {
        this.song = null;
        this.musicService.getSong(id).subscribe({
          next: (data) => (this.song = data),
          error: () => this.errorPopup.showError('Failed to load song'),
        });
      }
    });
  }

  ngOnDestroy() {
    this.routeSub?.unsubscribe();
  }

  formatDuration(seconds: number): string {
    const m = Math.floor(seconds / 60);
    const s = seconds % 60;
    return `${m}:${s.toString().padStart(2, '0')}`;
  }
}
