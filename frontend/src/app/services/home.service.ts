import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';

@Injectable({ providedIn: 'root' })
export class HomeService {
  private http = inject(HttpClient);
  private apiUrl = environment.apiUrl;

  getFeaturedAlbums() {
    return this.http.get<any[]>(`${this.apiUrl}/api/albums/featured`);
  }

  getFeaturedArtists() {
    return this.http.get<any[]>(`${this.apiUrl}/api/artists/featured`);
  }

  getRecentReleases() {
    return this.http.get<any[]>(`${this.apiUrl}/api/albums/recent`);
  }
}
