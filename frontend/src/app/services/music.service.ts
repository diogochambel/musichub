import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';

@Injectable({ providedIn: 'root' })
export class MusicService {
  private http = inject(HttpClient);
  private apiUrl = environment.apiUrl;

  getArtist(id: string) {
    return this.http.get<any>(`${this.apiUrl}/api/artists/${id}`);
  }

  getAlbum(id: string) {
    return this.http.get<any>(`${this.apiUrl}/api/albums/${id}`);
  }

  getSong(id: string) {
    return this.http.get<any>(`${this.apiUrl}/api/songs/${id}`);
  }

  searchArtists(query: string) {
    return this.http.get<any[]>(`${this.apiUrl}/api/artists/search`, {
      params: { q: query },
    });
  }

  searchAlbums(query: string) {
    return this.http.get<any[]>(`${this.apiUrl}/api/albums/search`, {
      params: { q: query },
    });
  }

  getAlbums() {
    return this.http.get<any[]>(`${this.apiUrl}/api/albums`);
  }
}
