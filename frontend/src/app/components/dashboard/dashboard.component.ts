import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { AuthService } from '../../services/auth.service';
import { MusicService } from '../../services/music.service';
import { ErrorPopupService } from '../../services/error-popup.service';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule],
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.css']
})


export class DashboardComponent {
  authService = inject(AuthService);
  musicService = inject(MusicService);
  errorPopupService = inject(ErrorPopupService);
  router = inject(Router);

  searchQuery = '';
  searchResults: any[] | null = null;
  searchPerformed = false;
  searchMode: 'artists' | 'albums' = 'artists';

  getWelcomeName(): string {
    const name = this.authService.username();
    if (!name) return '';
    return name.charAt(0).toUpperCase() + name.slice(1);
  }

  onSearch() {
    const query = this.searchQuery.trim();
    if (!query) {
      this.searchResults = null;
      this.searchPerformed = false;
      return;
    }

    const search$ =
      this.searchMode === 'albums'
        ? this.musicService.searchAlbums(query)
        : this.musicService.searchArtists(query);

    search$.subscribe({
      next: (data) => {
        this.searchResults = data;
        this.searchPerformed = true;
      },
      error: () => {
        this.errorPopupService.showError(`Failed to search ${this.searchMode}`);
        this.searchResults = [];
        this.searchPerformed = true;
      },
    });
  }

  clearSearch() {
    this.searchQuery = '';
    this.searchResults = null;
    this.searchPerformed = false;
  }

  onSearchModeChange() {
    this.searchResults = null;
    this.searchPerformed = false;
  }

  artistTypeLabel(type: string): string {
    if (type === 'solo') return 'Solo Artist';
    if (type === 'group') return 'Group';
    return type;
  }

}