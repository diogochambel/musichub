import { Component, inject, OnInit } from '@angular/core';
import { RouterLink } from '@angular/router';
import { HomeService } from '../../services/home.service';
import { ErrorPopupService } from '../../services/error-popup.service';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-home',
  imports: [RouterLink],
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.css'],
})
export class HomeComponent implements OnInit {
  private errorPopup = inject(ErrorPopupService);
  private homeService = inject(HomeService);
  authService = inject(AuthService);

  featuredAlbums: any[] = [];
  featuredArtists: any[] = [];
  recentReleases: any[] = [];

  ngOnInit() {
    this.homeService.getFeaturedAlbums().subscribe({
      next: (data) => (this.featuredAlbums = data),
      error: () => this.errorPopup.showError('Failed to load featured albums'),
    });
    this.homeService.getFeaturedArtists().subscribe({
      next: (data) => (this.featuredArtists = data),
      error: () => this.errorPopup.showError('Failed to load featured artists'),
    });
    this.homeService.getRecentReleases().subscribe({
      next: (data) => (this.recentReleases = data),
      error: () => this.errorPopup.showError('Failed to load recent releases'),
    });
  }

  getWelcomeName(): string {
    const name = this.authService.username();
    if (!name) return '';
    return name.charAt(0).toUpperCase() + name.slice(1);
  }
}
